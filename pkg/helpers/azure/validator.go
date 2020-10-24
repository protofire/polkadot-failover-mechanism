package azure

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"
)

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

	return FindValidator(metrics, aggregator, 1)

}
