package aws

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/aws"

	helperErrors "github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

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

	failover := &Failover{}
	err := failover.FromIDOrSchema(d)

	if err != nil {
		return diag.FromErr(err)
	}

	if !failover.Initialized() {
		d.SetId("")
		return nil
	}

	if failover.IsDistributedMode() {
		log.Printf("[DEBUG] failover: Failover mode is %q. Using predefined number of instances", failover.FailoverMode)
		failover.SetCounts(failover.Instances...)
		return failover.SetSchemaValuesDiag(d)
	}

	log.Printf("[DEBUG] failover: Failover mode is %q", failover.FailoverMode)

	awsClients := meta.([]*Client)
	ec2Clients := make([]*ec2.EC2, len(awsClients))
	cloudWatchClients := make([]*cloudwatch.CloudWatch, len(awsClients))
	autoscalingClients := make([]*autoscaling.AutoScaling, len(awsClients))

	for idx, client := range awsClients {
		ec2Clients[idx] = client.ec2conn
		cloudWatchClients[idx] = client.cloudwatchconn
		autoscalingClients[idx] = client.autoscalingconn
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	log.Printf("[DEBUG] failover: Getting instances list...")

	log.Printf("[DEBUG] failover: Getting ags groups...")
	asgsGroupsList, err := aws.GetASGs(ctx, autoscalingClients, failover.Prefix)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] failover: Found %d instance groups", asgsGroupsList.GroupsCount())

	positions := asgsGroupsList.InstancesCountPerRegion()

	log.Printf("[DEBUG] failover: Found instance numbers per region: %v", positions)

	failover.SetCounts(positions...)

	failover.FillDefaultCountsIfNotSet()

	log.Printf("[DEBUG] failover: Set instance numbers per region: %v", failover.FailoverInstances)

	return failover.SetSchemaValuesDiag(d)
}

func resourcePolkadotFailoverCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	failover := &Failover{}
	err := failover.FromSchema(d)

	if err != nil {
		return diag.FromErr(err)
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

	awsClients := meta.([]*Client)
	ec2Clients := make([]*ec2.EC2, len(awsClients))
	cloudWatchClients := make([]*cloudwatch.CloudWatch, len(awsClients))
	autoscalingClients := make([]*autoscaling.AutoScaling, len(awsClients))

	for _, client := range awsClients {
		ec2Clients = append(ec2Clients, client.ec2conn)
		cloudWatchClients = append(cloudWatchClients, client.cloudwatchconn)
		autoscalingClients = append(autoscalingClients, client.autoscalingconn)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	log.Printf("[DEBUG] failover: Getting ags groups...")
	asgsGroupsList, err := aws.GetASGs(ctx, autoscalingClients, failover.Prefix)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] failover: Getting failover validator...")
	validator, err := aws.GetValidator(ctx, cloudWatchClients, asgsGroupsList, failover.MetricNameSpace, failover.MetricName)

	if err != nil {
		validatorError := &helperErrors.ValidatorError{}
		if errors.As(err, validatorError) {
			log.Printf("[WARNING] failover: Cannot get validator: %s", validatorError)
		} else {
			log.Printf("[ERROR] failover: Cannot get validator: %s", err)
			return diag.FromErr(err)
		}
	}

	if validator.InstanceID != "" {
		log.Printf("[DEBUG] failover: Found the validator instance %q in auto scale group %q", validator.InstanceID, validator.ASGName)
	} else {
		log.Printf("[DEBUG] failover: Have not found the validator instance")
	}

	positions := make([]int, len(awsClients))

	if validator.InstanceID != "" {
		positions[validator.RegionID] = 1
	}

	failover.SetCounts(positions...)
	failover.FillDefaultCountsIfNotSet()

	// delete all instances besides the validator instance. In case we did not find the validator, or we found multiple validators,
	// we will delete all instances

	instancesToDelete := make(aws.AsgToInstancesByRegion, len(awsClients))

	for regionID, groups := range asgsGroupsList {
		for _, group := range groups {
			for _, instance := range group.Instances {
				asgToInstances := &instancesToDelete[regionID]
				if *instance.InstanceId != validator.InstanceID {
					(*asgToInstances)[*group.AutoScalingGroupName] = append((*asgToInstances)[*group.AutoScalingGroupName], *instance.InstanceId)
				}
			}
		}
	}

	log.Printf(
		"[DEBUG] failover: Deleting %d asg instances: %q",
		instancesToDelete.InstancesCount(),
		strings.Join(instancesToDelete.InstancesIDs(), ", "),
	)
	for regionID, mp := range instancesToDelete {
		for asgName, instances := range mp {
			if regionID < len(autoscalingClients) {
				err := aws.DetachASGInstances(ctx, autoscalingClients[regionID], asgName, instances)
				if err != nil {
					return diag.FromErr(err)
				}
				err = aws.DeleteInstances(ctx, ec2Clients[regionID], instances)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	if validator.InstanceID != "" {
		log.Printf("[DEBUG] failover: waiting for validator...")
		_, err := aws.WaitForValidator(ctx, cloudWatchClients, asgsGroupsList, failover.MetricNameSpace, failover.MetricName, 5)
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
