package polkadot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/timeouts"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type metricNames map[string]int

func (m *metricNames) String() string {
	var parts []string
	for name, count := range *m {
		parts = append(parts, fmt.Sprintf("%s => %d", name, count))
	}
	return strings.Join(parts, ", ")
}

func (m *metricNames) metric() string {
	for name := range *m {
		return name
	}
	return ""
}

func dataSourcePolkadotMetricDefinition() *schema.Resource {

	return &schema.Resource{

		ReadContext: dateSourcePolkadotMetricDefinitionRead,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 60),
			Update: schema.DefaultTimeout(time.Minute * 60),
			Read:   schema.DefaultTimeout(time.Minute * 60),
			Delete: schema.DefaultTimeout(time.Minute * 60),
		},

		Schema: map[string]*schema.Schema{

			ResourceGroupFieldName: azure.SchemaResourceGroupName(),

			ScaleSetsFieldName: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 3,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			MetricOutputNameFieldName: {
				Type:     schema.TypeString,
				Computed: true,
			},

			MetricNameFieldName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
			},

			PrefixFieldName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validate.Prefix),
			},

			MetricNamespaceFieldName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
			},
		},
	}
}

func dateSourcePolkadotMetricDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	metricSource := MetricSource{}
	_ = metricSource.FromSchema(d)

	client := meta.(*clients.Client)

	ctx, cancel := timeouts.ForRead(ctx, d)
	defer cancel()

	vmssNames, err := azure.CheckVirtualMachineScaleSetVMsWithClient(
		ctx,
		client.Polkadot.VMScaleSetsClient,
		client.Polkadot.VMScaleSetVMsClient,
		metricSource.ResourceGroup,
		metricSource.ScaleSets...,
	)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf(
		"[DEBUG] failover: Metrics. Filtered virtual machine scale sets %s. Requested %s",
		vmssNames,
		metricSource.ScaleSets,
	)

	if len(vmssNames) == 0 {
		metricSource.SetMetric(metricSource.MetricName)
		id, err := metricSource.ID()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		return metricSource.SetSchemaValuesDiag(d)
	}

	vmScaleSetToMetricName, err := azure.WaitValidatorMetricNamesForMetricNamespace(
		ctx,
		client.Polkadot.MetricDefinitionsClient,
		vmssNames,
		metricSource.ResourceGroup,
		metricSource.MetricName,
		metricSource.MetricNameSpace,
		5,
		20,
	)

	if err != nil {
		return diag.FromErr(err)
	}

	metricNamesCount := make(metricNames)

	for _, metricName := range vmScaleSetToMetricName {
		metricNamesCount[metricName]++
	}

	if len(metricNamesCount) > 1 {
		return diag.Errorf(
			"found more than 1 metrics %d for namespace %q and scale sets %s: %s",
			len(metricNamesCount),
			metricSource.MetricNameSpace,
			metricSource.ScaleSets,
			metricNamesCount.String(),
		)
	}

	if len(metricNamesCount) == 0 {
		return diag.Errorf(
			"not found metrics for namespace %q and scale sets %s",
			metricSource.MetricNameSpace,
			metricSource.ScaleSets,
		)
	}

	log.Printf(
		"[DEBUG] failover: Metrics. Found metric definition for metric namespace %q and virtual machines scale sets %s. Metric name: %q",
		metricSource.MetricNameSpace,
		metricSource.ScaleSets,
		metricNamesCount.metric(),
	)

	metricSource.SetMetric(metricNamesCount.metric())

	id, err := metricSource.ID()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return metricSource.SetSchemaValuesDiag(d)

}
