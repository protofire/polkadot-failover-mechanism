package gcp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
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
			if len(prefix) > 0 && !strings.HasPrefix(ig.Name, helpers.GetPrefix(prefix)) {
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
			if len(prefix) > 0 && !strings.HasPrefix(igm.Name, helpers.GetPrefix(prefix)) {
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
		resp, err := client.RegionInstanceGroupManagers.ListManagedInstances(project, helpers.LastPartOnSplit(igm.Region, "/"), igm.Name).Context(ctx).Do()
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
			if len(prefix) > 0 && !strings.HasPrefix(in.Name, helpers.GetPrefix(prefix)) {
				continue
			}
			instances = append(instances, in)
		}
	}
	return instances, nil
}

// nolint
func getInstanceTemplates(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.InstanceTemplate, error) {

	var instanceTemplates []*compute.InstanceTemplate
	instanceTemplatesList, err := client.InstanceTemplates.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get instance templates: %w", err)
	}

	for _, instanceTemplate := range instanceTemplatesList.Items {
		if len(prefix) > 0 && !strings.HasPrefix(instanceTemplate.Name, helpers.GetPrefix(prefix)) {
			continue
		}
		instanceTemplates = append(instanceTemplates, instanceTemplate)
	}
	return instanceTemplates, nil
}

// InstanceTemplatesClean cleans instance templates
func InstanceTemplatesClean(project, prefix string, dryRun bool) error {

	ctx := context.Background()
	client, err := compute.NewService(ctx)

	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %w", err)
	}

	instanceTemplates, err := getInstanceTemplates(ctx, client, project, prefix)
	if err != nil {
		return err
	}

	if len(instanceTemplates) == 0 {
		log.Println("Not found instance templates to delete")
		return nil
	}

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, instanceTemplate := range instanceTemplates {

		wg.Add(1)

		go func(name string) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			log.Printf("Deleting instance template: %q\n", name)

			if dryRun {
				return
			}

			if op, err = client.InstanceTemplates.Delete(project, name).Context(ctx).Do(); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete instance template: %s. Status: %d\n", name, gErr.Code)
					return
				}

				ch <- fmt.Errorf("Could not delete instance template: %s. %w", name, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete instance template: %s. %s", name, err.Message)
				}
				return
			}

			log.Printf("Waiting till deleting operation is being processed")

			if err := waitForOperation(ctx, op, prepareGlobalGetOp(ctx, client, project)); err != nil {
				ch <- fmt.Errorf("Delete operations for instance template %q has not finished in time: %w", name, err)
			}

			log.Printf("Successfully deleted instance template: %s.", name)

		}(instanceTemplate.Name)
	}

	return helpers.WaitOnErrorChannel(ch, wg)

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
						"instance %q in state %q",
						helpers.LastPartOnSplit(instance.Instance, "/"),
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

		region := helpers.LastPartOnSplit(instanceGroup.Region, "/")

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

			if err := waitForOperation(ctx, op, prepareRegionGetOp(ctx, client, project, region)); err != nil {
				ch <- fmt.Errorf("Delete operations for instance group %q has not finished in time: %w", ig.Name, err)
			}

			log.Printf("Successfully deleted instance group: %s.", ig.Name)

		}(instanceGroup, wg)
	}

	return helpers.WaitOnErrorChannel(ch, wg)

}
