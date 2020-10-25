package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/golang/protobuf/ptypes/timestamp"

	"google.golang.org/api/iterator"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
)

type InstanceMetricPoints map[string][]*monitoringpb.Point

// nolint
func getMetricsClient() (*monitoring.MetricClient, error) {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func prepareFilter(prefix, resourceType, metricsNamespace, metricsFamily, metricName string, instanceNames ...string) string {

	var instanceFilters []string

	for _, name := range instanceNames {
		instanceFilters = append(instanceFilters, fmt.Sprintf(`resource.label.instance_id = "%s"`, name))
	}

	instancesFilter := strings.Join(instanceFilters, " OR ")

	mainFilter := fmt.Sprintf(
		`resource.type = "%s" AND metric.type = "custom.googleapis.com/%s/%s/%s" AND metric.label.prefix = "%s"`,
		resourceType,
		metricsNamespace,
		metricsFamily,
		metricName,
		prefix,
	)

	if instancesFilter != "" {
		return fmt.Sprintf("%s AND (%s)", mainFilter, instancesFilter)
	}
	return mainFilter

}

func listMetrics(
	ctx context.Context,
	client *monitoring.MetricClient,
	project,
	prefix,
	resourceType,
	metricsNamespace,
	metricsFamily,
	metricName string,
	interval int,
	alignmentPeriod int,
	instanceNames ...string,
) (InstanceMetricPoints, error) {

	filter := prepareFilter(prefix, resourceType, metricsNamespace, metricsFamily, metricName, instanceNames...)
	startTime := time.Now().UTC().Add(time.Minute * -time.Duration(interval))
	endTime := time.Now().UTC()

	timeInterval := &monitoringpb.TimeInterval{
		StartTime: &timestamp.Timestamp{
			Seconds: startTime.Unix(),
		},
		EndTime: &timestamp.Timestamp{
			Seconds: endTime.Unix(),
		},
	}

	aggregation := &monitoringpb.Aggregation{
		AlignmentPeriod:    &durationpb.Duration{Seconds: int64(alignmentPeriod)},
		PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_MAX,
		CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
		GroupByFields:      []string{"resource.label.instance_id"},
	}

	req := monitoringpb.ListTimeSeriesRequest{
		Name:        "projects/" + project,
		Filter:      filter,
		Interval:    timeInterval,
		Aggregation: aggregation,
		View:        monitoringpb.ListTimeSeriesRequest_FULL,
	}

	timeSeriesIterator := client.ListTimeSeries(ctx, &req)
	result := make(InstanceMetricPoints)

	for {
		timeSeries, err := timeSeriesIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return result, err
		}

		metric := timeSeries.Metric
		resource := timeSeries.Resource
		instanceID := resource.Labels["instance_id"]
		projectID := resource.Labels["project_id"]
		metricPrefix := metric.Labels["prefix"]

		if projectID != project || metricPrefix != prefix {
			continue
		}
		result[instanceID] = timeSeries.Points

	}

	return result, nil

}

func GetValidatorMetrics(
	ctx context.Context,
	client *monitoring.MetricClient,
	project,
	prefix string,
	instanceNames ...string,
) (InstanceMetricPoints, error) {
	return listMetrics(
		ctx,
		client,
		project,
		prefix,
		"gce_instance",
		"polkadot",
		"validator",
		"value",
		1,
		60,
		instanceNames...,
	)
}
