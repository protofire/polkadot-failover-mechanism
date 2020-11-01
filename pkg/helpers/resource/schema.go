package resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/failover/tags"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"
)

func GetPolkadotSchema() map[string]*schema.Schema {

	return map[string]*schema.Schema{

		TagsFieldName: tags.Schema(),

		InstancesFieldName: {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			MaxItems: 3,
			MinItems: 3,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},

		LocationsFieldName: {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: true,
			MaxItems: 3,
			MinItems: 3,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		PrefixFieldName: {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.DiagFunc(validate.Prefix),
		},

		MetricNameFieldName: {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
		},

		MetricNamespaceFieldName: {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
		},

		FailoverModeFieldName: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			ValidateDiagFunc: validate.DiagFunc(validation.StringInSlice([]string{
				string(FailOverModeDistributed),
				string(FailOverModeSingle),
			}, false)),
		},

		FailoverInstancesFieldName: {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},

		PrimaryCountFieldName: {
			Type:        schema.TypeInt,
			Description: "Polkadot nodes count in primary location. Primary locations is first one in locations parameter",
			Computed:    true,
		},

		SecondaryCountFieldName: {
			Type:        schema.TypeInt,
			Description: "Polkadot nodes count in secondary location. Secondary locations is second one in locations parameter",
			Computed:    true,
		},

		TertiaryCountFieldName: {
			Type:        schema.TypeInt,
			Description: "Polkadot nodes count in tertiary location. Tertiary locations is third one in locations parameter",
			Computed:    true,
		},
	}
}
