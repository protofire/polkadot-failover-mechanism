package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// nolint
func getInstanceGroups(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.InstanceGroup, error) {

	var instanceGroups []*compute.InstanceGroup
	instanceGroupsList, err := client.InstanceGroups.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get instance group list: %w", err)
	}

	for _, items := range instanceGroupsList.Items {
		for _, ig := range items.InstanceGroups {
			if len(prefix) > 0 && !strings.HasPrefix(ig.Name, getPrefix(prefix)) {
				continue
			}
			instanceGroups = append(instanceGroups, ig)
		}
	}
	return instanceGroups, nil
}

func getInstanceGroupManagers(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.InstanceGroupManager, error) {

	var instanceGroupManagers []*compute.InstanceGroupManager
	instanceGroupManagerList, err := client.InstanceGroupManagers.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get instance group managers list: %w", err)
	}

	for _, items := range instanceGroupManagerList.Items {
		for _, igm := range items.InstanceGroupManagers {
			if len(prefix) > 0 && !strings.HasPrefix(igm.Name, getPrefix(prefix)) {
				continue
			}
			instanceGroupManagers = append(instanceGroupManagers, igm)
		}
	}
	return instanceGroupManagers, nil
}

func getManagementIntances(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.ManagedInstance, error) {

	instanceGroupManagers, err := getInstanceGroupManagers(ctx, client, project, prefix)

	if err != nil {
		return nil, err
	}

	var managedInstances []*compute.ManagedInstance

	for _, igm := range instanceGroupManagers {
		resp, err := client.RegionInstanceGroupManagers.ListManagedInstances(project, lastPartOnSplit(igm.Region, "/"), igm.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("Cannot get managed instances for instance group manager %q: %w", igm.Name, err)
		}
		managedInstances = append(managedInstances, resp.ManagedInstances...)

	}

	return managedInstances, nil

}

// nolint
func getInstances(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.Instance, error) {

	var instances []*compute.Instance
	instancesList, err := client.Instances.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get instances list: %w", err)
	}

	for _, items := range instancesList.Items {
		for _, in := range items.Instances {
			if len(prefix) > 0 && !strings.HasPrefix(in.Name, getPrefix(prefix)) {
				continue
			}
			instances = append(instances, in)
		}
	}
	return instances, nil
}

// HealthStatusCheck checks instances health status
func HealthStatusCheck(prefix, project string) error {

	ctx := context.Background()
	client, err := compute.NewService(ctx)

	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %w", err)
	}

	instances, err := getManagementIntances(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get instances: %w", err)
	}

	if len(instances) == 0 {
		return errors.New("Not found instances")
	}

	var result *multierror.Error

	for _, instance := range instances {
		for _, instanceHealth := range instance.InstanceHealth {
			if instanceHealth.DetailedHealthState != "HEALTHY" {
				result = multierror.Append(
					result,
					fmt.Errorf(
						"Instance %q in state %q",
						lastPartOnSplit(instance.Instance, "/"),
						instanceHealth.DetailedHealthState,
					),
				)
			}
		}
	}

	return result.ErrorOrNil()

}

// InstanceGroupsClean cleans instance groups
func InstanceGroupsClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := compute.NewService(ctx)

	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %w", err)
	}

	instanceGroups, err := getInstanceGroupManagers(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("Cannot get instances groups: %w", err)
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
