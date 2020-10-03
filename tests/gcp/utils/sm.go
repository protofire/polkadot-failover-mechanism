package utils

// This file contains all the supplementary functions that are required to query SM API

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func getParameterValue(ctx context.Context, client *secretmanager.Client, keyName string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: keyName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", err
	}
	data := result.Payload.Data
	return string(data), nil
}

func getKeyName(project string, paths ...string) string {
	return fmt.Sprintf("projects/%s/secrets/%s/versions/latest", project, strings.Join(paths, "_"))
}

// valueComparator Supplementary function:
// Checks that given parameter in each parameter exists and has the right value
func valueComparator(ctx context.Context, t *testing.T, client *secretmanager.Client, keyName, expectedValue string) int {
	smValue, err := getParameterValue(ctx, client, keyName)
	if !assert.NoErrorf(t, err, "Error getting SM key %s", keyName) {
		return 0
	}
	if !assert.Equalf(t, smValue, expectedValue, "ERROR! No match for SM parameter %s. Expected value: %s. ActualValue: %s", keyName, expectedValue, smValue) {
		return 0
	}
	return 1
}

// SMCheck checks sm parameters
func SMCheck(t *testing.T, prefix, project string) bool {

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	require.NoError(t, err)

	result := valueComparator(ctx, t, client, getKeyName(project, prefix, "cpulimit"), "1") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "ramlimit"), "1") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "name"), "test") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "key1", "type"), "gran") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "key1", "seed"), "favorite liar zebra assume hurt cage any damp inherit rescue delay panic") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "key1", "key"), "0x6ce96ae5c300096b09dbd4567b0574f6a1281ae0e5cfe4f6b0233d1821f6206b") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "key2", "type"), "aura") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "key2", "seed"), "expire stage crawl shell boss any story swamp skull yellow bamboo copy") *
		valueComparator(ctx, t, client, getKeyName(project, prefix, "key2", "key"), "0x3ff0766f9ebbbceee6c2f40d9323164d07e70c70994c9d00a9512be6680c2394")

	return result == 1

}

// SMClean cleans all SM keys with prefix
func SMClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize secrets client: %#w", err)
	}

	var secretNames []string

	req := &secretmanagerpb.ListSecretsRequest{Parent: fmt.Sprintf("projects/%s", project)}
	secretIterator := client.ListSecrets(ctx, req)

	for {
		secret, err := secretIterator.Next()
		if err != nil {
			if !errors.Is(err, iterator.Done) {
				return fmt.Errorf("Got secret API error: %#w", err)
			}
			break
		}
		log.Printf("got secret: %s\n", secret.Name)

		secretName := lastPartOnSplit(secret.Name, "/")

		if strings.HasPrefix(secretName, fmt.Sprintf("%s_", prefix)) {
			secretNames = append(secretNames, secret.Name)
		}
	}

	if len(secretNames) == 0 {
		log.Println("Not found secrets to delete")
		return nil
	}

	log.Printf("Prepared keys to delete:\n%s\n", strings.Join(secretNames, "\n"))
	c := make(chan error)
	wg := &sync.WaitGroup{}

	for _, key := range secretNames {

		wg.Add(1)

		go func(key string, wg *sync.WaitGroup) {

			defer wg.Done()

			req := &secretmanagerpb.DeleteSecretRequest{
				Name: key,
			}

			log.Printf("Deleting secret: %s", key)

			if dryRun {
				return
			}

			if err := client.DeleteSecret(ctx, req); err != nil {
				c <- fmt.Errorf("Could not delete secret key %s. %#w", key, err)
				return
			}
			log.Printf("Successfully deleted key: %s\n", key)

		}(key, wg)

	}

	go func() {
		defer close(c)
		wg.Wait()
	}()

	var result *multierror.Error

	for err := range c {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()

}
