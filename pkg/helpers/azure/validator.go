package azure

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"
)

type Validator struct {
	ScaleSetName string
	Hostname     string
	Metric       int
}

func GetCurrentValidator(
	ctx context.Context,
	client *insights.MetricsClient,
	vmScaleSetNames []string,
	resourceGroup,
	metricName,
	metricNameSpace string,
	aggregator insights.AggregationType,
) (Validator, error) {

	log.Printf("[DEBUG]. Getting metrics for vm scale sets: %s", strings.Join(vmScaleSetNames, ", "))

	metrics, err := GetValidatorMetricsForVMScaleSets(
		ctx,
		client,
		vmScaleSetNames,
		resourceGroup,
		metricName,
		metricNameSpace,
		aggregator,
	)

	if err != nil {
		return Validator{}, fmt.Errorf("[ERROR]. Cannot get metric %s for namespace %s: %w", metricName, metricNameSpace, err)
	}

	return findValidator(metrics, aggregator, 1)

}

func findValidator(metrics map[string]insights.Metric, aggregationType insights.AggregationType, checkValue int) (Validator, error) {

	var validators []Validator

	for vmScaleSetName, metric := range metrics {
		if metric.Timeseries != nil && len(*metric.Timeseries) > 0 {
			series := (*metric.Timeseries)[0]
			if series.Data != nil && len(*series.Data) > 0 {
				for _, data := range *series.Data {
					metadata := series.Metadatavalues
					hostname := ""
					if metadata != nil && len(*series.Metadatavalues) > 0 {
						for _, meta := range *series.Metadatavalues {
							if meta.Name != nil && meta.Value != nil && meta.Name.Value != nil && *meta.Name.Value == "host" {
								hostname = *meta.Value
							}
						}
					}
					if getDataAggregation(data, aggregationType, checkValue) == checkValue {
						validators = append(validators, Validator{
							ScaleSetName: vmScaleSetName,
							Hostname:     hostname,
							Metric:       checkValue,
						})
						break
					}
				}
			}
		}
	}

	switch len(validators) {
	case 0:
		return Validator{}, errors.NewValidatorError("cannot find validators", errors.ValidatorErrorNotFound)
	case 1:
		return validators[0], nil
	default:
		return Validator{}, errors.NewValidatorError(fmt.Sprintf("found %d validators: %#v", len(validators), validators), errors.ValidatorErrorMultiple)
	}

}

// WaitForValidator waits while validator metrics is being appeared
func WaitForValidator(
	ctx context.Context,
	client *insights.MetricsClient,
	vmScaleSetNames []string,
	resourceGroup,
	metricName,
	metricNamespace string,
	period int,
) (Validator, error) {

	ticker := time.NewTicker(time.Duration(period) * time.Second)
	tickerChan := ticker.C

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return Validator{}, fmt.Errorf("timeout waiting for validator. context has been cancelled")
		case <-tickerChan:
			validator, err := GetCurrentValidator(
				ctx,
				client,
				vmScaleSetNames,
				resourceGroup,
				metricName,
				metricNamespace,
				insights.Maximum,
			)
			if err == nil && validator.ScaleSetName != "" {
				return validator, err
			}
		}
	}
}
