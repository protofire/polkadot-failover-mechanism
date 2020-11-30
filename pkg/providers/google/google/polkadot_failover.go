package google

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
	helperErrors "github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/gcp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

		Schema: resource.GetPolkadotSchema(),
	}
}

func resourcePolkadotFailoverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Config)
	failover := &GCPFailover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		log.Printf("[DEBUG] failover: Read. Error reading failover state: %v", err)
		return diag.FromErr(err)
	}

	if failover.Project == "" {
		failover.Project, err = getProject(d, config)
		if err != nil {
			log.Printf("[DEBUG] failover: Read. Error getting google project: %v", err)
			return diag.FromErr(err)
		}
	}

	if !failover.Initialized() {
		log.Printf("[DEBUG] failover: Read. Non initialized state. Resetting resource ID")
		d.SetId("")
		return nil
	}

	if failover.IsDistributedMode() {
		log.Printf("[DEBUG] failover: Read. Failover mode is %q. Using predefined number of instances", failover.FailoverMode)
		failover.SetCounts(failover.Instances...)
		return failover.SetSchemaValuesDiag(d)
	}

	log.Printf("[DEBUG] failover: Read. Failover mode is %q", failover.FailoverMode)

	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return diag.FromErr(err)
	}
	computeClient := config.NewComputeClient(userAgent)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Printf("[DEBUG] failover: Read. Getting instances list...")

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

	log.Printf("[DEBUG] failover: Read. Found %d managent instance groups", len(instanceGroups))

	positions := make([]int, len(failover.Locations))

	for _, group := range instanceGroups {
		regionPosition := helpers.FindStrIndex(group.Region, failover.Locations)
		if regionPosition == -1 {
			log.Printf("[ERROR] failover: Read. Cannot find region %s in locations list: %s", group.Region, strings.Join(failover.Locations, ", "))
			continue
		}
		positions[regionPosition] = len(group.Instances)
	}

	log.Printf("[DEBUG] failover: Read. Found instance numbers per region: %v", positions)

	failover.SetCounts(positions...)

	failover.FillDefaultCountsIfNotSet()

	log.Printf("[DEBUG] failover: Read. Set instance numbers per region: %v", failover.FailoverInstances)

	return failover.SetSchemaValuesDiag(d)
}

func resourcePolkadotFailoverCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Config)
	failover := &GCPFailover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		log.Printf("[DEBUG] failover: Create. Error reading failover state: %v", err)
		return diag.FromErr(err)
	}

	if failover.Project == "" {
		failover.Project, err = getProject(d, config)
		if err != nil {
			log.Printf("[DEBUG] failover: Create. Error getting google project: %v", err)
			return diag.FromErr(err)
		}
	}

	if failover.IsDistributedMode() {
		log.Printf("[DEBUG] failover: Create. Failover mode is %q. Using predefined number of instances", failover.FailoverMode)
		failover.SetCounts(failover.Instances...)
		id, err := failover.ID()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(id)
		return resourcePolkadotFailoverRead(ctx, d, meta)
	}

	log.Printf("[DEBUG] failover: Create. Failover mode is %q", failover.FailoverMode)

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

	if err != nil {
		validatorError := &helperErrors.ValidatorError{}
		if errors.As(err, validatorError) {
			log.Printf("[WARNING] failover: Create. Cannot get validator: %s", validatorError)
		} else {
			log.Printf("[ERROR] failover: Create. Cannot get validator: %s", err)
			return diag.FromErr(err)
		}
	}

	if validator.InstanceName != "" {
		log.Printf("[DEBUG] failover: Create. Found validator instance: %s", validator.InstanceName)
	} else {
		log.Printf("[DEBUG] failover: Create. Have not found the validator instance")
	}

	instanceGroups, err := gcp.GetInstanceGroupManagersForRegions(
		ctx,
		computeClient,
		failover.Project,
		failover.Prefix,
		failover.Locations...,
	)

	if err != nil {
		log.Printf("[ERROR] failover: Create. Cannot get management instance groups: %s", err)
		return diag.FromErr(err)
	}

	log.Printf(
		"[DEBUG] failover: Create. Found %d managent instance groups with %d instances",
		len(instanceGroups),
		instanceGroups.InstancesCount(),
	)

	initialInstancesCount := instanceGroups.InstancesCount()
	positions := make([]int, len(failover.Locations))

	for i := 0; i < len(instanceGroups); i++ {
		group := &instanceGroups[i]
		if validatorInstance := group.SearchAndRemoveInstanceByName(validator.InstanceName); validatorInstance != nil {
			log.Printf("[DEBUG] failover: Create. Processing validator instance: %s", validatorInstance.Instance)
			regionPosition := helpers.FindStrIndex(group.Region, failover.Locations)
			if regionPosition == -1 {
				log.Printf("[ERROR] failover: Create. Cannot find region %s in locations list: %s", group.Region, strings.Join(failover.Locations, ", "))
				continue
			}
			positions[regionPosition] = 1
			break
		}
	}

	failover.SetCounts(positions...)
	failover.FillDefaultCountsIfNotSet()

	// delete all instances besides the validator instance. In case we did not find the validator, or we found multiple validators,
	// we will delete all instances
	log.Printf(
		"[DEBUG] failover: Create. Deleting %d managent instances: %q",
		instanceGroups.InstancesCount(),
		strings.Join(instanceGroups.InstanceNames(), ", "),
	)
	err = gcp.DeleteManagementInstances(ctx, computeClient, failover.Project, instanceGroups)
	if err != nil {
		return diag.FromErr(err)
	}

	if initialInstancesCount > 0 {
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
	}

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
