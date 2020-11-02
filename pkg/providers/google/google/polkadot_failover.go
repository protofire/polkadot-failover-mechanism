package google

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/gcp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/failover/tags"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"
)

func resourcePolkadotFailover() *schema.Resource {
	return &schema.Resource{

		ReadContext:   resourcePolkadotFailoverRead,
		CreateContext: resourcePolkadotFailoverCreateOrUpdate,
		UpdateContext: resourcePolkadotFailoverCreateOrUpdate,
		DeleteContext: resourcePolkadotFailoverDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 30),
			Update: schema.DefaultTimeout(time.Minute * 60),
			Read:   schema.DefaultTimeout(time.Minute * 30),
			Delete: schema.DefaultTimeout(time.Minute * 30),
		},

		Schema: map[string]*schema.Schema{

			resource.InstancesFieldName: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 3,
				MinItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},

			resource.LocationsFieldName: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 3,
				MinItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			resource.PrefixFieldName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validate.Prefix),
			},

			resource.TagsFieldName: tags.Schema(),

			resource.MetricNameFieldName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
			},

			resource.MetricNamespaceFieldName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringIsNotEmpty),
			},

			resource.FailoverModeFieldName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validate.DiagFunc(validation.StringInSlice([]string{
					string(resource.FailOverModeDistributed),
					string(resource.FailOverModeSingle),
				}, false)),
			},

			resource.FailoverInstancesFieldName: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},

			resource.PrimaryCountFieldName: {
				Type:        schema.TypeInt,
				Description: "Polkadot nodes count in primary location. Primary locations is first one in locations parameter",
				Computed:    true,
			},

			resource.SecondaryCountFieldName: {
				Type:        schema.TypeInt,
				Description: "Polkadot nodes count in secondary location. Secondary locations is second one in locations parameter",
				Computed:    true,
			},

			resource.TertiaryCountFieldName: {
				Type:        schema.TypeInt,
				Description: "Polkadot nodes count in tertiary location. Tertiary locations is third one in locations parameter",
				Computed:    true,
			},
		},
	}
}

func resourcePolkadotFailoverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Config)
	failover := &resource.GCPFailover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		return diag.FromErr(err)
	}

	if failover.Project == "" {
		failover.Project, err = getProject(d, config)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if !failover.Initialized() {
		d.SetId("")
		return nil
	}

	if failover.IsDistributedMode() {
		log.Printf("[DEBUG] failover: Failover mode is %q. Using predefined number of instances", failover.FailoverMode)
		failover.SetCounts(failover.Instances...)
		if err := failover.SetSchemaValues(d); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	log.Printf("[DEBUG] failover: Failover mode is %q", failover.FailoverMode)

	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return diag.FromErr(err)
	}
	computeClient := config.NewComputeClient(userAgent)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Printf("[DEBUG] failover: Getting instances list...")

	instanceGroups, err := gcp.GetInstanceGroupManagersForRegions(
		ctx,
		computeClient,
		failover.Project,
		failover.Prefix,
		failover.Locations...,
	)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] failover: Found %d managent instance groups", len(instanceGroups))

	positions := make([]int, len(failover.Locations))

	for _, group := range instanceGroups {
		regionPosition := helpers.FindStrIndex(group.Region, failover.Locations)
		if regionPosition == -1 {
			log.Printf("[ERROR] failover: Cannot find region %s in locations list: %s", group.Region, strings.Join(failover.Locations, ", "))
			continue
		}
		positions[regionPosition] = len(group.Instances)
	}

	log.Printf("[DEBUG] failover: Found instance numbers per region: %v", positions)

	failover.SetCounts(positions...)

	failover.FillDefaultCountsIfNotSet()

	log.Printf("[DEBUG] failover: Set instance numbers per region: %v", failover.FailoverInstances)

	if err := failover.SetSchemaValues(d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePolkadotFailoverCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Config)
	failover := &resource.GCPFailover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		return diag.FromErr(err)
	}

	if failover.Project == "" {
		failover.Project, err = getProject(d, config)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[DEBUG] failover: Failover mode is %q", failover.FailoverMode)

	if failover.IsDistributedMode() {
		log.Printf("[DEBUG] failover: Failover mode is %q. Using predefined number of instances", failover.FailoverMode)
		failover.SetCounts(failover.Instances...)
		id, err := failover.ID()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		return resourcePolkadotFailoverRead(ctx, d, meta)
	}

	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return diag.FromErr(err)
	}

	metricsClient := config.NewMetricsClient(userAgent)

	if metricsClient == nil {
		return diag.Errorf("cannot initialize metric client")
	}

	computeClient := config.NewComputeClient(userAgent)

	if computeClient == nil {
		return diag.Errorf("cannot initialize compute client")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	validator, err := gcp.GetValidatorWithClient(
		ctx,
		metricsClient,
		failover.Project,
		failover.Prefix,
		failover.MetricNameSpace,
		failover.MetricName,
		1,
	)

	if err != nil && !errors.As(err, &gcp.ValidatorError{}) {
		log.Printf("[ERROR] failover: Cannot get validator: %s", err)
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] failover: Found validator instance: %s", validator.InstanceName)

	instanceGroups, err := gcp.GetInstanceGroupManagersForRegions(
		ctx,
		computeClient,
		failover.Project,
		failover.Prefix,
		failover.Locations...,
	)

	if err != nil {
		log.Printf("[ERROR] failover: Cannot get management instance groups: %s", err)
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] failover: Found %d managent instance groups", len(instanceGroups))

	positions := make([]int, len(failover.Locations))

	for i := 0; i < len(instanceGroups); i++ {
		group := &instanceGroups[i]
		if validatorInstance := group.SearchAndRemoveInstanceByName(validator.InstanceName); validatorInstance != nil {
			log.Printf("[DEBUG] failover: Processing validator instance: %s", validatorInstance.Instance)
			regionPosition := helpers.FindStrIndex(group.Region, failover.Locations)
			if regionPosition == -1 {
				log.Printf("[ERROR] failover: Cannot find region %s in locations list: %s", group.Region, strings.Join(failover.Locations, ", "))
				continue
			}
			positions[regionPosition] = 1
			break
		}
	}

	failover.SetCounts(positions...)

	// delete all instances besides the validator instance. In case we did not find the validator, or we found multiple validators,
	// we will delete all instances
	log.Printf(
		"[DEBUG] failover: Deleting %d managent instances: %q",
		instanceGroups.InstancesCount(),
		strings.Join(instanceGroups.InstanceNames(), ", "),
	)
	err = gcp.DeleteManagementInstances(ctx, computeClient, failover.Project, instanceGroups)
	if err != nil {
		return diag.FromErr(err)
	}

	err = gcp.WaitForInstancesCount(
		ctx,
		computeClient,
		failover.Project,
		failover.Prefix,
		failover.InstancesCount(),
		failover.Locations...,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	failover.FillDefaultCountsIfNotSet()
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
