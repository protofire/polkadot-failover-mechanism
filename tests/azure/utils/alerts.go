package utils

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/alertsmanagement/mgmt/2019-03-01/alertsmanagement"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-03-01/insights"
	"github.com/protofire/polkadot-failover-mechanism/tests/helpers"
)

// nolint
func getAlertsClient(subscriptionID, resourceGroup string) (alertsmanagement.AlertsClient, error) {

	//scope := fmt.Sprintf("subscriptions/%s/resourceGroups/%s", subscriptionID, resourceGroup)
	scope := fmt.Sprintf("subscriptions/%s", subscriptionID)

	client := alertsmanagement.NewAlertsClient(scope, subscriptionID, "")

	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil
}

// nolint
func getMetricAlertsClient(subscriptionID string) (insights.MetricAlertsClient, error) {
	client := insights.NewMetricAlertsClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil
}

// nolint
func getMetricsAlerts(subscriptionID, resourceGroup string) ([]insights.MetricAlertResource, error) {
	client, err := getMetricAlertsClient(subscriptionID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	rulesCollection, err := client.ListByResourceGroup(ctx, resourceGroup)

	if err != nil {
		return nil, err
	}

	rules := *rulesCollection.Value

	return rules, nil

}

// nolint
func getAlerts(subscriptionID, resourceGroup string) ([]alertsmanagement.Alert, error) {

	client, err := getAlertsClient(subscriptionID, resourceGroup)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	result, err := client.GetAll(
		ctx,
		"",
		"",
		resourceGroup,
		"",
		alertsmanagement.Fired,
		"",
		"",
		"",
		"",
		nil,
		nil,
		nil,
		"",
		"",
		"",
		alertsmanagement.Oneh,
		"",
	)

	if err != nil {
		return nil, err
	}

	values := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		values = append(values, result.Values()...)
	}

	return values, nil

}

// nolint
func filterMetricAlerts(rules *[]insights.MetricAlertResource, handler func(rule insights.MetricAlertResource) bool) {
	start := 0
	for i := start; i < len(*rules); i++ {
		if !handler((*rules)[i]) {
			// rules will be deleted
			continue
		}
		if i != start {
			(*rules)[start], (*rules)[i] = (*rules)[i], (*rules)[start]
		}
		start++
	}

	*rules = (*rules)[:start]

}

// nolint
func checkAlerts(alerts []alertsmanagement.Alert) error {
	if len(alerts) > 0 {
		for _, alert := range alerts {
			fmt.Printf("%#v\n", alert)
		}
		return fmt.Errorf("There are fired %d alerts", len(alerts))
	}
	return nil
}

func getTestMetricAlertNames(prefix string) []string {
	return []string{
		fmt.Sprintf("%s-disk-primary", prefix),
		fmt.Sprintf("%s-health-primary", prefix),
		fmt.Sprintf("%s-validator-primary", prefix),
		fmt.Sprintf("%s-validator-secondary", prefix),
		fmt.Sprintf("%s-health-secondary", prefix),
		fmt.Sprintf("%s-disk-secondary", prefix),
		fmt.Sprintf("%s-health-tertiary", prefix),
		fmt.Sprintf("%s-validator-tertiary", prefix),
		fmt.Sprintf("%s-disk-tertiary", prefix),
	}
}

// AlertsCheck checks all resource groups have been created
func AlertsCheck(prefix, subscriptionID, resourceGroup string) error {
	alerts, err := getMetricsAlerts(subscriptionID, resourceGroup)
	if err != nil {
		return err
	}

	filterMetricAlerts(&alerts, func(rule insights.MetricAlertResource) bool {
		return strings.HasPrefix(*rule.Name, helpers.GetPrefix(prefix))
	})

	testMetricAlertNames := getTestMetricAlertNames(prefix)

	var metricAlertNames []string

	for _, metric := range alerts {
		metricAlertNames = append(metricAlertNames, *metric.Name)
	}

	sort.Strings(testMetricAlertNames)
	sort.Strings(metricAlertNames)

	if !reflect.DeepEqual(testMetricAlertNames, metricAlertNames) {
		return fmt.Errorf("Test metric alerts do not coincide with actual: %#v, %#v", testMetricAlertNames, metricAlertNames)
	}

	alertsSummary, err := getAlerts(subscriptionID, resourceGroup)

	if err != nil {
		return err
	}

	return checkAlerts(alertsSummary)

}
