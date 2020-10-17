package provider

import (
	"context"
	"testing"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/clients"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var providerFactories map[string]func() (*schema.Provider, error)

func init() {
	providerFactories = make(map[string]func() (*schema.Provider, error))
	providerFactories["polkadot"] = func() (*schema.Provider, error) { // nolint
		return TestAzurePolkadotProvider(), nil
	}
}

func TestProvider(t *testing.T) {
	prov, err := providerFactories["polkadot"]()
	require.NoError(t, err)
	err = prov.InternalValidate()
	require.NoError(t, err)
}

func TestPolkadotConfigureCheck(t *testing.T) {
	ctx := context.Background()
	prov, err := providerFactories["polkadot"]()
	require.NoError(t, err)
	diagnostics := prov.Configure(ctx, terraform.NewResourceConfigRaw(map[string]interface{}{
		"client_id":                          "x",
		"client_secret":                      "x",
		"subscription_id":                    "x",
		"tenant_id":                          "x",
		"use_msi":                            true,
		"skip_provider_registration":         true,
		"delete_vms_with_api_in_single_mode": false,
	}))
	require.Len(t, diagnostics, 0)
	ds := prov.DataSources()[0]
	require.Equal(t, "polkadot_failover", ds.Name)
	futures := prov.Meta().(*clients.Client).Features.PolkadotFailOverFeature
	require.False(t, futures.DeleteVmsWithAPIInSingleMode)
}
