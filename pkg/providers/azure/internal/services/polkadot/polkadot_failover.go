package polkadot

import (
	"context"
	"log"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/location"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/failover"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/failover/tags"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/timeouts"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/common"
)

func dataSourcePolkadotFailOver() *schema.Resource {

	return &schema.Resource{

		ReadContext: dateSourcePolkadotFailOverRead,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 30),
			Update: schema.DefaultTimeout(time.Minute * 60),
			Read:   schema.DefaultTimeout(time.Minute * 30),
			Delete: schema.DefaultTimeout(time.Minute * 30),
		},

		Schema: map[string]*schema.Schema{

			"resource_group_name": azure.SchemaResourceGroupName(),

			"instances": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 3,
				MinItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},

			"locations": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 3,
				MinItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"prefix": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validate.Prefix),
			},

			"tags": tags.Schema(),

			"metric_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
			},

			"metric_namespace": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
			},

			"failover_mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringInSlice([]string{
					string(common.FailOverModeDistributed),
					string(common.FailOverModeSingle),
				}, false)),
			},

			"failover_instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},

			"primary_count": {
				Type:        schema.TypeInt,
				Description: "Polkadot nodes count in primary location. Primary locations is first one in locations parameter",
				Computed:    true,
			},

			"secondary_count": {
				Type:        schema.TypeInt,
				Description: "Polkadot nodes count in secondary location. Secondary locations is second one in locations parameter",
				Computed:    true,
			},

			"tertiary_count": {
				Type:        schema.TypeInt,
				Description: "Polkadot nodes count in tertiary location. Tertiary locations is third one in locations parameter",
				Computed:    true,
			},
		},
	}
}

func dateSourcePolkadotFailOverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	failoverMode := common.FailOverMode(d.Get("failover_mode").(string))
	log.Printf("[DEBUG]. Failover mode: %s", failoverMode)

	instanceLocationsRaw := d.Get("locations").([]interface{})
	instanceLocations := resource.ExpandString(instanceLocationsRaw)
	log.Printf("[DEBUG]. Regions: %#v", instanceLocations)

	instancesRaw := d.Get("instances").([]interface{})
	instances := resource.ExpandInt(instancesRaw)
	log.Printf("[DEBUG]. Instances per regions: %#v", instances)

	prefix := d.Get("prefix").(string)
	metricName := d.Get("metric_name").(string)
	metricNameSpace := d.Get("metric_namespace").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	instanceLocations = location.NormalizeSlice(instanceLocations)

	var diagnostics diag.Diagnostics

	d.SetId(resource.PrepareID(resourceGroup, prefix, string(failoverMode), metricName, metricNameSpace))

	if failoverMode == common.FailOverModeDistributed {
		log.Printf("[DEBUG]. Failover mode is %q. Using predefined number of nstances", failoverMode)
		return resource.SetSchemaValues(d, diagnostics, instances[0], instances[1], instances[2])
	}

	client := meta.(*clients.Client)

	ctx, cancel := timeouts.ForCreate(ctx, d)
	defer cancel()

	features := meta.(*clients.Client).Features.PolkadotFailOverFeature

	vmScaleSetNames, err := azure.GetVMScaleSetNames(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		resourceGroup,
		prefix,
	)

	if err != nil {
		return diag.Errorf("[ERROR]. Cannot get VM scale sets: %v", err)
	}

	log.Printf("[DEBUG]. Got %d VM scale sets", len(vmScaleSetNames))

	if len(vmScaleSetNames) == 0 {
		counts := failover.CalculateInstancesForSingleFailOverMode(instances)
		return resource.SetSchemaValues(d, diagnostics, counts[0], counts[1], counts[2])
	}

	aggregator := insights.Maximum

	validator, err := azure.GetCurrentValidator(
		ctx,
		client.Polkadot.MetricsClient,
		vmScaleSetNames,
		resourceGroup,
		metricName,
		metricNameSpace,
		aggregator,
	)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG]. Cot validator scale set %q, host %q", validator.ScaleSetName, validator.Hostname)

	vms, err := azure.GetVirtualMachineScaleSetVMsWithClient(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		client.Polkadot.VMScaleSetVMsClient,
		prefix,
		resourceGroup,
	)

	if err != nil {
		return diag.Errorf("[ERROR]. Cannot get scale set VMs: %+v", err)
	}

	_, locationIDx := getValidatorLocations(vms, instanceLocations, validator.ScaleSetName)

	if locationIDx == -1 {
		locationIDx = 0
	}

	singleModeInstances := make([]int, len(instanceLocations))
	singleModeInstances[locationIDx] = 1

	log.Printf("[DEBUG]. We will delete instance with API requests: true/false: %t", features.DeleteVmsWithAPIInSingleMode)

	if features.DeleteVmsWithAPIInSingleMode {
		vmsToDelete := getVmsToDelete(vms, instanceLocations, singleModeInstances, validator.ScaleSetName, validator.Hostname)
		for vmSSName, vmsIDs := range vmsToDelete {
			err := deleteVMs(ctx, client.Polkadot.VMScaleSetsClient, resourceGroup, vmSSName, vmsIDs)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resource.SetSchemaValues(d, diagnostics, singleModeInstances[0], singleModeInstances[1], singleModeInstances[2])

}
