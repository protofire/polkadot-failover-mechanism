package tags

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"
)

// SchemaDataSource returns the Schema which should be used for Tags on a Data Source
func SchemaDataSource() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

// ForceNewSchema returns the Schema which should be used for Tags when changes
// require recreation of the resource
func ForceNewSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Optional:         true,
		ForceNew:         true,
		ValidateDiagFunc: validate.DiagFunc(Validate),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

// Schema returns the Schema used for Tags
func Schema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Optional:         true,
		ValidateDiagFunc: validate.DiagFunc(Validate),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

// Schema returns the Schema used for Tags
func SchemaEnforceLowerCaseKeys() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Optional:         true,
		ValidateDiagFunc: validate.DiagFunc(EnforceLowerCaseKeys),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}
