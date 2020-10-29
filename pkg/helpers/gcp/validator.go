package gcp

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
)

type Validator struct {
	InstanceName string
	Metric       int
}

func GetValidatorWithClient(
	ctx context.Context,
	client *monitoring.MetricClient,
	project,
	prefix string,
	checkValue int,
	instanceNames ...string,
) (Validator, error) {
	points, err := GetValidatorMetrics(ctx, client, project, prefix, instanceNames...)

	if err != nil {
		return Validator{}, err
	}

	var validators []Validator

	for instanceName, points := range points {
		if len(points) == 0 {
			continue
		}

		lastPoint := points[0]

		value := int(lastPoint.Value.GetDoubleValue())

		if value == checkValue {
			validators = append(validators, Validator{
				InstanceName: instanceName,
				Metric:       checkValue,
			})
			continue
		}

	}

	switch len(validators) {
	case 0:
		return Validator{}, fmt.Errorf("cannot find validators")
	case 1:
		return validators[0], nil
	default:
		return Validator{}, fmt.Errorf("found %d validators: %#v", len(validators), validators)
	}

}

// WaitForValidator waits while validator metrics is being appeared
func WaitForValidator(
	project,
	prefix string,
	checkValue int,
	timeout int,
	instanceNames ...string,
) (Validator, error) {

	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	timerChan := timer.C

	defer timer.Stop()

	client, err := getMetricsClient()

	if err != nil {
		return Validator{}, err
	}

	ctx := context.Background()

	for {
		select {
		case <-timerChan:
			return Validator{}, fmt.Errorf("timeout waiting for validator")
		default:
			validator, err := GetValidatorWithClient(ctx, client, project, prefix, checkValue, instanceNames...)
			if err == nil && validator.InstanceName != "" {
				return validator, nil
			}
			time.Sleep(5 * time.Second)
		}
	}
}
