package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type subnetItem struct {
	name    string
	region  string
	network string
}

func getNetworks(ctx context.Context, client *compute.Service, project, prefix string) ([]string, error) {
	var networkNames []string

	networksList, err := client.Networks.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Cannot get networks list: %w", err)
	}

	networks := networksList.Items

	for _, network := range networks {

		if strings.HasPrefix(network.Name, getPrefix(prefix)) {
			networkNames = append(networkNames, network.Name)
		}
	}
	return networkNames, nil
}

func deleteNetworkSubnets(ctx context.Context, client *compute.Service, project, prefix string, networkNames []string, dryRun bool) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	subnetsList, err := client.Subnetworks.AggregatedList(project).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("Cannot get network subnets: %w", err)
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

			log.Printf("Deleting subnetwork: %s", subnet.name)

			if dryRun {
				return
			}

			if op, err = client.Subnetworks.Delete(project, subnet.region, subnet.name).Context(ctx).Do(); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete subnet: %q. Status: %d\n", subnet.name, gErr.Code)
					return
				}
				ch <- fmt.Errorf("Could not delete subnetwork %q in region %q. %w", subnet.name, subnet.region, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete subnetwork %q in region %q. %s", subnet.name, subnet.region, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, op, prepareRegionGetOp(ctx, client, project, subnet.region)); err != nil {
				ch <- fmt.Errorf("Delete operations for subnet %q in region %q has not finished in time: %w", subnet.name, subnet.region, err)
			}

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

func getFirewalls(ctx context.Context, client *compute.Service, project, prefix string) ([]*compute.Firewall, error) {
	firewallsList, err := client.Firewalls.List(project).Context(ctx).Do()

	if err != nil {
		return nil, fmt.Errorf("Cannot get firewalls: %w", err)
	}

	var firewalls []*compute.Firewall

	for _, firewall := range firewallsList.Items {
		if strings.HasPrefix(firewall.Name, getPrefix(prefix)) {
			firewalls = append(firewalls, firewall)
		}
	}

	return firewalls, nil

}

func deleteNetworkFirewalls(ctx context.Context, client *compute.Service, project, prefix string, networkNames []string, dryRun bool) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	firewallsList, err := client.Firewalls.List(project).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("Cannot get firewalls: %w", err)
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

			log.Printf("Deleting firewall: %q", firewall)

			if dryRun {
				return
			}

			if op, err = client.Firewalls.Delete(project, firewall).Context(ctx).Do(); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete firewall: %q. Status: %d\n", firewall, gErr.Code)
					return
				}
				ch <- fmt.Errorf("Could not delete firewall %q. %w", firewall, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete firewall %q. %s", firewall, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, op, prepareGlobalGetOp(ctx, client, project)); err != nil {
				ch <- fmt.Errorf("Delete operations for firewall %q has not finished in time: %w", firewall, err)

			}

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

func deleteNetworks(ctx context.Context, client *compute.Service, project string, networkNames []string, dryRun bool) error {

	ch := make(chan error)
	wg := &sync.WaitGroup{}

	for _, networkName := range networkNames {

		wg.Add(1)

		go func(network string, wg *sync.WaitGroup) {

			defer wg.Done()

			var op *compute.Operation
			var err error

			log.Printf("Deleting network: %s", network)

			if dryRun {
				return
			}

			if op, err = client.Networks.Delete(project, network).Context(ctx).Do(); err != nil {
				if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
					log.Printf("Cannot delete network: %q. Status: %d\n", network, gErr.Code)
					return
				}
				ch <- fmt.Errorf("Could not delete network %q. %w", network, err)
				return
			}

			if op.Error != nil {
				for _, err := range op.Error.Errors {
					ch <- fmt.Errorf("Could not delete network %q. %s", network, err.Message)
				}
				return
			}

			if err := waitForOperation(ctx, op, prepareGlobalGetOp(ctx, client, project)); err != nil {
				ch <- fmt.Errorf("Delete operations for network %q has not finished in time: %w", network, err)
			}

			log.Printf("Successfully deleted network: %q\n", network)

		}(networkName, wg)

	}

	return helpers.WaitOnErrorChannel(ch, wg)

}

// NetworkClean cleans all compute networks with prefix
func NetworkClean(project, prefix string, dryRun bool) error {
	ctx := context.Background()

	client, err := compute.NewService(ctx)
	if err != nil {
		return fmt.Errorf("Cannot initialize compute client: %w", err)
	}

	networkNames, err := getNetworks(ctx, client, project, prefix)

	if err != nil {
		return err
	}

	if len(networkNames) == 0 {
		log.Println("Not found networks to delete")
		return nil
	}

	log.Printf("Prepared networks to delete: %s\n", strings.Join(networkNames, ", "))

	if err := deleteNetworkSubnets(ctx, client, project, prefix, networkNames, dryRun); err != nil {
		return err
	}
	if err := deleteNetworkFirewalls(ctx, client, project, prefix, networkNames, dryRun); err != nil {
		return err
	}
	if err := deleteNetworks(ctx, client, project, networkNames, dryRun); err != nil {
		return err
	}

	return nil

}

func adjustFirewall(f *compute.Firewall) {
	sort.Strings(f.SourceRanges)
	for _, al := range f.Allowed {
		sort.Strings(al.Ports)
	}

	sort.Slice(f.Allowed, func(i, j int) bool {
		allowed := f.Allowed
		if allowed[i].IPProtocol < allowed[j].IPProtocol {
			return true
		}
		ports1 := allowed[i].Ports
		ports2 := allowed[j].Ports
		for p, k := 0, 0; p < len(ports1) && k < len(ports2); p, k = p+1, k+1 {
			if ports1[p] == ports2[k] {
				continue
			}
			return ports1[p] < ports2[k]
		}
		return len(ports1) < len(ports2)
	})
}

func compareFirewall(f1, f2 *compute.Firewall) bool {
	return f1.Direction == f2.Direction &&
		int(f1.Priority) == int(f2.Priority) &&
		cmp.Equal(f1.SourceRanges, f2.SourceRanges) &&
		cmp.Equal(f1.Allowed, f2.Allowed)
}

func prepareTestFirewalls() []*compute.Firewall {
	fw1 := &compute.Firewall{
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"8080"},
			},
		},
		Direction:    "INGRESS",
		SourceRanges: []string{"35.191.0.0/16", "130.211.0.0/22"},
		Priority:     1003,
	}
	fw2 := &compute.Firewall{
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"8500", "8600", "8300", "8301", "8302"},
			},
			{
				IPProtocol: "udp",
				Ports:      []string{"8500", "8600", "8301", "8302"},
			},
		},
		Direction: "INGRESS",
		Priority:  1001,
	}
	fw3 := &compute.Firewall{
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      []string{"30333"},
			},
			{
				IPProtocol: "udp",
				Ports:      []string{"30333"},
			},
			{
				IPProtocol: "tcp",
				Ports:      []string{"22"},
			},
		},
		Direction:    "INGRESS",
		SourceRanges: []string{"0.0.0.0/0"},
		Priority:     1000,
	}

	firewalls := []*compute.Firewall{fw1, fw2, fw3}

	for _, f := range firewalls {
		adjustFirewall(f)
	}
	return firewalls
}

// FirewallCheck checks created firewalls
func FirewallCheck(prefix, project string) error {

	ctx := context.Background()
	client, err := compute.NewService(ctx)

	if err != nil {
		return fmt.Errorf("Cannot create compute client: %w", err)
	}

	firewalls, err := getFirewalls(ctx, client, project, prefix)
	if err != nil {
		return fmt.Errorf("Cannot get list of firewalls: %w", err)
	}
	for _, f := range firewalls {
		adjustFirewall(f)
	}

	testFireWalls := prepareTestFirewalls()

	for _, testFirewall := range testFireWalls {
		found := false
		for _, firewall := range firewalls {
			if compareFirewall(testFirewall, firewall) {
				found = true
				break
			}
		}
		if !found {
			testFirewallRepresentation, _ := json.MarshalIndent(testFirewall, "", "  ")
			firewallsRepresentation, _ := json.MarshalIndent(firewalls, "", "  ")
			return fmt.Errorf("Cannot find firewall %s from list %s", testFirewallRepresentation, firewallsRepresentation)
		}
	}

	return nil

}
