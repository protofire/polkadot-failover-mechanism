package aws

/*
Set AWS_ACCESS_KEY, AWS_SECRET_KEY, PREFIX before running these scripts

Additional envs:
	POLKADOT_TEST_NO_POST_TF_CLEANUP    - no terraform destroy command after tests
	POLKADOT_TEST_INITIAL_TF_CLEANUP    - terraform destroy command before test
	POLKADOT_TEST_NO_INITIAL_TF_APPLY   - no terraform apply command before test
*/

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	helpers2 "github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	taws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/aws"
	"github.com/protofire/polkadot-failover-mechanism/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gather environmental variables and set reasonable defaults
var (
	awsRegions    = []string{"us-east-1", "eu-central-1", "us-west-1"}
	awsAccessKeys = []string{os.Getenv("AWS_ACCESS_KEY")}
	awsSecretKeys = []string{os.Getenv("AWS_SECRET_KEY")}
	sshUser       = "ec2-user"
	terraformDir  = "../../aws/"
)

// A collection of tests that will be run
func TestBundle(t *testing.T) {

	require.NotEmpty(t, awsAccessKeys[0], "AWS_ACCESS_KEY env required")
	require.NotEmpty(t, awsSecretKeys[0], "AWS_SECRET_KEY env required")

	// Set backend variables
	var s3bucket, s3key, s3region, prefix string
	var ok bool

	if prefix, ok = os.LookupEnv("PREFIX"); !ok {
		prefix = helpers2.RandStringBytes(4)
	}

	if s3bucket, ok = os.LookupEnv("TF_STATE_BUCKET"); !ok {
		s3bucket = fmt.Sprintf("%s-polkadot-validator-failover-tfstate", prefix)
	}

	if s3key, ok = os.LookupEnv("TF_STATE_KEY"); !ok {
		s3key = "terraform.tfstate"
	}

	if s3region, ok = os.LookupEnv("TF_STATE_REGION"); !ok {
		s3region = "us-east-1"
	}

	bucketCreated, err := aws.EnsureTFBucket(s3bucket, s3region)
	require.NoError(t, err)
	t.Logf("TF state bucket %q has been ensured", s3bucket)

	require.NoError(t, helpers.ClearLocalTFState(terraformDir))

	// Generate new SSH key for test virtual machines
	sshKey := helpers.GenerateSSHKeys(t)

	// Configure Terraform - set backend, minimum set of infrastructure variables. Also expose ssh
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: terraformDir,

		BackendConfig: map[string]interface{}{
			"bucket": s3bucket,
			"region": s3region,
			"key":    prefix + "-" + s3key,
		},

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"aws_access_keys":       awsAccessKeys,
			"aws_secret_keys":       awsSecretKeys,
			"aws_regions":           helpers.BuildRegionParams(awsRegions...),
			"validator_keys":        "{key1={key=\"0x6ce96ae5c300096b09dbd4567b0574f6a1281ae0e5cfe4f6b0233d1821f6206b\",type=\"gran\",seed=\"favorite liar zebra assume hurt cage any damp inherit rescue delay panic\"},key2={key=\"0x3ff0766f9ebbbceee6c2f40d9323164d07e70c70994c9d00a9512be6680c2394\",type=\"aura\",seed=\"expire stage crawl shell boss any story swamp skull yellow bamboo copy\"}}",
			"key_name":              "test",
			"key_content":           sshKey.PublicKey,
			"prefix":                prefix,
			"delete_on_termination": true,
			"cpu_limit":             "1",
			"ram_limit":             "1",
			"validator_name":        "test",
			"expose_ssh":            true,
			"expose_prometheus":     true,
			"node_key":              "fc9c7cf9b4523759b0a43b15ff07064e70b9a2d39ef16c8f62391794469a1c5e",
			"chain":                 "westend",
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	helpers.SetPostTFCleanUp(t, func() {
		if _, ok := os.LookupEnv("POLKADOT_TEST_NO_POST_TF_CLEANUP"); !ok {
			terraform.Destroy(t, terraformOptions)
			if bucketCreated {
				require.NoError(t, aws.DeleteTFBucket(s3bucket, s3region))
			} else {
				require.NoError(t, aws.ClearTFBucket(s3bucket, s3region))
			}
		} else {
			t.Log("Skipping terrafrom deferred cleanup...")
		}
	})

	// Run `terraform init`
	terraform.Init(t, terraformOptions)

	terraform.RunTerraformCommand(t, terraformOptions, terraform.FormatArgs(terraformOptions, "validate")...)

	helpers.SetInitialTFCleanUp(t, terraformOptions)

	// Run `terraform apply` and fail the test if there are any errors
	terraform.Apply(t, terraformOptions)

	require.True(t, t.Run("DistributedMode", func(t *testing.T) {
		// TEST 1: Verify that there are healthy instances in each region with public ips assigned
		var instanceIDs []string
		var publicIPs []string
		for _, region := range awsRegions {
			// GetHealthyEc2InstanceIdsByTag located in ec2.go file
			regionInstances := aws.GetHealthyEc2InstanceIdsByTag(t, region, "prefix", prefix)
			require.GreaterOrEqualf(t, len(regionInstances), 1, "ERROR! No instances found in %s region.", region)
			t.Logf("INFO. The following instances found in %s region: %s.", region, strings.Join(regionInstances, ","))

			instanceIDs = append(instanceIDs, regionInstances...)
			// Fetching PublicIPs for the instances we have found
			regionIPs := taws.GetPublicIpsOfEc2Instances(t, regionInstances, region)

			require.GreaterOrEqualf(t, len(regionIPs), 1, "ERROR! No public IPs found for instances in %s region.", region)
			for k, v := range regionIPs {
				publicIPs = append(publicIPs, v)
				t.Logf("InstanceID: %s, InstanceIP: %s", k, v)
			}
		}

		t.Logf("INFO. Instances IDs found in all regions: %s", strings.Join(instanceIDs, ","))

		// TEST 2: Verify the number of existing EC2 instances - should be an odd number
		t.Run("Instance count", func(t *testing.T) {

			instanceCount := len(instanceIDs)

			require.Equal(t, instanceCount%2, 1, "INFO. There are odd instances running")
			t.Log("INFO. There are odd instances running")

			// TEST 3: Verify the number of existing EC2 instances - should be at least 3
			require.Greaterf(t, instanceCount, 2, "ERROR! Minimum viable instance count (3) not reached. There are %d instances running.", instanceCount)
			t.Logf("INFO. Minimum viable instance count (3) reached. There are %d instances running.", instanceCount)
		})

		// TEST 4: Verify the number of Consul locks each instance is aware about. Should be exactly 1 lock on each instnace
		t.Run("Consul verifications", func(t *testing.T) {

			if assert.True(t, helpers.ConsulLockCheck(t, publicIPs, sshKey, sshUser)) {
				t.Log("INFO. Consul lock check passed. Each Consul node can see exactly 1 lock.")
			}

			// TEST 5: All of the Consul nodes should be healthy
			if assert.True(t, helpers.ConsulCheck(t, publicIPs, sshKey, sshUser)) {
				t.Log("INFO. Consul check passed. Each node can see full cluster, all nodes are healthy")
			}

		})

		t.Run("Polkadot verifications", func(t *testing.T) {

			// TEST 6: Verify that there is only one Polkadot node working in Validator mode at a time
			if assert.True(t, helpers.LeadersCheck(t, publicIPs, sshKey, sshUser)) {
				t.Log("INFO. Leaders check passed. Exactly 1 leader found")
			}

			// TEST 7: Verify that all Polkadot nodes are health
			if assert.True(t, helpers.PolkadotCheck(t, publicIPs, sshKey, sshUser)) {
				t.Log("INFO. Polkadot node check passed. All instances are healthy")
			}

		})

		// TEST 8: All the validator keys were successfully uploaded to SSM in each region
		t.Run("SSM tests", func(t *testing.T) {

			if assert.True(t, aws.SSMCheck(t, awsRegions, prefix), awsRegions, prefix) {
				t.Log("INFO. All keys were uploaded. Private key is encrypted.")
			}
		})

		// TEST 9: Verify that all the groups that are used by the nodes are valid and contains verified rules only.
		t.Run("Security groups tests", func(t *testing.T) {

			if assert.True(t, aws.SGCheck(t, awsRegions, prefix)) {
				t.Log("INFO. Security groups contains only an appropriate set of rules.")
			}
		})

		// TEST 10: Check that there are no unassigned volumes after the nodes started
		t.Run("Volumes tests", func(t *testing.T) {
			if assert.Truef(t, aws.VolumesCheck(t, awsRegions, prefix), "WARNING! An unattached disk was detected with prefix %s", prefix) {
				t.Log("INFO. No disks left unattached.")
			}
		})

		// TEST 11: Check that no CloudWatch alarm were triggered
		t.Run("CloudWatch tests", func(t *testing.T) {
			expectedAlertsPerRegion := map[string]int{
				awsRegions[0]: 5,
				awsRegions[1]: 3,
				awsRegions[2]: 3,
			}
			if assert.True(t, aws.CloudWatchCheck(t, awsRegions, prefix, expectedAlertsPerRegion), "ERROR! Cloud Watch alarms are not in a good state") {
				t.Log("INFO. All Cloud Watch alarms were created. No Cloud Watch alarm were triggered.")
			}
		})

		// TEST 12: Check that ELB and each target group confirms that all the instances are healthy
		t.Run("NLB tests", func(t *testing.T) {
			if assert.True(t, aws.NLBCheck(t, terraform.OutputList(t, terraformOptions, "lbs"), awsRegions)) {
				t.Log("INFO. NLB is configured. All target groups do exists. Health checks responds that instance state is OK.")
			}
		})
		// TEST 13: Check that there are exactly 5 keys in the keystore
		t.Run("Keystore tests", func(t *testing.T) {
			if assert.True(t, helpers.KeystoreCheck(t, publicIPs, sshKey, sshUser)) {
				t.Log("INFO. There are exactly 5 keys in the Keystore")
			}
		})
	}))

	log.Printf("[DEBUG] failover: Getting validator in distributed mode....")
	validatorBefore, err := aws.WaitForValidatorRegions(awsRegions, prefix, "validator_value", prefix, 1200, 5)
	require.NoError(t, err)
	require.NotEmpty(t, validatorBefore.InstanceID)

	terraformOptions.Vars["failover_mode"] = "single"
	terraform.Apply(t, terraformOptions)

	t.Run("SingleMode", func(t *testing.T) {
		t.Run("CheckValidator", func(t *testing.T) {
			log.Printf("[DEBUG] failover: Getting validator in single mode....")
			validatorAfter, err := aws.WaitForValidatorRegions(awsRegions, prefix, "validator_value", prefix, 1200, 5)
			require.NoError(t, err)
			require.NotEmpty(t, validatorAfter.InstanceID)
			require.Equal(t, validatorBefore.InstanceID, validatorAfter.InstanceID)
		})

		t.Run("CheckVirtualMachines", func(t *testing.T) {
			count, err := aws.CheckVmsCount(awsRegions, prefix)
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

	})
}
