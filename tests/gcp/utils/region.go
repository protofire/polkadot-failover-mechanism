package utils

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
)

//nolint
func getRegionZones(ctx context.Context, client *compute.Service, project string) (map[string][]string, error) {
	zonesList, err := client.Zones.List(project).Context(ctx).Do()

	if err != nil {
		return nil, fmt.Errorf("Cannot get zones: %#w", err)
	}

	regionZones := make(map[string][]string)

	for _, zone := range zonesList.Items {
		region := lastPartOnSplit(zone.Region, "/")
		regionZones[region] = append(regionZones[region], zone.Name)
	}

	return regionZones, nil

}
