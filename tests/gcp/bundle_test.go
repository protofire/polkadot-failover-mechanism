package gcp

/*
Set PREFIX, GCP_PROJECT, and GOOGLE_APPLICATION_CREDENTIALS credentials before running these scripts

Additional environments:
	POLKADOT_TEST_NO_POST_TF_CLEANUP    - no terraform destroy command after tests
	POLKADOT_TEST_INITIAL_TF_CLEANUP    - terraform destroy command before test
	POLKADOT_TEST_NO_INITIAL_TF_APPLY   - no terraform apply command before test
	POLKADOT_TEST_CLEANUP               - clean gcp infrastructure finding all resources with test prefix, it uses GCP API requests
	POLKADOT_TEST_EXIT_AFTER_CLEANUP    - exit after intension cleanup
	DRY_RUN                             - dry run force cleanup

IAM Rules for tests:

* Editor
* Role Editor
* Secret Manager Editor
* Project IAM Editor
* Monitoring Editor

POLKADOT_TEST_NO_POST_TF_CLEANUP=yes POLKADOT_TEST_INITIAL_TF_CLEANUP=yes make gcp

* Starts clean up without terrafrom apply, only GCP API
POLKADOT_TEST_CLEANUP=yes POLKADOT_TEST_EXIT_AFTER_CLEANUP=yes DRY_RUN=yes make gcp

*/

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/protofire/polkadot-failover-mechanism/tests/gcp/utils"
	"github.com/protofire/polkadot-failover-mechanism/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//Gather environmental variables and set reasonable defaults
var (
	gcpRegion     = []string{"us-east1", "us-east4", "us-west1"}
	gcpProject    = os.Getenv("GCP_PROJECT")
	forceCleanup  = len(os.Getenv("POLKADOT_TEST_CLEANUP")) > 0
	exitOnCleanup = len(os.Getenv("POLKADOT_TEST_EXIT_AFTER_CLEANUP")) > 0
	noApply       = len(os.Getenv("POLKADOT_TEST_NO_INITIAL_TF_APPLY")) > 0
	dryRun        = len(os.Getenv("DRY_RUN")) > 0
	sshUser       = "ubuntu"
)

func TestBundle(t *testing.T) {

	require.NotEmpty(t, gcpProject, "GCP_PROJECT env required")
	require.NotEmpty(t, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "GOOGLE_APPLICATION_CREDENTIALS env required")

	var (
		prefix    string
		gcpBucket string
		ok        bool
	)

	if prefix, ok = os.LookupEnv("PREFIX"); !ok {
		prefix = helpers.RandStringBytes(4)
	}

	if gcpBucket, ok = os.LookupEnv("TF_STATE_BUCKET"); !ok {
		gcpBucket = fmt.Sprintf("%s-polkadot-validator-failover-tfstate", prefix)
	}

	bucketCreated, err := utils.EnsureTFBucket(gcpProject, gcpBucket)
	require.NoError(t, err)
	t.Logf("TF state bucket %q has been ensured", gcpBucket)

	if forceCleanup {
		err = utils.CleanResources(gcpProject, prefix, dryRun)
		require.NoError(t, err)
		err = utils.ClearTFBucket(gcpProject, gcpBucket)
		require.NoError(t, err)
		if exitOnCleanup {
			return
		}
	}

	// Generate new SSH key for test virtual machines
	sshKey := ssh.GenerateRSAKeyPair(t, 4096)

	// Configure Terraform - set backend, minimum set of infrastructure variables. Also expose ssh
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../gcp/",

		BackendConfig: map[string]interface{}{
			"bucket": gcpBucket,
			"prefix": prefix,
		},

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"gcp_regions":           helpers.BuildRegionParams(gcpRegion...),
			"gcp_project":           gcpProject,
			"validator_keys":        "{key1={key=\"0x6ce96ae5c300096b09dbd4567b0574f6a1281ae0e5cfe4f6b0233d1821f6206b\",type=\"gran\",seed=\"favorite liar zebra assume hurt cage any damp inherit rescue delay panic\"},key2={key=\"0x3ff0766f9ebbbceee6c2f40d9323164d07e70c70994c9d00a9512be6680c2394\",type=\"aura\",seed=\"expire stage crawl shell boss any story swamp skull yellow bamboo copy\"}}",
			"gcp_ssh_user":          sshUser,
			"gcp_ssh_pub_key":       sshKey.PublicKey,
			"prefix":                prefix,
			"delete_on_termination": "true",
			"cpu_limit":             "1",
			"ram_limit":             "1",
			"validator_name":        "test",
			"expose_ssh":            "true",
			"node_key":              "fc9c7cf9b4523759b0a43b15ff07064e70b9a2d39ef16c8f62391794469a1c5e",
			"chain":                 "westend",
			"admin_email":           "1627_DEV@altoros.com",
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created

	helpers.SetPostTFCleanUp(t, func() {
		if _, ok := os.LookupEnv("POLKADOT_TEST_NO_POST_TF_CLEANUP"); !ok {
			terraform.Destroy(t, terraformOptions)
			if bucketCreated {
				require.NoError(t, utils.DeleteTFBucket(gcpProject, gcpBucket))
			} else {
				require.NoError(t, utils.ClearTFBucket(gcpProject, gcpBucket))
			}
		} else {
			t.Log("Skipping terrafrom deferred cleanup...")
		}
	})

	if !noApply {
		// Run `terraform init`
		terraform.Init(t, terraformOptions)

		terraform.RunTerraformCommand(t, terraformOptions, terraform.FormatArgs(terraformOptions, "validate")...)

		helpers.SetInitialTFCleanUp(t, terraformOptions)

		// Run `terraform apply` and fail the test if there are any errors
		terraform.Apply(t, terraformOptions)
	}
	// TEST 1: Verify that there are healthy instances in each region with public ips assigned
	var instanceIPs []string

	t.Run("Instances", func(t *testing.T) {
		for _, value := range gcpRegion {
			regionInstanceIPs := gcp.FetchRegionalInstanceGroup(t, gcpProject, value, fmt.Sprintf("%s-instance-group-manager", prefix)).GetPublicIps(t, gcpProject)

			require.GreaterOrEqualf(t, len(regionInstanceIPs), 1, "ERROR! No instances found in %s region.", value)
			t.Logf("INFO. The following instances found in %s region: %s.", value, strings.Join(regionInstanceIPs, ","))

			// Fetching PublicIPs for the instances we have found
			instanceIPs = append(instanceIPs, regionInstanceIPs...)
			t.Logf("INFO. Instances IPs found in all regions: %s", strings.Join(instanceIPs, ","))
		}
	})

	// TEST 2: Verify the number of existing GCP instances - should be an odd number
	t.Run("InstanceCount", func(t *testing.T) {

		instanceCount := len(instanceIPs)

		require.Equal(t, instanceCount%2, 1, "ERROR! There are even instances running")
		t.Log("INFO. There are odd instances running")

		// TEST 3: Verify the number of existing EC2 instances - should be at least 3
		require.Greaterf(t, instanceCount, 2, "ERROR! Minimum viable instance count (3) not reached. There are %d instances running.", instanceCount)
		t.Logf("INFO. Minimum viable instance count (3) reached. There are %d instances running.", instanceCount)
	})

	// TEST 4: Verify the number of Consul locks each instance is aware about. Should be exactly 1 lock on each instnance
	t.Run("ConsulVerifications", func(t *testing.T) {

		if assert.True(t, helpers.ConsulLockCheck(t, instanceIPs, sshKey, sshUser)) {
			t.Log("INFO. Consul lock check passed. Each Consul node can see exactly 1 lock.")
		}

		// TEST 5: All of the Consul nodes should be healthy
		if assert.True(t, helpers.ConsulCheck(t, instanceIPs, sshKey, sshUser)) {
			t.Log("INFO. Consul check passed. Each node can see full cluster, all nodes are healthy")
		}

	})

	t.Run("PolkadotVerifications", func(t *testing.T) {

		// TEST 6: Verify that there is only one Polkadot node working in Validator mode at a time
		if assert.True(t, helpers.LeadersCheck(t, instanceIPs, sshKey, sshUser)) {
			t.Log("INFO. Leaders check passed. Exactly 1 leader found")
		}
		// TEST 7: Verify that all Polkadot nodes are healthy
		if assert.True(t, helpers.PolkadotCheck(t, instanceIPs, sshKey, sshUser)) {
			t.Log("INFO. Polkadot node check passed. All instances are healthy")
		}

	})

	// TEST 8: All the validator keys were successfully uploaded
	t.Run("SMTests", func(t *testing.T) {
		if assert.True(t, utils.SMCheck(t, prefix, gcpProject)) {
			t.Log("INFO. All keys were uploaded. Private key is encrypted.")
		}
	})

	// TEST 9: All the firewalls were successfully created
	t.Run("FirewallTests", func(t *testing.T) {
		if assert.NoError(t, utils.FirewallCheck(prefix, gcpProject)) {
			t.Log("INFO. All firewalls were successfully created")
		}
	})

	// TEST 10: Check that all disks are being mounted
	t.Run("VolumesTests", func(t *testing.T) {
		if assert.NoError(t, utils.VolumesCheck(prefix, gcpProject)) {
			t.Log("INFO. All volumes were successfully created and attached")
		}
	})

	// TEST 11: Check that all alert policies have been created
	t.Run("AlertsTests", func(t *testing.T) {
		if assert.NoError(t, utils.AlertsPoliciesCheck(prefix, gcpProject)) {
			t.Log("INFO. All alerts policies were successfully created")
		}
	})

	// TEST 12: Instances health check
	t.Run("HealthCheckTests", func(t *testing.T) {
		if assert.NoError(t, utils.HealthStatusCheck(prefix, gcpProject)) {
			t.Log("INFO. There are all healthy instances")
		}
	})

	// TEST 13: Check that there are exactly 5 keys in the keystore
	t.Run("KeystoreTests", func(t *testing.T) {
		if assert.True(t, helpers.KeystoreCheck(t, instanceIPs, sshKey, sshUser)) {
			t.Log("INFO. There are exactly 5 keys in the Keystore")
		}
	})

}
