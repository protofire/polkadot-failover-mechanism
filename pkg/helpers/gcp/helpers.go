package gcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type getOp func(opName string) (*compute.Operation, error)

func prepareGlobalGetOp(ctx context.Context, client *compute.Service, project string) getOp {
	return func(opName string) (*compute.Operation, error) {
		return client.GlobalOperations.Get(project, opName).Context(ctx).Do()
	}
}

func prepareRegionGetOp(ctx context.Context, client *compute.Service, project, region string) getOp {
	return func(opName string) (*compute.Operation, error) {
		return client.RegionOperations.Get(project, region, opName).Context(ctx).Do()
	}
}

//nolint
func prepareZoneGetOp(ctx context.Context, client *compute.Service, project, zone string) getOp {
	return func(opName string) (*compute.Operation, error) {
		return client.ZoneOperations.Get(project, zone, opName).Context(ctx).Do()
	}
}

func waitForOperation(ctx context.Context, op *compute.Operation, getOp getOp) error {

	ticker := time.NewTicker(5 * time.Second)
	waiter := time.NewTimer(600 * time.Second)
	defer ticker.Stop()
	defer waiter.Stop()

	result := op

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			waiter.Stop()
			return fmt.Errorf("cancelled waiting for operation %q to complete", result.Name)
		case <-waiter.C:
			return fmt.Errorf("timeout waiting for operation %q to complete", result.Name)
		case <-ticker.C:
			var err error
			result, err = getOp(result.Name)
			if result == nil {
				result = op
			}
			if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
				log.Printf("Not found operation %q with type %q status %q and kind %q", result.Name, result.OperationType, result.Status, result.Kind)
				break
			}
			if err != nil {
				return fmt.Errorf("Cannot get operations: %q. %w", result.Name, err)
			}

			log.Printf("Operation %q with type %q status %q and kind %q", result.Name, result.OperationType, result.Status, result.Kind)

			if result.Status != "DONE" {
				break
			}
			if result.Error != nil {
				var errors []string
				for _, e := range result.Error.Errors {
					errors = append(errors, e.Message)
				}
				return fmt.Errorf("operation %q failed with error(s): %s", result.Name, strings.Join(errors, ", "))
			}
			return nil
		}
	}
}
