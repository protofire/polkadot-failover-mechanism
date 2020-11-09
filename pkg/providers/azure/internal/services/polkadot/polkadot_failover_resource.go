package polkadot

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	helperErrors "github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/timeouts"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePolkadotFailOver() *schema.Resource {

	polkadotSchema := resource.GetPolkadotSchema()
	polkadotSchema[ResourceGroupFieldName] = azure.SchemaResourceGroupName()

	return &schema.Resource{

		ReadContext:   resourcePolkadotFailoverRead,
		CreateContext: resourcePolkadotFailoverCreateOrUpdate,
		UpdateContext: resourcePolkadotFailoverCreateOrUpdate,
		DeleteContext: resourcePolkadotFailoverDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 90),
			Update: schema.DefaultTimeout(time.Minute * 90),
			Read:   schema.DefaultTimeout(time.Minute * 30),
			Delete: schema.DefaultTimeout(time.Minute * 30),
		},

		Schema: polkadotSchema,
	}
}

func deleteVms(
	ctx context.Context,
	client *clients.Client,
	failover *AzureFailover,
	vms azure.VMSMap,
	validator azure.Validator,
	updateVMssCapacity bool,
) error {

	vmsToDelete := getVmsToDelete(vms, validator.Hostname)
	log.Printf("[DEBUG] failover: Create. We will delete instances %s with API requests", vmsToDelete)

	var vmsToDeleteInf []interface{}
	for _, vmss := range vmsToDelete {
		vmsToDeleteInf = append(vmsToDeleteInf, vmss)
	}

	if err := fanout.ReadErrorsChannel(fanout.ConcurrentResponseErrors(ctx, func(ctx context.Context, value interface{}) error {
		vmss := value.(vmssWithInstances)
		return azure.DeleteVMs(
			ctx,
			client.Polkadot.VMScaleSetsClient,
			failover.ResourceGroup,
			vmss.vmssName,
			vmss.vmsIDs,
			updateVMssCapacity,
		)
	}, vmsToDeleteInf...)); err != nil {
		return err
	}

	waitForCount := 1
	if validator.ScaleSetName == "" {
		waitForCount = 0
	}

	log.Printf("[DEBUG] failover: Create. Waiting for VMs count: %d", waitForCount)

	vmss, err := azure.WaitForVirtualMachineScaleSetVMsWithClient(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		client.Polkadot.VMScaleSetVMsClient,
		failover.Prefix,
		failover.ResourceGroup,
		waitForCount,
		5,
	)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] failover: Create. Ensured VMs count: %d", waitForCount)

	var vmScaleSetNames []string

	for name, vms := range vmss {
		if len(vms) > 0 {
			vmScaleSetNames = append(vmScaleSetNames, name)
		}
	}

	if validator.ScaleSetName != "" && len(vmScaleSetNames) > 0 {

		log.Printf("[DEBUG] failover: Create. Waiting for validator...")

		validator, err := azure.WaitForValidator(
			ctx,
			client.Polkadot.MetricsClient,
			vmScaleSetNames,
			failover.ResourceGroup,
			failover.MetricName,
			failover.MetricNameSpace,
			5,
		)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] failover: Create. Ensured validator: %#v", validator)
	}

	return nil

}

func resourcePolkadotFailoverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	failover := &AzureFailover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		return diag.FromErr(err)
	}

	if !failover.Initialized() {
		d.SetId("")
		return nil
	}

	if failover.IsDistributedMode() {
		log.Printf(
			"[DEBUG] failover: Read. Failover mode is %q. Using predefined number of instances %d",
			failover.FailoverMode,
			failover.Instances,
		)
		failover.SetCounts(failover.Instances...)
		return failover.SetSchemaValuesDiag(d)
	}

	log.Printf("[DEBUG] failover: Read. Failover mode is %q", failover.FailoverMode)

	client := meta.(*clients.Client)

	ctx, cancel := timeouts.ForRead(ctx, d)
	defer cancel()

	positions := make([]int, len(failover.Locations))

	vmss, err := azure.GetVirtualMachineScaleSetVMsWithClient(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		client.Polkadot.VMScaleSetVMsClient,
		failover.Prefix,
		failover.ResourceGroup,
	)

	if err != nil {
		return diag.Errorf("[ERROR] failover: Cannot get scale set VMs: %v", err)
	}

	log.Printf("[DEBUG] failover: Read. Found %d virtual machines", vmss.Size())
	log.Printf("[DEBUG] failover: Read. Found %d virtual machines scale sets", len(vmss))

	var vmScaleSetNames []string

	for name, vms := range vmss {
		if len(vms) > 0 {
			vmScaleSetNames = append(vmScaleSetNames, name)
		}
	}

	validator, err := azure.GetCurrentValidator(
		ctx,
		client.Polkadot.MetricsClient,
		vmScaleSetNames,
		failover.ResourceGroup,
		failover.MetricName,
		failover.MetricNameSpace,
		insights.Maximum,
	)

	if err != nil {
		validatorError := &helperErrors.ValidatorError{}
		if errors.As(err, validatorError) {
			log.Printf("[WARNING] failover: Read. Cannot get validator: %s", validatorError)
		} else {
			log.Printf("[ERROR] failover: Read. Cannot get validator: %s", err)
			return diag.FromErr(err)
		}
	} else {
		log.Printf("[DEBUG] failover: Read. Found validator scale set %q, host %q", validator.ScaleSetName, validator.Hostname)
	}

	log.Printf("[DEBUG] failover: Read. Getting instances list...")

	locationIDx := getValidatorLocation(vmss, failover.Locations, validator.ScaleSetName)

	if locationIDx == -1 {
		locationIDx = 0
	}

	positions[locationIDx] = 1

	log.Printf("[DEBUG] failover: Read. Found instance numbers per region: %v", positions)
	failover.SetCounts(positions...)
	failover.FillDefaultCountsIfNotSet()
	log.Printf("[DEBUG] failover: Read. Set instance numbers per region: %v", failover.FailoverInstances)

	return failover.SetSchemaValuesDiag(d)

}

func resourcePolkadotFailoverCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*clients.Client)

	ctx, cancel := timeouts.ForCreate(ctx, d)
	defer cancel()

	features := meta.(*clients.Client).Features.PolkadotFailOverFeature

	failover := &AzureFailover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		return diag.FromErr(err)
	}

	if failover.IsDistributedMode() {
		log.Printf(
			"[DEBUG] failover: Create. Failover mode is %q. Using predefined number of instances: %d",
			failover.FailoverMode,
			failover.Instances,
		)
		failover.SetCounts(failover.Instances...)
		id, err := failover.ID()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		return resourcePolkadotFailoverRead(ctx, d, meta)
	}

	log.Printf("[DEBUG] failover: Create. Failover mode is %q", failover.FailoverMode)

	positions := make([]int, len(failover.Locations))

	vmss, err := azure.GetVirtualMachineScaleSetVMsWithClient(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		client.Polkadot.VMScaleSetVMsClient,
		failover.Prefix,
		failover.ResourceGroup,
	)

	if err != nil {
		return diag.Errorf("[ERROR] failover: Cannot get scale set VMs: %v", err)
	}

	log.Printf("[DEBUG] failover: Create. Found %d virtual machines", vmss.Size())
	log.Printf("[DEBUG] failover: Create. Found %d virtual machines scale sets", len(vmss))

	var vmScaleSetNames []string

	for name, vms := range vmss {
		if len(vms) > 0 {
			vmScaleSetNames = append(vmScaleSetNames, name)
		}
	}

	if len(vmScaleSetNames) == 0 {
		failover.SetCounts(positions...)
		failover.FillDefaultCountsIfNotSet()
		id, err := failover.ID()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		return resourcePolkadotFailoverRead(ctx, d, meta)
	}

	log.Printf("[DEBUG] failover: Create. Getting validator...")
	validator, err := azure.GetCurrentValidator(
		ctx,
		client.Polkadot.MetricsClient,
		vmScaleSetNames,
		failover.ResourceGroup,
		failover.MetricName,
		failover.MetricNameSpace,
		insights.Maximum,
	)

	if err != nil {
		validatorError := &helperErrors.ValidatorError{}
		if errors.As(err, validatorError) {
			log.Printf("[WARNING] failover: Create. Cannot get validator: %s", validatorError)
		} else {
			log.Printf("[ERROR] failover: Create. Cannot get validator: %s", err)
			return diag.FromErr(err)
		}
	} else {
		log.Printf("[DEBUG] failover: Create. Found validator scale set %q, host %q", validator.ScaleSetName, validator.Hostname)
	}

	if validator.ScaleSetName != "" {
		log.Printf("[DEBUG] failover: Create. Found validator %#v", validator)
	} else {
		log.Printf("[DEBUG] failover: Create. Did not find validator")
	}
	if features.DeleteVmsWithAPIInSingleMode {
		if err := deleteVms(ctx, client, failover, vmss, validator, false); err != nil {
			return diag.FromErr(err)
		}
		vmss, err := azure.GetVirtualMachineScaleSetVMsWithClient(
			ctx,
			client.Polkadot.VMScaleSetsClient,
			client.Polkadot.VMScaleSetVMsClient,
			failover.Prefix,
			failover.ResourceGroup,
		)

		if err != nil {
			return diag.Errorf("[ERROR] failover: Cannot get scale set VMs: %v", err)
		}

		log.Printf("[DEBUG] failover: Create. Found %d virtual machines", vmss.Size())
		log.Printf("[DEBUG] failover: Create. Found %d virtual machines scale sets", len(vmss))
	}

	if locationIDx := getValidatorLocation(vmss, failover.Locations, validator.ScaleSetName); locationIDx != -1 {
		positions[locationIDx] = 1
	}

	log.Printf("[DEBUG] failover: Create. Found instance numbers per region: %v", positions)
	failover.SetCounts(positions...)
	failover.FillDefaultCountsIfNotSet()
	log.Printf("[DEBUG] failover: Create. Set instance numbers per region: %v", failover.FailoverInstances)

	id, err := failover.ID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return resourcePolkadotFailoverRead(ctx, d, meta)

}

func resourcePolkadotFailoverDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
