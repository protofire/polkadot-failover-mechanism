package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/iterator"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/golang/protobuf/ptypes/timestamp"

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

func prepareFilter(prefix, project, resourceType, metricsNamespace, metricName string, instanceNames ...string) string {

	var instanceFilters []string

	for _, name := range instanceNames {
		instanceFilters = append(instanceFilters, fmt.Sprintf(`resource.label.instance_id = "%s"`, name))
	}

	instancesFilter := strings.Join(instanceFilters, " OR ")

	mainFilter := fmt.Sprintf(
		`resource.type = "%s" AND resource.label.project_id = "%s" AND metric.type = "custom.googleapis.com/%s/%s" AND metric.label.prefix = "%s"`,
		resourceType,
		project,
		metricsNamespace,
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
	metricName string,
	interval int,
	alignmentPeriod int,
	instanceNames ...string,
) (InstanceMetricPoints, error) {

	filter := prepareFilter(
		prefix,
		project,
		resourceType,
		metricsNamespace,
		metricName,
		instanceNames...,
	)

	log.Printf("[DEBUG]. Filtering time series with filter: %s", filter)

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
	results := make(InstanceMetricPoints)

	for {
		timeSeries, err := timeSeriesIterator.Next()

		if err != nil {
			if err == iterator.Done {
				break
			}
			return results, err
		}

		resource := timeSeries.Resource
		// this is from instance hostname. It should coincide with names in other API responses
		instanceID := resource.Labels["instance_id"]
		projectID := resource.Labels["project_id"]

		if projectID != project || !strings.HasPrefix(instanceID, helpers.GetPrefix(prefix)) {
			continue
		}
		results[instanceID] = timeSeries.Points

	}

	return results, nil

}

func GetValidatorMetrics(
	ctx context.Context,
	client *monitoring.MetricClient,
	project,
	prefix,
	metricNamespace,
	metricName string,
	instanceNames ...string,
) (InstanceMetricPoints, error) {
	return listMetrics(
		ctx,
		client,
		project,
		prefix,
		resourceType,
		metricNamespace,
		metricName,
		5,
		60,
		instanceNames...,
	)
}
