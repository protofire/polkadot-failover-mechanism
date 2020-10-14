package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/protofire/polkadot-failover-mechanism/tests/helpers"
	"google.golang.org/api/compute/v1"
)

func getHealthChecks(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.HealthCheck, error) {

	var healthChecks []*compute.HealthCheck

	healthChecksList, err := client.HealthChecks.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return healthChecks, fmt.Errorf("Cannot get health checks list: %w", err)
	}

	items := healthChecksList.Items

	for _, item := range items {
		for _, check := range item.HealthChecks {
			if strings.HasPrefix(lastPartOnSplit(check.Name, "/"), getPrefix(prefix)) {
				healthChecks = append(healthChecks, check)
			}
		}
	}

	return healthChecks, nil

}

// HealthCheckClean cleans all SM keys with prefix
func HealthCheckClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()

	client, err := compute.NewService(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %w", err)
	}

	healthChecks, err := getHealthChecks(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get health checks list: %w", err)
	}

	if len(healthChecks) == 0 {
		log.Println("Not found health checks to delete")
		return nil
	}

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, healthCheck := range healthChecks {

		wg.Add(1)

		go func(healthCheckName string, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			log.Printf("Deleting health check: %q", healthCheckName)

			if dryRun {
				return
			}

			if op, err = client.HealthChecks.Delete(project, healthCheckName).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete health check %q. %w", healthCheckName, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete health check %q. %q", healthCheckName, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, op, prepareGlobalGetOp(ctx, client, project, op.Name)); err != nil {
				ch <- fmt.Errorf("Delete operations for health check %q has not finished in time: %w", healthCheckName, err)
			}

			log.Printf("Successfully deleted health check: %q\n", healthCheckName)

		}(healthCheck.Name, wg)

	}

	return helpers.WaitOnErrorChannel(ch, wg)

}
