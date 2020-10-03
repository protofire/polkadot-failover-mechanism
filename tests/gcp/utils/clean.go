package utils

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
)

type cleanFunc func(t *testing.T, project, prefix string) error

func getPrefix(prefix string) string {
	return fmt.Sprintf("%s-", prefix)
}

func contains(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func lastPartOnSplit(s, delimiter string) string {
	return s[strings.LastIndex(s, delimiter)+1:]
}

func waitForOperation(ctx context.Context, service *compute.Service, project string, op *compute.Operation) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for operation to complete")
		case <-ticker.C:
			result, err := service.GlobalOperations.Get(project, op.Name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("GlobalOperations. Get: %s", err)
			}

			if result.Status == "DONE" {
				if result.Error != nil {
					var errors []string
					for _, e := range result.Error.Errors {
						errors = append(errors, e.Message)
					}
					return fmt.Errorf("operation %q failed with error(s): %s", op.Name, strings.Join(errors, ", "))
				}
				return nil
			}
		}
	}
}

// CleanResources cleans gcp resources
func CleanResources(t *testing.T, project, prefix string) error {
	var result *multierror.Error

	funcs := []cleanFunc{SMClean, HealthCheckClean, SAClean, NetworkClean}

	for _, fnc := range funcs {
		err := fnc(t, project, prefix)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
