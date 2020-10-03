package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/compute/v1"
)

type subnetItem struct {
	name    string
	region  string
	network string
}

func getNetworks(ctx context.Context, t *testing.T, client *compute.Service, project, prefix string) ([]string, error) {
	var networkNames []string

	networksList, err := client.Networks.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get networks list: %#w", err)
	}

	networks := networksList.Items

	for _, network := range networks {
		t.Logf("Processing network %s", network.Name)

		if strings.HasPrefix(network.Name, getPrefix(prefix)) {
			networkNames = append(networkNames, network.Name)
		}
	}
	return networkNames, nil
}

func deleteNetworkSubnets(ctx context.Context, t *testing.T, client *compute.Service, project, prefix string, networkNames []string) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	subnetsList, err := client.Subnetworks.AggregatedList(project).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("Cannot get network subnets: %#w", err)
	}

	var subnets []subnetItem

	for _, subnet := range subnetsList.Items {
		for _, subnet := range subnet.Subnetworks {
			networkName := lastPartOnSplit(subnet.Network, "/")
			if _, ok := contains(networkNames, networkName); !ok {
				continue
			}
			regionName := lastPartOnSplit(subnet.Region, "/")
			if strings.HasPrefix(subnet.Name, getPrefix(prefix)) {
				subnets = append(subnets, subnetItem{
					network: networkName,
					region:  regionName,
					name:    subnet.Name,
				})
			}
		}
	}

	for _, subnet := range subnets {
		wg.Add(1)

		go func(subnet subnetItem, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			if op, err = client.Subnetworks.Delete(project, subnet.region, subnet.name).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete subnetwork %s in region %s. %#w", subnet.name, subnet.region, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete subnetwork %s in region %s. %s", subnet.name, subnet.region, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, client, project, op); err != nil {
				ch <- fmt.Errorf("Delete operations for subnet %s in region %s has not finished in time", subnet.name, subnet.region)
			}

			t.Logf("Successfully deleted subnetwork: %s", subnet.name)

		}(subnet, wg)

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

func deleteNetworkFirewals(ctx context.Context, t *testing.T, client *compute.Service, project, prefix string, networkNames []string) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	firewallsList, err := client.Firewalls.List(project).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("Cannot get firewalls: %#w", err)
	}

	var firewalls []string

	for _, firewall := range firewallsList.Items {
		networkName := lastPartOnSplit(firewall.Network, "/")
		if _, ok := contains(networkNames, networkName); !ok {
			continue
		}
		if strings.HasPrefix(firewall.Name, getPrefix(prefix)) {
			firewalls = append(firewalls, firewall.Name)
		}
	}

	for _, firewall := range firewalls {
		wg.Add(1)

		go func(firewall string, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			if op, err = client.Firewalls.Delete(project, firewall).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete firewall %s. %#w", firewall, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete firewall %s. %s", firewall, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, client, project, op); err != nil {
				ch <- fmt.Errorf("Delete operations for firewall %s has not finished in time", firewall)

			}

			t.Logf("Successfully deleted firewall: %s.", firewall)

		}(firewall, wg)

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

func deleteNetworks(ctx context.Context, t *testing.T, client *compute.Service, project, prefix string, networkNames []string) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, networkName := range networkNames {

		wg.Add(1)

		go func(network string, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			if op, err = client.Networks.Delete(project, network).Context(ctx).Do(); err != nil {
				ch <- fmt.Errorf("Could not delete network %s. %#w", network, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete network %s. %s", network, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, client, project, op); err != nil {
				ch <- fmt.Errorf("Delete operations for network %s has not finished in time", network)
			}

			t.Logf("Successfully deleted network: %s.", network)

		}(networkName, wg)

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

// NetworkClean cleans all compute networks with prefix
func NetworkClean(t *testing.T, project, prefix string) error {

	ctx := context.Background()

	client, err := compute.NewService(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %#w", err)
	}

	networkNames, err := getNetworks(ctx, t, client, project, prefix)

	if err != nil {
		return err
	}

	if len(networkNames) == 0 {
		t.Logf("Not found networks to delete")
		return nil
	}

	t.Logf("Prepared networks to delete: %s", strings.Join(networkNames, ", "))

	if err := deleteNetworkSubnets(ctx, t, client, project, prefix, networkNames); err != nil {
		return err
	}
	if err := deleteNetworkFirewals(ctx, t, client, project, prefix, networkNames); err != nil {
		return err
	}
	if err := deleteNetworks(ctx, t, client, project, prefix, networkNames); err != nil {
		return err
	}

	networkNames, err = getNetworks(ctx, t, client, project, prefix)

	if err != nil {
		return err
	} else if len(networkNames) > 0 {
		return fmt.Errorf("Cannot delete networks. Existing networks: %s", strings.Join(networkNames, ", "))
	}

	return nil

}
