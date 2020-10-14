package helpers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
)

// NodeQuery Supplementary function: perform given SSH query on the node
func NodeQuery(t *testing.T, publicIPs []string, key *ssh.KeyPair, command, user string) []string {

	var resultArray []string

	for _, publicInstanceIP := range publicIPs {

		publicHost := ssh.Host{
			Hostname:    publicInstanceIP,
			SshKeyPair:  key,
			SshUserName: user,
		}

		// It can take a minute or so for the Instance to boot up, so retry a few times
		maxRetries := 10
		timeBetweenRetries := 5 * time.Second
		description := fmt.Sprintf("SSH to public host %s", publicInstanceIP)

		t.Log("DEBUG. Querying instance " + publicInstanceIP + " with command `" + command + "`")
		// Verify that we can SSH to the Instance and run commands
		result := retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
			result, err := ssh.CheckSshCommandE(t, publicHost, command)

			if err != nil {
				return "", err
			}

			return strings.TrimSpace(result), nil
		})

		t.Log("DEBUG. Command output: " + result)
		resultArray = append(resultArray, result)
	}
	return resultArray
}
