package aws

// This file contains all the supplementary functions that are required to query Cloud Watch (AWS)

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	helperErrors "github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type Validator struct {
	Value int
	AsgInstancePair
}

func getValidatorMetric(ctx context.Context, client *cloudwatch.CloudWatch, asgName, instanceID, metricNamespace, metricName string) (int, error) {
	endTime := time.Now()
	duration, _ := time.ParseDuration("-5m")
	startTime := endTime.Add(duration)
	period := int64(300)
	metricID := "m1"
	stat := "Maximum"
	metricDim1Name := "asg_name"
	metricDim2Name := "instance_id"
	query := &cloudwatch.MetricDataQuery{
		Id: &metricID,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Namespace:  &metricNamespace,
				MetricName: &metricName,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  &metricDim1Name,
						Value: &asgName,
					},
					{
						Name:  &metricDim2Name,
						Value: &instanceID,
					},
				},
			},
			Period: &period,
			Stat:   &stat,
		},
	}
	resp, err := client.GetMetricDataWithContext(ctx, &cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		StartTime:         &startTime,
		MetricDataQueries: []*cloudwatch.MetricDataQuery{query},
	})
	if err != nil {
		return 0, fmt.Errorf("cannot get metrics %q for asg %q. Region %q: %w", metricName, asgName, client.SigningRegion, err)
	}

	if len(resp.Messages) == 0 {
		return 0, nil
	}

	message := resp.Messages[len(resp.Messages)-1]

	if message.Value != nil {
		value, err := strconv.Atoi(*message.Value)
		if err != nil {
			return 0, fmt.Errorf(
				"cannot get metrics %q for asg %q. Region %q. cannot parse value %v: %w",
				metricName,
				asgName,
				client.SigningRegion,
				*message.Value,
				err,
			)
		}
		return value, nil
	}

	return 0, nil

}

func getValidatorMetrics(
	ctx context.Context,
	clients []*cloudwatch.CloudWatch,
	asgs AgsGroupsList,
	metricNamespace, metricName string,
) ([]Validator, error) {

	var pairs []interface{}

	for _, pair := range asgs.AsgInstancePairs() {
		pairs = append(pairs, pair)
	}

	out := fanout.ConcurrentResponseItems(ctx, func(ctx context.Context, value interface{}) (interface{}, error) {
		pair := value.(AsgInstancePair)
		metric, err := getValidatorMetric(
			ctx,
			clients[pair.RegionID],
			pair.ASGName,
			pair.InstanceID,
			metricNamespace,
			metricName,
		)

		if err != nil {
			return Validator{}, err
		}

		return Validator{
			AsgInstancePair: pair,
			Value:           metric,
		}, nil

	}, pairs...)

	items, err := fanout.ReadItemChannel(out)

	result := make([]Validator, len(pairs))

	if err != nil {
		return result, err
	}

	for _, item := range items {
		result = append(result, item.(Validator))
	}

	return result, nil
}

func GetValidator(
	ctx context.Context,
	clients []*cloudwatch.CloudWatch,
	asgs []AgsGroups,
	metricNamespace,
	metricName string,
) (Validator, error) {

	metricItems, err := getValidatorMetrics(ctx, clients, asgs, metricNamespace, metricName)

	if err != nil {
		return Validator{}, err
	}

	var validators []Validator

	for _, metric := range metricItems {
		if metric.Value != 0 {
			validators = append(validators, metric)
		}
	}

	switch len(validators) {
	case 0:
		return Validator{}, helperErrors.NewValidatorError("cannot find validators", helperErrors.ValidatorErrorNotFound)
	case 1:
		return validators[0], nil
	default:
		return Validator{}, helperErrors.NewValidatorError(
			fmt.Sprintf("found %d validators: %#v", len(validators), validators),
			helperErrors.ValidatorErrorMultiple,
		)
	}

}

func WaitForValidator(
	ctx context.Context,
	clients []*cloudwatch.CloudWatch,
	asgs AgsGroupsList,
	metricNamespace,
	metricName string,
	period int,
) (Validator, error) {

	ticker := time.NewTicker(time.Duration(period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return Validator{}, fmt.Errorf("timeout waiting for validator")
		case <-ticker.C:
			validator, err := GetValidator(ctx, clients, asgs, metricNamespace, metricName)
			if err != nil {
				validatorError := &helperErrors.ValidatorError{}
				if errors.As(err, validatorError) {
					if validatorError.MultipleValidators() {
						log.Printf("[ERROR] failover: %v", err)
					}
				} else {
					log.Printf("[ERROR] failover: Unexpected error getting validator: %v", err)
				}
				continue
			}
			return validator, nil
		}
	}

}
