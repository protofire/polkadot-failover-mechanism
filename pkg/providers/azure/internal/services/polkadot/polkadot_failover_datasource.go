package polkadot

import (
	"context"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/timeouts"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePolkadotFailOver() *schema.Resource {

	polkadotSchema := resource.GetPolkadotSchema()
	polkadotSchema[ResourceGroupFieldName] = azure.SchemaResourceGroupName()

	return &schema.Resource{

		ReadContext: dateSourcePolkadotFailOverRead,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 90),
			Update: schema.DefaultTimeout(time.Minute * 90),
			Read:   schema.DefaultTimeout(time.Minute * 30),
			Delete: schema.DefaultTimeout(time.Minute * 30),
		},

		Schema: polkadotSchema,
	}
}

func dateSourcePolkadotFailOverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	failover := &AzureFailover{}
	_ = failover.FromSchema(d)

	if failover.IsDistributedMode() {
		log.Printf("[DEBUG] failover: Read. Failover mode is %q. Using predefined number of instances", failover.FailoverMode)
		failover.SetCounts(failover.Instances...)
		return failover.SetSchemaValuesDiag(d)
	}

	positions := make([]int, len(failover.Locations))

	client := meta.(*clients.Client)

	ctx, cancel := timeouts.ForCreate(ctx, d)
	defer cancel()

	features := meta.(*clients.Client).Features.PolkadotFailOverFeature

	log.Printf("[DEBUG] failover: Read. Getting instances list...")

	vmScaleSetNames, err := azure.GetVMScaleSetNames(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		failover.ResourceGroup,
		failover.Prefix,
	)

	if err != nil {
		return diag.Errorf("[ERROR] failover: Cannot get VM scale sets: %v", err)
	}

	log.Printf("[DEBUG] failover: Read. Found %d VM scale sets", len(vmScaleSetNames))

	if len(vmScaleSetNames) == 0 {
		failover.FillDefaultCountsIfNotSet()
		id, err := failover.ID()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		return failover.SetSchemaValuesDiag(d)
	}

	aggregator := insights.Maximum

	validator, err := azure.GetCurrentValidator(
		ctx,
		client.Polkadot.MetricsClient,
		vmScaleSetNames,
		failover.ResourceGroup,
		failover.MetricName,
		failover.MetricNameSpace,
		aggregator,
	)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] failover: Read. Found validator scale set %q, host %q", validator.ScaleSetName, validator.Hostname)

	vms, err := azure.GetVirtualMachineScaleSetVMsWithClient(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		client.Polkadot.VMScaleSetVMsClient,
		failover.Prefix,
		failover.ResourceGroup,
	)

	if err != nil {
		return diag.Errorf("[ERROR] failover: Cannot get scale set VMs: %v", err)
	}

	locationIDx := getValidatorLocation(vms, failover.Locations, validator.ScaleSetName)

	if locationIDx == -1 {
		locationIDx = 0
	}

	positions[locationIDx] = 1

	if features.DeleteVmsWithAPIInSingleMode {
		if err := deleteVms(ctx, client, failover, vmScaleSetNames, vms, validator, false); err != nil {
			return diag.FromErr(err)
		}
	}

	failover.SetCounts(positions...)
	failover.FillDefaultCountsIfNotSet()
	id, err := failover.ID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return failover.SetSchemaValuesDiag(d)

}
