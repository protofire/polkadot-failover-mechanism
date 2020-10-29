package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"
)

func getMetricsResourceURL(subscriptionID, resourceGroup, vmScaleSetName string) string {
	return fmt.Sprintf(
		"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachineScaleSets/%s",

		subscriptionID,
		resourceGroup,
		vmScaleSetName,
	)
}

// GetValidatorMetricsForVMScaleSet ...
func GetValidatorMetricsForVMScaleSet(
	ctx context.Context,
	client *insights.MetricsClient,
	resourceGroup,
	vmScaleSetName,
	metricsName,
	metricNameSpace string,
	aggregationType insights.AggregationType,
) (insights.Metric, error) {

	interval := "PT1M"
	timespan := fmt.Sprintf(
		"%s/%s",
		time.Now().UTC().Add(time.Duration(-5)*time.Minute).Format("2006-01-02T15:04:05"),
		time.Now().UTC().Format("2006-01-02T15:04:05"),
	)
	resourceURI := getMetricsResourceURL(client.SubscriptionID, resourceGroup, vmScaleSetName)
	result, err := client.List(
		ctx,
		resourceURI,
		timespan,
		&interval,
		metricsName,
		string(aggregationType),
		nil,
		string(aggregationType),
		"host eq '*'",
		"",
		metricNameSpace,
	)

	if err != nil {
		return insights.Metric{}, err
	}

	if result.Value == nil || len(*result.Value) == 0 {
		return insights.Metric{}, nil
	}

	return (*result.Value)[len(*result.Value)-1], nil

}

func GetMetricsClient(subscriptionID string) (insights.MetricsClient, error) {
	client := insights.NewMetricsClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("cannot get authorizer: %w", err)
	}
	client.Authorizer = auth
	return client, nil
}

// GetValidatorMetricsForVMScaleSets ...
func GetValidatorMetricsForVMScaleSets(
	ctx context.Context,
	client *insights.MetricsClient,
	vmScaleSetNames []string,
	resourceGroup,
	metricsName,
	metricNameSpace string,
	aggregationType insights.AggregationType,
) (map[string]insights.Metric, error) {

	result := make(map[string]insights.Metric, len(vmScaleSetNames))

	type metricItem struct {
		metric         insights.Metric
		vmScaleSetName string
	}

	var names []interface{}

	for _, name := range vmScaleSetNames {
		names = append(names, name)
	}

	out := fanout.ConcurrentResponseItems(ctx, func(ctx context.Context, value interface{}) (interface{}, error) {
		vmScaleSetName := value.(string)
		metric, err := GetValidatorMetricsForVMScaleSet(
			ctx,
			client,
			resourceGroup,
			vmScaleSetName,
			metricsName,
			metricNameSpace,
			aggregationType,
		)

		if err != nil {
			return metricItem{}, err
		}

		return metricItem{
			metric:         metric,
			vmScaleSetName: vmScaleSetName,
		}, nil

	}, names...)

	items, err := fanout.ReadItemChannel(out)

	if err != nil {
		return result, err
	}

	for _, item := range items {
		mi := item.(metricItem)
		result[mi.vmScaleSetName] = mi.metric
	}

	return result, err

}

func getDataAggregation(data insights.MetricValue, aggregationType insights.AggregationType, checkValue int) int {

	switch aggregationType {
	case insights.Maximum:
		if data.Maximum != nil && int(*data.Maximum) == checkValue {
			return checkValue
		}
	case insights.Minimum:
		if data.Minimum != nil && int(*data.Minimum) == checkValue {
			return checkValue
		}
	case insights.Average:
		if data.Average != nil && int(*data.Average) == checkValue {
			return checkValue
		}
	case insights.Count:
		if data.Count != nil && int(*data.Count) == checkValue {
			return checkValue
		}
	case insights.Total:
		if data.Total != nil && int(*data.Total) == checkValue {
			return checkValue
		}
	}
	return -1
}

// LogMetrics ...
func LogMetrics(metrics map[string]insights.Metric, level string) {
	for vmScaleSetName, metric := range metrics {
		body, err := json.MarshalIndent(metric, "", "    ")
		if err == nil {
			log.Printf("[%s]. Got metrics for vm scale set %s:\n%s", level, vmScaleSetName, string(body))
		} else {
			log.Printf("[%s]. Got metrics for vm scale set %s - %#v: %#v", level, vmScaleSetName, metric, err)
		}
	}
}
