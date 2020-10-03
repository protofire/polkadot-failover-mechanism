package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
)

// HealthCheckClean cleans all SM keys with prefix
func HealthCheckClean(project, prefix string, dryRun bool) error {

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
		log.Println("Not found health checks to delete")
		return nil
	}

	log.Printf("Prepared health checks to delete: %s\n", strings.Join(healthCheckNames, ", "))

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, healthCheckName := range healthCheckNames {

		wg.Add(1)

		go func(healthCheckName string, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			log.Printf("Deleting health check: %s", healthCheckName)

			if dryRun {
				return
			}

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

			if err := waitForOperation(ctx, op, prepareGlobalGetOp(ctx, client, project, op.Name)); err != nil {
				ch <- fmt.Errorf("Delete operations for health check %q has not finished in time: %w", healthCheckName, err)
			}

			log.Printf("Successfully deleted health check: %s\n", healthCheckName)

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
