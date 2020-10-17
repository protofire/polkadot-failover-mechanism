package azure

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/emicklei/go-restful/log"

	kv "github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2016-10-01/keyvault"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
)

func getKeyVaultClient(subscriptionID string) (keyvault.VaultsClient, error) {
	client := keyvault.NewVaultsClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, err
	}
	client.Authorizer = auth
	return client, nil
}

// nolint
func getKeyVaultSecretsClient() (kv.BaseClient, error) {
	client := kv.New()
	auth, err := getVaultAuthorizer()
	if err != nil {
		return client, err
	}
	client.Authorizer = auth
	return client, nil
}

func getVaultNames(ctx context.Context, client *keyvault.VaultsClient) ([]string, error) {

	result, err := client.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	var vaults []string

	for _, res := range result.Values() {
		vaults = append(vaults, path.Base(*res.ID))
	}

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		for _, res := range result.Values() {
			vaults = append(vaults, path.Base(*res.ID))
		}

	}

	return vaults, nil

}

func getVaultSecrets(ctx context.Context, client *kv.BaseClient, vaultURL string) (map[string]string, error) {

	result, err := client.GetSecrets(ctx, vaultURL, nil)

	if err != nil {
		return nil, err
	}

	items := make(map[string]string)

	for _, secret := range result.Values() {
		secretName := path.Base(*secret.ID)
		secretBundle, err := client.GetSecret(ctx, vaultURL, secretName, "")
		if err != nil {
			return nil, err
		}
		items[secretName] = *secretBundle.Value
	}
	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		for _, secret := range result.Values() {
			secretName := path.Base(*secret.ID)
			secretBundle, err := client.GetSecret(ctx, vaultURL, secretName, "")
			if err != nil {
				return nil, err
			}
			items[secretName] = *secretBundle.Value
		}
	}
	return items, nil

}

// getKeyVaultSecrets returns map of secret name to secret value
func getKeyVaultSecrets(prefix, subscriptionID, resourceGroup string) (map[string]string, error) {

	vaultClient, err := getKeyVaultClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	vaults, err := getVaultNames(ctx, &vaultClient)

	if err != nil {
		return nil, err
	}

	helpers.FilterStrings(&vaults, func(item string) bool {
		return strings.HasPrefix(item, helpers.GetPrefix(prefix))
	})

	if len(vaults) == 0 {
		return nil, fmt.Errorf("No secrets found")
	}

	vaultName := vaults[0]

	vault, err := vaultClient.Get(ctx, resourceGroup, vaultName)
	if err != nil {
		return nil, err
	}
	vaultURL := *vault.Properties.VaultURI

	vaultSecretsClient, err := getKeyVaultSecretsClient()

	if err != nil {
		return nil, err
	}

	items, err := getVaultSecrets(ctx, &vaultSecretsClient, vaultURL)

	if err != nil {
		return nil, err
	}

	return items, nil
}

func checkKeyName(items map[string]string, value string, keys ...string) error {
	keyName := strings.Join(keys, "-")
	val, ok := items[keyName]
	if !ok {
		return fmt.Errorf("cannot find vault secret: %q", keyName)
	}
	if val != value {
		return fmt.Errorf("value key %q value %q is not equal %q", keyName, val, value)
	}
	return nil
}

// SMCheck checks key vault keys
func SMCheck(prefix, subscriptionID, resourceGroup string) error {

	items, err := getKeyVaultSecrets(prefix, subscriptionID, resourceGroup)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG]. Found vault secrets: %#v", items)

	checkParams := [][]string{
		{
			"1", "polkadot", prefix, "ramlimit",
		},
		{
			"1", "polkadot", prefix, "cpulimit",
		},
		{
			"test", "polkadot", prefix, "name",
		},
		{
			"gran", "polkadot", prefix, "keys", "key1", "type",
		},
		{
			"favorite liar zebra assume hurt cage any damp inherit rescue delay panic", "polkadot", prefix, "keys", "key1", "seed",
		},
		{
			"0x6ce96ae5c300096b09dbd4567b0574f6a1281ae0e5cfe4f6b0233d1821f6206b", "polkadot", prefix, "keys", "key1", "key",
		},
		{
			"aura", "polkadot", prefix, "keys", "key2", "type",
		},
		{
			"expire stage crawl shell boss any story swamp skull yellow bamboo copy", "polkadot", prefix, "keys", "key2", "seed",
		},
		{
			"0x3ff0766f9ebbbceee6c2f40d9323164d07e70c70994c9d00a9512be6680c2394", "polkadot", prefix, "keys", "key2", "key",
		},
	}

	for _, keys := range checkParams {
		if err := checkKeyName(items, keys[0], keys[1:]...); err != nil {
			return err
		}
	}
	return nil

}
