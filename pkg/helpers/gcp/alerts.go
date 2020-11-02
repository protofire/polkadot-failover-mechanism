package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func listAlertPolicies(ctx context.Context, client *monitoring.AlertPolicyClient, project, prefix string) ([]*monitoringpb.AlertPolicy, error) {

	fullPrefix := helpers.GetPrefix(prefix)

	alertsReq := &monitoringpb.ListAlertPoliciesRequest{
		Name:   "projects/" + project,
		Filter: fmt.Sprintf("name = starts_with('%s') OR display_name = starts_with('%s')", fullPrefix, fullPrefix),
		// Filter:  "", // See https://cloud.google.com/monitoring/api/v3/sorting-and-filtering.
		// OrderBy: "", // See https://cloud.google.com/monitoring/api/v3/sorting-and-filtering.
	}
	alertPolicyIterator := client.ListAlertPolicies(ctx, alertsReq)

	var alertPolicies []*monitoringpb.AlertPolicy

	for {
		alertPolicy, err := alertPolicyIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return alertPolicies, err
		}

		shortDisplayName := helpers.LastPartOnSplit(alertPolicy.DisplayName, "/")

		if strings.HasPrefix(shortDisplayName, fullPrefix) {
			alertPolicies = append(alertPolicies, alertPolicy)
		}
	}
	return alertPolicies, nil
}

func deleteAlertPolicies(ctx context.Context, client *monitoring.AlertPolicyClient, alertPolicies []*monitoringpb.AlertPolicy, dryRun bool) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, alertPolicy := range alertPolicies {

		wg.Add(1)

		go func(alertPolicy *monitoringpb.AlertPolicy, wg *sync.WaitGroup) {

			defer wg.Done()

			log.Printf("Deleting alert policy: %q", alertPolicy.Name)

			if dryRun {
				return
			}

			req := &monitoringpb.DeleteAlertPolicyRequest{
				Name: alertPolicy.Name,
			}

			if err := client.DeleteAlertPolicy(ctx, req); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete alert: %q. Status: %d\n", alertPolicy.Name, gErr.Code)
					return
				}
				ch <- fmt.Errorf("Could not delete alert %q. %w", alertPolicy.Name, err)
				return
			}

			log.Printf("Successfully deleted alert: %q\n", alertPolicy.Name)

		}(alertPolicy, wg)

	}

	return helpers.WaitOnErrorChannel(ch, wg)

}

// AlertPolicyClean cleans test notification alerts
func AlertPolicyClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := monitoring.NewAlertPolicyClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create notification alerts client: %w", err)
	}
	alertPolicies, err := listAlertPolicies(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get notification alerts list: %w", err)
	}

	if len(alertPolicies) == 0 {
		log.Println("Not found alerts to delete")
		return nil
	}

	return deleteAlertPolicies(ctx, client, alertPolicies, dryRun)

}

// AlertsPoliciesCheck checks created alert policies
func AlertsPoliciesCheck(prefix, project string) error {

	ctx := context.Background()

	client, err := monitoring.NewAlertPolicyClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create notification alerts client: %w", err)
	}
	alertPolicies, err := listAlertPolicies(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get notification alerts list: %w", err)
	}

	if len(alertPolicies) != 1 {
		return fmt.Errorf("Wrong alert policies count: %d", len(alertPolicies))
	}

	alertPolicyConditions := alertPolicies[0].Conditions

	if len(alertPolicyConditions) != 6 {
		return fmt.Errorf("Wrong alert policy conditions count: %d", len(alertPolicyConditions))
	}

	conditionNames := []string{
		"Health not OK",
		"Health not OK",
		"Validator less than 1",
		"Validator more than 1",
		"Disk consumed more than 90 percent",
		"Validator metric timeseries absent",
	}

	var idx int
	var ok bool
	for _, condition := range alertPolicyConditions {
		if idx, ok = helpers.StringsContains(condition.DisplayName, conditionNames); !ok {
			return fmt.Errorf("cannot find alert policy condition with name: %q", condition.DisplayName)
		}
		helpers.RemoveFromSlice(conditionNames, idx)
	}

	return nil

}
