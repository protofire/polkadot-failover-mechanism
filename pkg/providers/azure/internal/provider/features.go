package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/features"
)

func schemaFeatures() *schema.Schema {
	resultFeatures := map[string]*schema.Schema{
		"delete_vms_with_api_requests": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: resultFeatures,
		},
	}
}

func expandFeatures(input []interface{}) features.UserFeatures {
	// these are the defaults if omitted from the config
	expandedFeatures := features.UserFeatures{
		// NOTE: ensure all nested objects are fully populated
		PolkadotFailOverFeature: features.PolkadotFailOverFeatures{
			DeleteVmsWithAPIInSingleMode: true,
		},
	}

	if len(input) == 0 || input[0] == nil {
		return expandedFeatures
	}

	val := input[0].(map[string]interface{})

	if v, ok := val["delete_vms_with_api_requests"]; ok {
		expandedFeatures.PolkadotFailOverFeature.DeleteVmsWithAPIInSingleMode = v.(bool)
	}

	return expandedFeatures
}
