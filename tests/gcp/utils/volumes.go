package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
)

// VolumesCheck checl that we do not have unattached volumes
func VolumesCheck(prefix, project string) error {

	ctx := context.Background()
	client, err := compute.NewService(ctx)

	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %#w", err)
	}

	volumes, err := client.Disks.AggregatedList(project).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("Cannot get volumes: %#w", err)
	}

	var diskErrors []*compute.Disk

	for _, volume := range volumes.Items {
		for _, disk := range volume.Disks {

			if !strings.HasPrefix(disk.Name, getPrefix(prefix)) {
				continue
			}

			if disk.Status != "READY" {
				diskErrors = append(diskErrors, disk)
				continue
			}

			if len(disk.Users) == 0 {
				diskErrors = append(diskErrors, disk)
			}

		}
	}

	var result *multierror.Error
	if len(diskErrors) > 0 {
		for _, disk := range diskErrors {
			result = multierror.Append(
				result,
				fmt.Errorf("Unattached or non-ready disk found: %q. Status: %q. Users mounted by: %s",
					disk.Name,
					disk.Status,
					strings.Join(disk.Users, "\n"),
				),
			)
		}
	}

	return result.ErrorOrNil()
}
