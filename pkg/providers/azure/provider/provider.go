package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/provider"
)

func Provider() *schema.Provider {
	return provider.AzurePolkadotProvider()
}
