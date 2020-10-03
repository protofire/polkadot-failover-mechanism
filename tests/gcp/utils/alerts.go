package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func listAlerts(ctx context.Context, client *monitoring.AlertPolicyClient, project, prefix string) ([]string, error) {

	fullPrefix := getPrefix(prefix)

	alertsReq := &monitoringpb.ListAlertPoliciesRequest{
		Name:   "projects/" + project,
		Filter: fmt.Sprintf("name = starts_with('%s') OR display_name = starts_with('%s')", fullPrefix, fullPrefix),
		// Filter:  "", // See https://cloud.google.com/monitoring/api/v3/sorting-and-filtering.
		// OrderBy: "", // See https://cloud.google.com/monitoring/api/v3/sorting-and-filtering.
	}
	alertIt := client.ListAlertPolicies(ctx, alertsReq)

	var alerts []string

	for {
		alert, err := alertIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return alerts, err
		}

		shortName := lastPartOnSplit(alert.Name, "/")
		shortDisplayName := lastPartOnSplit(alert.Name, "/")

		if strings.HasPrefix(shortName, fullPrefix) || strings.HasPrefix(shortDisplayName, fullPrefix) {
			alerts = append(alerts, alert.Name)
		}
		alerts = append(alerts, alert.Name)
	}
	return alerts, nil
}

func deleteAlerts(ctx context.Context, client *monitoring.AlertPolicyClient, project, prefix string, alertNames []string, dryRun bool) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, alertName := range alertNames {

		wg.Add(1)

		go func(alert string, wg *sync.WaitGroup) {

			defer wg.Done()

			log.Printf("Deleting alert: %s", alert)

			if dryRun {
				return
			}

			req := &monitoringpb.DeleteAlertPolicyRequest{
				Name: alert,
			}

			if err := client.DeleteAlertPolicy(ctx, req); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete alert: %q. Status: %d\n", alert, gErr.Code)
					return
				}
				ch <- fmt.Errorf("Could not delete alert %q. %w", alert, err)
				return
			}

			log.Printf("Successfully deleted alert: %q\n", alert)

		}(alertName, wg)

	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	var result *multierror.Error

	for err := range ch {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()

}

// AlertPolicyClean cleans test notification alerts
func AlertPolicyClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := monitoring.NewAlertPolicyClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create notification alerts client: %w", err)
	}
	alerts, err := listAlerts(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get notification alerts list: %w", err)
	}

	if len(alerts) == 0 {
		log.Println("Not found alerts to delete")
		return nil
	}

	return deleteAlerts(ctx, client, project, prefix, alerts, dryRun)

}
