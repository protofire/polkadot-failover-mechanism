package gcp

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
)

const (
	resourceType    = "gce_instance"
	metricNamespace = "polkadot"
	metricName      = "validator/value"
)

type ValidatorErrorType int

func (vt ValidatorErrorType) String() string {
	switch vt {
	case ValidatorErrorUnknown:
		return "ValidatorErrorUnknown"
	case ValidatorErrorNotFound:
		return "ValidatorErrorNotFound"
	case ValidatorErrorMultiple:
		return "ValidatorErrorMultiple"
	}
	return ""
}

const (
	ValidatorErrorUnknown ValidatorErrorType = iota
	ValidatorErrorNotFound
	ValidatorErrorMultiple
)

type Validator struct {
	InstanceName string
	Metric       int
}

type ValidatorError struct {
	Message string
	Kind    ValidatorErrorType
}

func NewValidatorError(message string, kind ValidatorErrorType) ValidatorError {
	return ValidatorError{
		Message: message,
		Kind:    kind,
	}
}

func (v ValidatorError) Error() string {
	return fmt.Sprintf("%s: Error type: %s", v.Message, v.Kind)
}

func (v ValidatorError) IsNotFound() bool {
	return v.Kind == ValidatorErrorNotFound
}

func (v ValidatorError) MultipleValidators() bool {
	return v.Kind == ValidatorErrorMultiple
}

func GetValidatorWithClient(
	ctx context.Context,
	client *monitoring.MetricClient,
	project,
	prefix,
	metricNamespace,
	metricName string,
	checkValue int,
	instanceNames ...string,
) (Validator, error) {
	points, err := GetValidatorMetrics(ctx, client, project, prefix, metricNamespace, metricName, instanceNames...)

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
		return Validator{}, NewValidatorError("cannot find validators", ValidatorErrorNotFound)
	case 1:
		return validators[0], nil
	default:
		return Validator{}, NewValidatorError(fmt.Sprintf("found %d validators: %#v", len(validators), validators), ValidatorErrorMultiple)
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
			validator, err := GetValidatorWithClient(ctx, client, project, prefix, metricNamespace, metricName, checkValue, instanceNames...)
			if err == nil && validator.InstanceName != "" {
				return validator, nil
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// WaitForValidatorWithClient waits while validator metrics is being appeared
func WaitForValidatorWithClient(
	ctx context.Context,
	client *monitoring.MetricClient,
	project,
	prefix string,
	checkValue int,
	timeout int,
	instanceNames ...string,
) (Validator, error) {

	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	timerChan := timer.C

	defer timer.Stop()

	for {
		select {
		case <-timerChan:
			return Validator{}, fmt.Errorf("timeout waiting for validator")
		default:
			validator, err := GetValidatorWithClient(ctx, client, project, prefix, metricNamespace, metricName, checkValue, instanceNames...)
			if err == nil && validator.InstanceName != "" {
				return validator, nil
			}
			time.Sleep(5 * time.Second)
		}
	}
}
