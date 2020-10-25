package azure

import (
	"context"
	"fmt"
	"log"
	"strings"

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
		return Validator{}, fmt.Errorf("cannot find validators")
	case 1:
		return validators[0], nil
	default:
		return Validator{}, fmt.Errorf("found %d validators: %#v", len(validators), validators)
	}

}
