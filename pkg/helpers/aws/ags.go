package aws

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type AgsGroups []*autoscaling.Group
type AgsGroupsList []AgsGroups
type AsgToInstancesByRegion []map[string][]string

type AsgInstancePair struct {
	InstanceID string
	ASGName    string
	RegionID   int
}

func (a *AsgToInstancesByRegion) InstancesCount() int {
	s := 0
	for _, mp := range *a {
		for _, values := range mp {
			s += len(values)
		}
	}
	return s
}

func (a *AsgToInstancesByRegion) InstancesIDs() []string {
	var ids []string
	for _, mp := range *a {
		for _, values := range mp {
			ids = append(ids, values...)
		}
	}
	return ids
}

func (a AgsGroups) AsgInstancePair(regionID int) []AsgInstancePair {
	var result []AsgInstancePair
	for _, group := range a {
		for _, instance := range group.Instances {
			result = append(result, AsgInstancePair{
				InstanceID: *instance.InstanceId,
				ASGName:    *group.AutoScalingGroupName,
				RegionID:   regionID,
			})
		}
	}
	return result
}

func (a AgsGroupsList) AsgInstancePairs() []AsgInstancePair {
	var result []AsgInstancePair
	for regionID, groups := range a {
		result = append(result, groups.AsgInstancePair(regionID)...)
	}
	return result
}

func (a AgsGroupsList) GroupsCount() int {
	s := 0
	for _, groups := range a {
		s += len(groups)
	}
	return s
}

func (a AgsGroupsList) InstancesCount() int {
	s := 0
	for _, groups := range a {
		for _, group := range groups {
			s += len(group.Instances)
		}
	}
	return s
}

func (a AgsGroupsList) InstancesCountPerRegion() []int {
	instances := make([]int, len(a))
	for regionID, groups := range a {
		for _, group := range groups {
			instances[regionID] += len(group.Instances)
		}
	}
	return instances
}

func processAwsError(err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr
	}
	return err
}

func filterASG(groups *AgsGroups, handler func(group *autoscaling.Group) bool) {

	start := 0
	for i := start; i < len(*groups); i++ {
		if !handler((*groups)[i]) {
			// vm will be deleted
			continue
		}
		if i != start {
			(*groups)[start], (*groups)[i] = (*groups)[i], (*groups)[start]
		}
		start++
	}

	*groups = (*groups)[:start]

}

func GetRegionASGs(ctx context.Context, client *autoscaling.AutoScaling, prefix string) (AgsGroups, error) {

	var groups AgsGroups

	resp, err := client.DescribeAutoScalingGroupsWithContext(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: nil,
		MaxRecords:            nil,
		NextToken:             nil,
	})
	if err != nil {
		return nil, processAwsError(err)
	}
	if resp.AutoScalingGroups != nil {
		groups = append(groups, resp.AutoScalingGroups...)
	}
	for resp.NextToken != nil && *resp.NextToken != "" {
		resp, err := client.DescribeAutoScalingGroupsWithContext(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: nil,
			MaxRecords:            nil,
			NextToken:             resp.NextToken,
		})
		if err != nil {
			return nil, processAwsError(err)
		}
		if resp.AutoScalingGroups != nil {
			groups = append(groups, resp.AutoScalingGroups...)
		}
	}

	filterASG(&groups, func(group *autoscaling.Group) bool {
		return strings.HasPrefix(*group.AutoScalingGroupName, prefix)
	})

	return groups, nil

}

func GetASGs(ctx context.Context, clients []*autoscaling.AutoScaling, prefix string) (AgsGroupsList, error) {

	type asgItem struct {
		groups []*autoscaling.Group
		index  int
	}

	var indexes []interface{}

	for index := range clients {
		indexes = append(indexes, index)
	}

	out := fanout.ConcurrentResponseItems(ctx, func(ctx context.Context, value interface{}) (interface{}, error) {
		index := value.(int)
		client := clients[index]
		groups, err := GetRegionASGs(
			ctx,
			client,
			prefix,
		)

		if err != nil {
			return asgItem{}, err
		}

		return asgItem{
			groups: groups,
			index:  index,
		}, nil

	}, indexes...)

	var groups []AgsGroups

	items, err := fanout.ReadItemChannel(out)

	if err != nil {
		return groups, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].(asgItem).index < items[j].(asgItem).index
	})

	for _, item := range items {
		groups = append(groups, item.(asgItem).groups)
	}

	return groups, nil
}

func checkActionActivities(activities []*autoscaling.Activity) ([]string, error) {
	var actionActivities []string
	for _, activity := range activities {
		switch *activity.StatusCode {
		case autoscaling.ScalingActivityStatusCodeFailed:
			return nil, fmt.Errorf("auto scale activity in state %s", autoscaling.ScalingActivityStatusCodeFailed)
		case autoscaling.ScalingActivityStatusCodeSuccessful, autoscaling.InstanceRefreshStatusCancelled:
			continue
		default:
			log.Printf(
				"[DEBUG] failover: Got unfinished activity with id %s, status %s, autoscale group name %q",
				*activity.ActivityId,
				*activity.StatusCode,
				*activity.AutoScalingGroupName,
			)
			actionActivities = append(actionActivities, *activity.ActivityId)
		}
	}
	return actionActivities, nil
}

func DetachASGInstances(ctx context.Context, client *autoscaling.AutoScaling, asgName string, instanceIDs []string) error {

	resp, err := client.DetachInstancesWithContext(ctx, &autoscaling.DetachInstancesInput{
		AutoScalingGroupName:           &asgName,
		InstanceIds:                    aws.StringSlice(instanceIDs),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	})

	if err != nil {
		return err
	}

	activities, err := checkActionActivities(resp.Activities)

	if err != nil {
		return err
	}

	for len(activities) > 0 {

		log.Printf("[DEBUG]: failover. Got %d unfinished activities: %s", len(activities), activities)

		resp, err := client.DescribeScalingActivitiesWithContext(ctx, &autoscaling.DescribeScalingActivitiesInput{
			ActivityIds:          aws.StringSlice(activities),
			AutoScalingGroupName: aws.String(asgName),
		})

		if err != nil {
			return err
		}

		newActivities, err := checkActionActivities(resp.Activities)

		if err != nil {
			return err
		}

		for resp.NextToken != nil {
			resp, err := client.DescribeScalingActivitiesWithContext(ctx, &autoscaling.DescribeScalingActivitiesInput{
				ActivityIds:          aws.StringSlice(activities),
				AutoScalingGroupName: aws.String(asgName),
				NextToken:            resp.NextToken,
			})

			if err != nil {
				return err
			}

			tokenActivities, err := checkActionActivities(resp.Activities)

			if err != nil {
				return err
			}

			newActivities = append(newActivities, tokenActivities...)
		}

		activities = newActivities

	}

	return nil
}

func DeleteInstances(ctx context.Context, client *ec2.EC2, instanceIDs []string) error {
	_, err := client.TerminateInstancesWithContext(ctx, &ec2.TerminateInstancesInput{
		DryRun:      aws.Bool(false),
		InstanceIds: aws.StringSlice(instanceIDs),
	})
	if err != nil {
		return err
	}
	return client.WaitUntilInstanceExistsWithContext(ctx, &ec2.DescribeInstancesInput{
		DryRun:      aws.Bool(false),
		InstanceIds: aws.StringSlice(instanceIDs),
	})
}
