package polkadot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Registration object
type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "Polkadot"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"Polkadot",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	dataSources := map[string]*schema.Resource{
		"polkadot_failover": dataSourcePolkadotFailOver(),
	}
	return dataSources
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{}
}
