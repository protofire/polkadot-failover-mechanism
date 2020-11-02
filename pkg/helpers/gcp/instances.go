package gcp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	"github.com/hashicorp/go-multierror"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

func getComputeClient() (*compute.Service, error) {
	ctx := context.Background()
	client, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type InstanceGroupManagerList []InstanceGroupManager

func (l InstanceGroupManagerList) InstancesCount() int {
	result := 0
	for _, group := range l {
		result += len(group.Instances)
	}
	return result
}

func (l InstanceGroupManagerList) InstanceNames() []string {
	var result []string
	for _, group := range l {
		result = append(result, group.InstanceNames()...)
	}
	return result
}

type InstanceGroupManager struct {
	Name      string
	Region    string
	Instances []*compute.ManagedInstance
}

func (g InstanceGroupManager) InstanceNames() []string {
	var result []string
	for _, instance := range g.Instances {
		result = append(result, instance.Instance)
	}
	return result
}

func (g *InstanceGroupManager) deleteInstance(i int) {
	lastIndex := len(g.Instances) - 1
	g.Instances[i] = g.Instances[lastIndex]
	g.Instances[lastIndex] = nil
	(*g).Instances = (*g).Instances[:lastIndex]
}

func (g *InstanceGroupManager) SearchAndRemoveInstanceByName(name string) *compute.ManagedInstance {
	for idx, instance := range g.Instances {
		if helpers.LastPartOnSplit(instance.Instance, "/") == name {
			g.deleteInstance(idx)
			return instance
		}
	}
	return nil
}

func operationDescription(op *compute.Operation) string {
	return fmt.Sprintf("Operation %q with type %q status %q and Kind %q", op.Name, op.OperationType, op.Status, op.Kind)
}

func processOpResult(op *compute.Operation, err error) error {

	if err != nil {
		if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
			log.Printf("Cannot process operation: %s. Ansent object. Status: %d\n", operationDescription(op), gErr.Code)
			return nil
		}

		return fmt.Errorf("could not process operation: %s. %w", operationDescription(op), err)
	}

	if op != nil && op.Error != nil {
		result := &multierror.Error{}
		for _, err := range op.Error.Errors {
			result = multierror.Append(result, fmt.Errorf("could not process operation: %s. %s", operationDescription(op), err.Message))
		}
		return result.ErrorOrNil()
	}
	return nil
}

func DeleteManagementInstances(ctx context.Context, client *compute.Service, project string, groups InstanceGroupManagerList) error {

	var values []interface{}

	for _, group := range groups {
		values = append(values, group)
	}

	out := fanout.ConcurrentResponseErrors(ctx, func(ctx context.Context, value interface{}) error {

		group := value.(InstanceGroupManager)

		if len(group.Instances) == 0 {
			return nil
		}

		var instanceIDs []string

		for _, instance := range group.Instances {
			instanceIDs = append(instanceIDs, instance.Instance)
		}
		op, err := client.RegionInstanceGroupManagers.DeleteInstances(
			project,
			group.Region,
			group.Name,
			&compute.RegionInstanceGroupManagersDeleteInstancesRequest{
				Instances: instanceIDs,
			},
		).Context(ctx).Do()
		if op == nil {
			log.Printf("[DEBUG] failover: Nil operation for delete instances: %s", strings.Join(instanceIDs, ", "))
			return nil
		}
		if err := processOpResult(op, err); err != nil {
			return err
		}
		log.Printf("[DEBUG] failover: Waiting while instances are being deleted: %s", strings.Join(instanceIDs, ", "))
		if err := waitForOperation(ctx, op, prepareRegionGetOp(ctx, client, project, group.Region)); err != nil {
			return fmt.Errorf(
				"delete operations for instances %s of managent group %q has not finished in time: %w",
				strings.Join(instanceIDs, ", "),
				group.Name,
				err,
			)
		}
		log.Printf("[DEBUG] failover: Instances have been deleted: %s", strings.Join(instanceIDs, ", "))
		return nil
	}, values...)

	return fanout.ReadErrorsChannel(out)

}

func WaitForInstancesCount(
	ctx context.Context,
	client *compute.Service,
	project, prefix string,
	expectedNumber int,
	regions ...string,
) error {

	ticker := time.NewTicker(5 * time.Second)
	waiter := time.NewTimer(600 * time.Second)
	defer ticker.Stop()
	defer waiter.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("[DEBUG] failover: cancelled waiting for expected instances count: %d", expectedNumber)
		case <-waiter.C:
			return fmt.Errorf("[DEBUG] failover: timeout waiting for expected instances count: %d", expectedNumber)
		case <-ticker.C:
			log.Printf("[DEBUG] failover: Getting management group instances")

			instanceGroups, err := GetInstanceGroupManagersForRegions(
				ctx,
				client,
				project,
				prefix,
				regions...,
			)

			if err != nil {
				log.Printf("[ERROR] failover: Cannot get management instance groups: %s", err)
			}

			if instanceGroups.InstancesCount() == expectedNumber {
				log.Printf(
					"[DEBUG] failover: Got expected instances count: %d. Instances: %s",
					instanceGroups.InstancesCount(),
					instanceGroups.InstanceNames(),
				)
				return nil
			}

			for _, group := range instanceGroups {
				log.Printf("[DEBUG] failover: Found instance group name: %s. Instances: %s", group.Name, group.InstanceNames())
			}
		}
	}
}

func getInstanceGroupManagersForRegions(ctx context.Context, client *compute.Service, project, prefix string, regions ...string) ([]*compute.InstanceGroupManager, error) {

	var instanceGroupManagers []*compute.InstanceGroupManager

	for _, region := range regions {
		instanceGroupManagerList, err := client.RegionInstanceGroupManagers.List(project, region).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("cannot get instance group managers list: %w", err)
		}

		for _, item := range instanceGroupManagerList.Items {
			if len(prefix) > 0 && !strings.HasPrefix(item.Name, helpers.GetPrefix(prefix)) {
				continue
			}
			instanceGroupManagers = append(instanceGroupManagers, item)
		}
	}
	return instanceGroupManagers, nil
}

func getInstanceGroupManagers(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.InstanceGroupManager, error) {

	var instanceGroupManagers []*compute.InstanceGroupManager
	instanceGroupManagerList, err := client.InstanceGroupManagers.AggregatedList(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot get instance group managers list: %w", err)
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

func getManagementInstances(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.ManagedInstance, error) {

	var managedInstances []*compute.ManagedInstance

	groups, err := GetInstanceGroupManagers(ctx, client, project, prefix)

	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		managedInstances = append(managedInstances, group.Instances...)
	}

	return managedInstances, nil

}

func getManagementInstancesFromGroups(ctx context.Context, client *compute.Service, project string, groups ...*compute.InstanceGroupManager) (InstanceGroupManagerList, error) {

	var groupManagers []InstanceGroupManager

	for _, igm := range groups {
		region := helpers.LastPartOnSplit(igm.Region, "/")
		resp, err := client.RegionInstanceGroupManagers.ListManagedInstances(project, region, igm.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("cannot get managed instances for instance group manager %q: %w", igm.Name, err)
		}
		groupManagers = append(groupManagers, InstanceGroupManager{
			Name:      igm.Name,
			Region:    region,
			Instances: resp.ManagedInstances,
		})
	}

	return groupManagers, nil

}

// GetManagementInstances ...
func GetInstanceGroupManagers(ctx context.Context, client *compute.Service, project, prefix string) (InstanceGroupManagerList, error) {

	instanceGroupManagers, err := getInstanceGroupManagers(ctx, client, project, prefix)

	if err != nil {
		return nil, err
	}

	return getManagementInstancesFromGroups(ctx, client, project, instanceGroupManagers...)
}

// GetInstanceGroupManagersForRegions ...
func GetInstanceGroupManagersForRegions(ctx context.Context, client *compute.Service, project, prefix string, regions ...string) (InstanceGroupManagerList, error) {

	instanceGroupManagers, err := getInstanceGroupManagersForRegions(ctx, client, project, prefix, regions...)

	if err != nil {
		return nil, err
	}

	return getManagementInstancesFromGroups(ctx, client, project, instanceGroupManagers...)
}

func GetInstanceGroupManagersForRegionsInnerClient(project, prefix string, regions ...string) (InstanceGroupManagerList, error) {

	client, err := getComputeClient()

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	return GetInstanceGroupManagersForRegions(ctx, client, project, prefix, regions...)

}

func getInstanceTemplates(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.InstanceTemplate, error) {

	var instanceTemplates []*compute.InstanceTemplate
	instanceTemplatesList, err := client.InstanceTemplates.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot get instance templates: %w", err)
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
		return fmt.Errorf("cannot initialize compute client: %w", err)
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

				ch <- fmt.Errorf("could not delete instance template: %s. %w", name, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("could not delete instance template: %s. %s", name, err.Message)
				}
				return
			}

			log.Printf("Waiting till deleting operation is being processed")

			if err := waitForOperation(ctx, op, prepareGlobalGetOp(ctx, client, project)); err != nil {
				ch <- fmt.Errorf("delete operations for instance template %q has not finished in time: %w", name, err)
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
		return fmt.Errorf("cannot initialize compute client: %w", err)
	}

	instances, err := getManagementInstances(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("cannot get instances: %w", err)
	}

	if len(instances) == 0 {
		return errors.New("not found instances")
	}

	result := &multierror.Error{}

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
		return fmt.Errorf("cannot initialize compute client: %w", err)
	}

	instanceGroups, err := getInstanceGroupManagers(ctx, client, project, prefix)

	if err != nil {
		return fmt.Errorf("cannot get instances groups: %w", err)
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

				ch <- fmt.Errorf("could not delete instance group: %s. Region: %s. %w", ig.Name, region, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("could not delete instance group: %s. Region: %s %s", ig.Name, region, err.Message)
				}
				return
			}

			log.Printf("Waiting till deleting operation is being processed")

			if err := waitForOperation(ctx, op, prepareRegionGetOp(ctx, client, project, region)); err != nil {
				ch <- fmt.Errorf("delete operations for instance group %q has not finished in time: %w", ig.Name, err)
			}

			log.Printf("Successfully deleted instance group: %s.", ig.Name)

		}(instanceGroup, wg)
	}

	return helpers.WaitOnErrorChannel(ch, wg)

}
