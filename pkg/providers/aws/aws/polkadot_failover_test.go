package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var providerFactories map[string]func() (*schema.Provider, error)

func init() {
	providerFactories = make(map[string]func() (*schema.Provider, error))
	providerFactories["polkadot"] = func() (*schema.Provider, error) { // nolint
		return Provider(), nil
	}
}

func TestValidateProvider(t *testing.T) {
	prov, err := providerFactories["polkadot"]()
	require.NoError(t, err)
	err = prov.InternalValidate()
	require.NoError(t, err)
}

func TestPolkadotConfigureCheck(t *testing.T) {
	ctx := context.Background()
	prov, err := providerFactories["polkadot"]()
	require.NoError(t, err)
	diagnostics := prov.Configure(ctx, terraform.NewResourceConfigRaw(map[string]interface{}{}))
	require.Len(t, diagnostics, 0)
	ds := prov.Resources()[0]
	require.Equal(t, "polkadot_failover", ds.Name)
}
