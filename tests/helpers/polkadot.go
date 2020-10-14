package helpers

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/ssh"
)

// LeadersCheck check leader
func LeadersCheck(t *testing.T, publicIPs []string, key *ssh.KeyPair, user string) bool {

	command := "curl -s -H \"Content-Type: application/json\" -d '{\"id\":1, \"jsonrpc\":\"2.0\", \"method\": \"system_nodeRoles\", \"params\":[]}' http://localhost:9933"
	array := NodeQuery(t, publicIPs, key, command, user)

	t.Log("Leaders output: " + strings.Join(array, ","))

	var leaders, nodes int = 0, 0

	if len(array) == 0 {
		return false
	}
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

	if leaders == 1 && nodes == len(publicIPs)-leaders {
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

// KeystoreCheck checks keystore
func KeystoreCheck(t *testing.T, publicIPs []string, key *ssh.KeyPair, user string) bool {

	command := "ls -lah /data/chains/westend2/keystore | wc -l"

	for i := 0; i < 5; i++ {
		array := NodeQuery(t, publicIPs, key, command, user)

		iterator := 0
		flag := false

		if len(array) == 0 {
			return false
		}

		for _, value := range array {

			var keysExpected string = "5"

			if value != keysExpected {
				t.Log("INFO. Lines at output: " + value + " (should be 3 lines more than keys expected)")
				if value != "3" {
					flag = true
				}
				iterator++
			}
		}

		if flag {
			t.Log("Seems that init script is still running, waiting...")
			time.Sleep(30 * time.Second)
			continue
		}

		if iterator != 2 {
			t.Error("ERROR! Keys count not matched. There should be exactly 5 keys on exactly 1 node.")
			return false
		}
		return true
	}
	t.Log("Some keystore are either empty or not full")
	return false
}

// ConsulCheck check consul members
func ConsulCheck(t *testing.T, publicIPs []string, key *ssh.KeyPair, user string) bool {

	command := "consul members --status alive | wc -l"
	array := NodeQuery(t, publicIPs, key, command, user)

	if len(array) == 0 {
		return false
	}

	for _, value := range array {

		intValue, err := strconv.Atoi(value)

		if err != nil {
			t.Error("ERROR! " + err.Error())
			return false
		}

		var instanceCountExpected int = len(publicIPs) + 1

		if intValue != instanceCountExpected {
			t.Errorf(
				"ERROR! Consul node count not matched. One of the nodes responded the following healthy"+
					"instance count: %d, while there should be %d instances",
				intValue,
				instanceCountExpected,
			)
			return false
		}
	}

	return true

}

// ConsulLockCheck check consul locking
func ConsulLockCheck(t *testing.T, publicIPs []string, key *ssh.KeyPair, user string) bool {

	command := "consul kv export | grep \"prefix/.lock\" | wc -l"

	for retry := 1; retry <= 5; retry++ {

		array := NodeQuery(t, publicIPs, key, command, user)

		if len(array) == 0 {
			if retry == 5 {
				return false
			}
			t.Log("Cannot get node results. Blank slice. Waiting...")
			time.Sleep(60 * time.Second)
			continue
		}

		for _, value := range array {

			intValue, err := strconv.Atoi(value)

			if err != nil {
				if retry == 5 {
					t.Error("ERROR! " + err.Error())
					return false
				}
				t.Logf("Cannot parse consul response: %s", err)
				time.Sleep(60 * time.Second)
				break
			}

			if intValue != 1 {
				if retry == 5 {
					t.Errorf("ERROR! Error while retrieving Consul lock. Got: %d locks, should be exactly 1 lock.", intValue)
					return false
				}
				t.Logf("Got wrong lock value %d. Retry...", intValue)
				time.Sleep(60 * time.Second)
				break
			}

		}
	}

	return true

}

//PolkadotCheck checks polkadot system
func PolkadotCheck(t *testing.T, publicIPs []string, key *ssh.KeyPair, user string) bool {

	command := "curl -s -H \"Content-Type: application/json\" -d '{\"id\":1, \"jsonrpc\":\"2.0\", \"method\": \"system_health\", \"params\":[]}' http://localhost:9933"
	array := NodeQuery(t, publicIPs, key, command, user)

	t.Log("Leaders output: " + strings.Join(array, ","))

	if len(array) == 0 {
		return false
	}

	// Parse JSON and verify that node has healthy state
	for _, v := range array {

		type resultHealth struct {
			Peers           int  `json:"peers,omitempty"`
			ShouldHavePeers bool `json:"shouldHavePeers,omitempty"`
			IsSyncing       bool `json:"isSyncing,omitempty"`
		}

		type Health struct {
			Result resultHealth `json:"result,omitempty"`
		}

		var result Health
		err := json.Unmarshal([]byte(v), &result)

		if err != nil {
			t.Error("ERROR! " + err.Error())
			return false
		}

		if result.Result.Peers < 2 && result.Result.ShouldHavePeers {
			t.Error("ERROR! Node does not have enough peers")
			return false
		}
	}

	return true

}
