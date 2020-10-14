package helpers

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/stretchr/testify/require"
)

// GenerateSSHKeys generate ot read ssh keys
func GenerateSSHKeys(t *testing.T) *ssh.KeyPair {

	sshPrivateKeyFile := os.Getenv("SSH_PRIVATE_KEY_FILE")
	sshPublicKeyFile := os.Getenv("SSH_PUBLIC_KEY_FILE")
	sshPrivateKey := os.Getenv("SSH_PRIVATE_KEY")
	sshPublicKey := os.Getenv("SSH_PUBLIC_KEY")

	keyPair := &ssh.KeyPair{}

	if len(sshPrivateKeyFile) > 0 && len(sshPublicKeyFile) > 0 {
		log.Print("Using ssh keys from SSH_PRIVATE_KEY_FILE and SSH_PUBLIC_KEY_FILE environment variables")
		content, err := ioutil.ReadFile(sshPrivateKeyFile)
		require.NoErrorf(t, err, "Path %s from SSH_PRIVATE_KEY_FILE does not exist: %w", sshPrivateKeyFile, err)
		keyPair.PrivateKey = string(content)
		content, err = ioutil.ReadFile(sshPublicKeyFile)
		require.NoErrorf(t, err, "Path %s from SSH_PUBLIC_KEY_FILE does not exist: %w", sshPublicKeyFile, err)
		keyPair.PublicKey = string(content)
		return keyPair
	} else if len(sshPrivateKey) > 0 && len(sshPublicKey) > 0 {
		log.Print("Using ssh keys from SSH_PRIVATE_KEY and SSH_PUBLIC_KEY environment variables")
		keyPair.PrivateKey = sshPrivateKey
		keyPair.PublicKey = sshPublicKey
		return keyPair
	}
	log.Print("Generating new ssh key pair")
	return ssh.GenerateRSAKeyPair(t, 4096)
}
