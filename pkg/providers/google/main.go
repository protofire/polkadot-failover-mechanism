package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/google/google"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: google.Provider})
}
