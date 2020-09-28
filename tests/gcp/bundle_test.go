package test

// Set PREFIX, GCP_PROJECT, and GOOGLE_APPLICATION_CREDENTIALS credentials before running these scripts

import (
	"testing"
	"os"
	"time"
	"strconv"
	"fmt"
	"encoding/json"
    "strings"

	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/terraform"
    "github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/retry"

	"github.com/google/go-cmp/cmp"

	"google.golang.org/api/compute/v1"
	"github.com/stretchr/testify/assert"
)

//Gather environmental variables and set reasonable defaults
var gcpRegion = [3]string{"us-east1", "us-east4", "us-west1"}
var prefix = os.Getenv("PREFIX")

func TestBundle(t *testing.T) {

    // Generate new SSH key for test virtual machines
	sshKey := ssh.GenerateRSAKeyPair(t, 4096)

	// Configure Terraform - set backend, minimum set of infrastructure variables. Also expose ssh 
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../gcp/",

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"gcp_regions": "[\"" + awsRegion[0] + "\", \"" + awsRegion[1] + "\", \"" + awsRegion[2] + "\"]",
			"validator_keys": "{key1={key=\"0x6ce96ae5c300096b09dbd4567b0574f6a1281ae0e5cfe4f6b0233d1821f6206b\",type=\"gran\",seed=\"favorite liar zebra assume hurt cage any damp inherit rescue delay panic\"},key2={key=\"0x3ff0766f9ebbbceee6c2f40d9323164d07e70c70994c9d00a9512be6680c2394\",type=\"aura\",seed=\"expire stage crawl shell boss any story swamp skull yellow bamboo copy\"}}",
			"gcp_ssh_user": "ubuntu",
			"gcp_ssh_pub_key": sshKey.PublicKey,
            "prefix": prefix,
			"delete_on_termination": "true",
			"cpu_limit": "1",
			"ram_limit": "1",
			"validator_name": "test",
			"expose_ssh": "true",
			"node_key": "fc9c7cf9b4523759b0a43b15ff07064e70b9a2d39ef16c8f62391794469a1c5e",
            "chain": "westend",
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

    // TEST 1: Verify that there are healthy instances in each region with public ips assigned
	var instanceIPs []string

	for _, value := range awsRegion {
        regionInstances := gcp.FetchRegionalInstanceGroup(t, value, os.Getenv("GCP_PROJECT"),fmt.Println(os.Getenv("PREFIX"),"-instance-group-manager")).GetPublicIps(t, os.Getenv("GCP_PROJECT"))
        
        
		if len(regionInstances) < 1 {
			t.Error("ERROR! No instances found in " + value + " region.")
		} else {
			t.Log("INFO. The following instances found in " + value + " region: " + strings.Join(regionInstances,",") + ".")
		}

		instanceIPs = append(instanceIPs, regionInstances...)
        // Fetching PublicIPs for the instances we have found
	}

	t.Log("INFO. Instances IPs found in all regions: " + strings.Join(instanceIPs,","))
	
      var test bool = false
    // TEST 2: Veriy the number of existing EC2 instances - should be an odd number
	t.Run("Instance count", func(t *testing.T) {

		instance_count := len(instanceIPs)

		test = assert.Equal(t, instance_count % 2, 1)
		if test {
			t.Log("INFO. There are odd instances running")
		} else {
			t.Error("ERROR! There are even instances running")
		}

    // TEST 3: Verify the number of existing EC2 instances - should be at least 3
		test = assert.True(t, instance_count > 2)
		if test {
			t.Log("INFO. Minimum viable instance count (3) reached. There are " + string(instance_count) + " instances running.")
		} else {
			t.Error("ERROR! Minimum viable instance count (3) not reached. There are " + string(instance_count) + " instances running.")
		}
	})
    
    // TEST 4: Verify the number of Consul locks each instance is aware about. Should be exactly 1 lock on each instnace
	t.Run("Consul verifications", func(t *testing.T) {

	        test = assert.True(t, ConsulLockCheck(t, publicIPs, sshKey))
	        if test {
		        t.Log("INFO. Consul lock check passed. Each Consul node can see exactly 1 lock.")
		}

    // TEST 5: All of the Consul nodes should be healthy
		test = assert.True(t, ConsulCheck(t, publicIPs, sshKey))
	        if test {
		        t.Log("INFO. Consul check passed. Each node can see full cluster, all nodes are healthy")
		}


	})

	t.Run("Polkadot verifications", func(t *testing.T) {

    // TEST 6: Verify that there is only one Polkadot node working in Validator mode at a time
		test = assert.True(t, LeadersCheck(t, publicIPs, sshKey))
		if test {
			t.Log("INFO. Leaders check passed. Exactly 1 leader found")
		}
        
    // TEST 7: Verify that all Polkadot nodes are health
		test = assert.True(t, PolkadotCheck(t, publicIPs, sshKey))
		if test {
			t.Log("INFO. Polkadot node check passed. All instances are healthy")
		}

	})
}

// TEST 4
func ConsulLockCheck(t *testing.T, publicIPs map[string]string, key *ssh.KeyPair) bool {

  command := "consul kv export | grep \"prefix/.lock\" | wc -l"
  array := NodeQuery(t, publicIPs, key, command)

  if len(array) == 0 {
	  return false
  } else {

    for _, value := range array {

      intValue, err := strconv.Atoi(value)

      if err != nil {
	      t.Error("ERROR! " + err.Error())
	      return false
      }

      if intValue != 1 {
	      t.Error("ERROR! Error while retrieving Consul lock. Got: " + string(intValue) + " locks, should be exactly 1 lock.")
	      return false
      }
    }
  }

  return true

}

// TEST 5
func ConsulCheck(t *testing.T, publicIPs map[string]string, key *ssh.KeyPair) bool {

  command := "consul members --status alive | wc -l"
  array := NodeQuery(t, publicIPs, key, command)

  if len(array) == 0 {
	  return false
  } else {

    for _, value := range array {

      intValue, err := strconv.Atoi(value)

      if err != nil {
	      t.Error("ERROR! " + err.Error())
	      return false
      }

      var instanceCountExpected int = len(publicIPs) + 1

      if intValue != instanceCountExpected {
	      t.Error("ERROR! Consul node count not matched. One of the nodes responded the following healthy instance count: " + string(intValue) + ", while there should be " + string(instanceCountExpected) + " instances")
	      return false
      }
    }
  }

  return true

}

// TEST 6
func LeadersCheck(t *testing.T, publicIPs map[string]string, key *ssh.KeyPair) bool {

  command := "curl -s -H \"Content-Type: application/json\" -d '{\"id\":1, \"jsonrpc\":\"2.0\", \"method\": \"system_nodeRoles\", \"params\":[]}' http://localhost:9933"
  array := NodeQuery(t, publicIPs, key, command)

  t.Log("Leaders output: " + strings.Join(array, ","))

  var leaders, nodes int = 0, 0

  if len(array) == 0 {
	  return false
  } else {

    // SSH into the node and ensure, that only one node returns "Authority" for system_nodeRoles call
    for _, value := range array {

      if value == "{\"jsonrpc\":\"2.0\",\"result\":[\"Authority\"],\"id\":1}" {
	      leaders++
      } else if value == "{\"jsonrpc\":\"2.0\",\"result\":[\"Full\"],\"id\":1}" {
	      nodes++
      } else {
	      t.Error("ERROR! Node working not in Full, not in Authority mode.")
	      return false
      }
    }
  }
    
  if leaders == 1 && nodes == len(publicIPs) - leaders {
	  t.Log("INFO. There are exactly one leader and the rest nodes are all working in a Full mode")
	  return true
  } else if leaders > 1 {
	  t.Error("ERROR! There are more than 1 leader at the same time.")
	  return false
  } else if leaders < 1 {
	  t.Error("ERROR! There are no leaders.")
	  return false
  } else {
	  t.Error("ERROR! Some of the full nodes are not working correctly.")
	  return false
  }
}

// TEST 7
func PolkadotCheck(t *testing.T, publicIPs map[string]string, key *ssh.KeyPair) bool {

  command := "curl -s -H \"Content-Type: application/json\" -d '{\"id\":1, \"jsonrpc\":\"2.0\", \"method\": \"system_health\", \"params\":[]}' http://localhost:9933"
  array := NodeQuery(t, publicIPs, key, command)

  t.Log("Leaders output: " + strings.Join(array, ","))

  if len(array) == 0 {
	  return false
  } else {

// Parse JSON and verify that node has healthy state
    for _,v := range array {

      type resultHealth struct {
	      peers int
	      shouldHavePeers bool
	      isSyncing bool
      }

      type Health struct {
	      jsonrpc string
	      result resultHealth
	      id int
      }

      var result Health
      err := json.Unmarshal([]byte(v), &result)

      if err != nil {
          t.Error("ERROR! " + err.Error())
	  return false
      }

      if result.result.peers < 2 && result.result.shouldHavePeers {
	  t.Error("ERROR! Node does not have enough peers")
	  return false
      }
    }
  }

  return true

}