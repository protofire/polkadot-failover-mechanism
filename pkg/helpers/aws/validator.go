package aws

// This file contains all the supplementary functions that are required to query Cloud Watch (AWS)

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	taws "github.com/gruntwork-io/terratest/modules/aws"

	helperErrors "github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type Validator struct {
	Value int
	AsgInstancePair
}

func getValidatorMetric(ctx context.Context, client *cloudwatch.CloudWatch, asgName, metricNamespace, metricName string) (int, error) {
	endTime := time.Now()
	duration, _ := time.ParseDuration("-5m")
	startTime := endTime.Add(duration)
	period := int64(300)
	metricID := "m1"
	stat := "Maximum"
	metricDim1Name := "group_name"
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

	if len(resp.MetricDataResults) == 0 {
		log.Printf(
			"[DEBUG] failover: Not found metric data messages for ASG %q, metric namespace %q, metric name %q",
			asgName,
			metricNamespace,
			metricName,
		)
		return 0, nil
	}

	result := resp.MetricDataResults[len(resp.MetricDataResults)-1]
	values := result.Values

	var intValues []int
	for _, value := range values {
		if value != nil {
			intValues = append(intValues, int(*value))
		}
	}

	if len(intValues) == 0 {
		log.Printf(
			"[DEBUG] failover: Not found metric data messages for ASG %q, metric namespace %q, metric name %q",
			asgName,
			metricNamespace,
			metricName,
		)
		return 0, nil
	}

	log.Printf("[DEBUG] failover: Got metric data values: %d", intValues)

	return intValues[len(intValues)-1], nil

}

func getValidatorMetrics(
	ctx context.Context,
	clients []*cloudwatch.CloudWatch,
	asgs AgsGroupsList,
	metricNamespace,
	metricName string,
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
	asgs AgsGroupsList,
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

func newCloudWatchClient(region string) (*cloudwatch.CloudWatch, error) {
	sess, err := taws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}
	return cloudwatch.New(sess), nil
}

func WaitForValidatorRegions(
	regions []string,
	metricNamespace,
	metricName,
	prefix string,
	timeout int,
	period int,
) (Validator, error) {

	var cloudWatchClients []*cloudwatch.CloudWatch
	var asgClients []*autoscaling.AutoScaling

	for _, region := range regions {
		cwClient, err := newCloudWatchClient(region)
		if err != nil {
			return Validator{}, err
		}
		asgClient, err := newAsgClient(region)
		if err != nil {
			return Validator{}, err
		}
		cloudWatchClients = append(cloudWatchClients, cwClient)
		asgClients = append(asgClients, asgClient)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(timeout))
	defer cancel()

	asgGroupsList, err := GetASGs(ctx, asgClients, prefix)
	if err != nil {
		return Validator{}, err
	}

	return WaitForValidator(ctx, cloudWatchClients, asgGroupsList, metricNamespace, metricName, period)

}
