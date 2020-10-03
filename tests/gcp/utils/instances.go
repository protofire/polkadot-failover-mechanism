package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func getInstanceGroups(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.InstanceGroupManager, error) {

	var instanceGroups []*compute.InstanceGroupManager
	instanceGroupManagerList, err := client.InstanceGroupManagers.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get instance group list: %#w", err)
	}

	for _, items := range instanceGroupManagerList.Items {
		for _, igm := range items.InstanceGroupManagers {
			if !strings.HasPrefix(igm.Name, getPrefix(prefix)) {
				continue
			}
			instanceGroups = append(instanceGroups, igm)
		}
	}
	return instanceGroups, nil
}

// InstanceGroupsClean cleans instance groups
func InstanceGroupsClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := compute.NewService(ctx)

	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %#w", err)
	}

	instanceGroups, err := getInstanceGroups(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get instances groups: %#w", err)
	}

	if len(instanceGroups) == 0 {
		log.Println("Not found instance groups to delete")
		return nil
	}

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, instanceGroup := range instanceGroups {

		region := lastPartOnSplit(instanceGroup.Region, "/")

		wg.Add(1)

		go func(ig *compute.InstanceGroupManager, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			log.Printf("Deleting instances group: %s. Region: %s\n", ig.Name, region)

			if dryRun {
				return
			}

			if op, err = client.RegionInstanceGroupManagers.Delete(project, region, ig.Name).Context(ctx).Do(); err != nil {

				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete instances group: %s. Region: %s. Status: %d\n", ig.Name, region, gErr.Code)
					return
				}

				ch <- fmt.Errorf("Could not delete instance group: %s. Region: %s. %w", ig.Name, region, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete instance group: %s. Region: %s %s", ig.Name, region, err.Message)
				}
				return
			}

			log.Printf("Waiting till deleting operation is being processed")

			if err := waitForOperation(ctx, op, prepareRegionGetOp(ctx, client, project, region, op.Name)); err != nil {
				ch <- fmt.Errorf("Delete operations for instance group %q has not finished in time: %w", ig.Name, err)
			}

			log.Printf("Successfully deleted instance group: %s.", ig.Name)

		}(instanceGroup, wg)
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
