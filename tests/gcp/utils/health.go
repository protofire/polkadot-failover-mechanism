package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
)

// HealthCheckClean cleans all SM keys with prefix
func HealthCheckClean(t *testing.T, project, prefix string) error {

	ctx := context.Background()

	client, err := compute.NewService(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %#w", err)
	}

	var healthCheckNames []string

	healthChecksList, err := client.HealthChecks.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Cannot get health checks list: %#w", err)
	}

	items := healthChecksList.Items

	for _, item := range items {
		for _, check := range item.HealthChecks {
			if strings.HasPrefix(lastPartOnSplit(check.Name, "/"), getPrefix(prefix)) {
				healthCheckNames = append(healthCheckNames, check.Name)
			}
		}
	}

	if len(healthCheckNames) == 0 {
		t.Logf("Not found health checks to delete")
		return nil
	}

	t.Logf("Prepared health checks to delete: %s", strings.Join(healthCheckNames, ", "))

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, healthCheckName := range healthCheckNames {

		wg.Add(1)

		go func(healthCheckName string, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			if op, err = client.HealthChecks.Delete(project, healthCheckName).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete health check %s. %#w", healthCheckName, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete health check %s. %s", healthCheckName, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, client, project, op); err != nil {
				ch <- fmt.Errorf("Delete operations for health check %s has not finished in time", healthCheckName)
			}

			t.Logf("Successfully deleted health check: %s", healthCheckName)

		}(healthCheckName, wg)

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
