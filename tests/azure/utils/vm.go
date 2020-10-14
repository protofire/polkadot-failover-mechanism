package utils

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/protofire/polkadot-failover-mechanism/tests/helpers"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
)

// VMSMap map location to list of virtual machines
type VMSMap map[string][]compute.VirtualMachineScaleSetVM

// IPAddress represents virtual machine scale set IP public address configuration
type IPAddress struct {
	// /subscriptions/6ad71a09-e4a3-44e0-8e5f-df997c709a74/resourceGroups/814_Protofire_Web3/providers/Microsoft.Compute/virtualMachineScaleSets/test-instance-primary/virtualMachines/0/networkInterfaces/test-ni-primary/ipConfigurations/primary-primary/publicIPAddresses/public-primary
	SubscriptionID      string
	ResourceGroup       string
	VMSSName            string
	VMIndex             string
	IFName              string
	IPConfigurationName string
	PublicAdddressName  string
}

// IPAddressFromString build struct from string
func IPAddressFromString(addr string) *IPAddress {
	parts := strings.Split(strings.Trim(addr, "/"), "/")
	return &IPAddress{
		SubscriptionID:      parts[1],
		ResourceGroup:       parts[3],
		VMSSName:            parts[7],
		VMIndex:             parts[9],
		IFName:              parts[11],
		IPConfigurationName: parts[13],
		PublicAdddressName:  parts[15],
	}
}

func getVMScaleSetClient(subscriptionID string) (compute.VirtualMachineScaleSetsClient, error) {

	client := compute.NewVirtualMachineScaleSetsClient(subscriptionID)

	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getVMScaleSetVMsClient(subscriptionID string) (compute.VirtualMachineScaleSetVMsClient, error) {

	client := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getInterfaceClient(subscriptionID string) (network.InterfacesClient, error) {

	client := network.NewInterfacesClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getPublicAddressClient(subscriptionID string) (network.PublicIPAddressesClient, error) {

	client := network.NewPublicIPAddressesClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("Cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getVMInstancesFromVMScaleSet(ctx context.Context, client *compute.VirtualMachineScaleSetVMsClient, resourceGroup, name string) ([]compute.VirtualMachineScaleSetVM, error) {
	result, err := client.List(ctx, resourceGroup, name, "", "", "")

	if err != nil {
		return nil, err
	}

	vms := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		vms = append(vms, result.Values()...)
	}

	return vms, nil

}

func getVMInstancesFromVMScaleSets(ctx context.Context, client *compute.VirtualMachineScaleSetVMsClient, resourceGroup string, vmss []compute.VirtualMachineScaleSet) (map[string][]compute.VirtualMachineScaleSetVM, error) {

	var wg sync.WaitGroup

	type item struct {
		vmssName string
		vms      []compute.VirtualMachineScaleSetVM
		err      error
	}

	ch := make(chan item)

	for _, vms := range vmss {

		wg.Add(1)

		go func(name string, wg *sync.WaitGroup, out chan item) {

			defer wg.Done()

			vmss, err := getVMInstancesFromVMScaleSet(ctx, client, resourceGroup, name)
			if err != nil {
				out <- item{err: fmt.Errorf("Cannot get virtual machines for scale set %q: %w", name, err)}
				return
			}
			out <- item{vms: vmss, vmssName: name}

		}(*vms.Name, &wg, ch)

	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	result := make(map[string][]compute.VirtualMachineScaleSetVM)
	var errs *multierror.Error

	for it := range ch {
		if it.err != nil {
			errs = multierror.Append(errs, it.err)
			continue
		}
		result[it.vmssName] = it.vms
	}

	return result, errs.ErrorOrNil()

}

func getVirtualMachinesScaleSetVMs(subscriptionID, resourceGroup string) (VMSMap, error) {

	client, err := getVMScaleSetClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	result, err := client.List(ctx, resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("Cannot get list of VMs in resource group: %w", err)
	}

	vmsClient, err := getVMScaleSetVMsClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	vms, err := getVMInstancesFromVMScaleSets(ctx, &vmsClient, resourceGroup, result.Values())

	if err != nil {
		return nil, fmt.Errorf("Cannot get list of VMs in resource group: %w", err)
	}

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		pageVms, err := getVMInstancesFromVMScaleSets(ctx, &vmsClient, resourceGroup, result.Values())
		if err != nil {
			return nil, err
		}
		for key, value := range pageVms {
			vms[key] = value
		}

	}

	return vms, nil

}

func getVirtualMachinesScaleSets(subscriptionID, resourceGroup string) ([]compute.VirtualMachineScaleSet, error) {

	client, err := getVMScaleSetClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	result, err := client.List(ctx, resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("Cannot get list of VMs in resource group: %w", err)
	}

	vmss := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		vmss = append(vmss, result.Values()...)
	}

	return vmss, nil

}

func getVirtualMachineScaleSetInterfaces(ctx context.Context, client *network.InterfacesClient, subscriptionID, resourceGroup string, vmss string) ([]network.Interface, error) {

	result, err := client.ListVirtualMachineScaleSetNetworkInterfaces(ctx, resourceGroup, vmss)

	if err != nil {
		return nil, fmt.Errorf("Cannot get virtual machine stateful set %q network interfaces: %w", vmss, err)
	}

	ifss := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		ifss = append(ifss, result.Values()...)
	}
	if err != nil {
		return ifss, err
	}

	return ifss, nil

}

func filterVirtualMachineScaleSets(vms *[]compute.VirtualMachineScaleSet, handler func(vm compute.VirtualMachineScaleSet) bool) {

	start := 0
	for i := start; i < len(*vms); i++ {
		if !handler((*vms)[i]) {
			// vm will be deleted
			continue
		}
		if i != start {
			(*vms)[start], (*vms)[i] = (*vms)[i], (*vms)[start]
		}
		start++
	}

	*vms = (*vms)[:start]

}

// GetVirtualMachineScaleSetVMs gets all test virtual machines
func GetVirtualMachineScaleSetVMs(prefix, subscriptionID, resourceGroup string) (VMSMap, error) {

	vms, err := getVirtualMachinesScaleSetVMs(subscriptionID, resourceGroup)
	if err != nil {
		return nil, err
	}

	for name := range vms {
		if !strings.HasPrefix(name, helpers.GetPrefix(prefix)) {
			delete(vms, name)
		}
	}

	return vms, nil

}

// GetVirtualMachineScaleSets gets all test virtual machines
func GetVirtualMachineScaleSets(prefix, subscriptionID, resourceGroup string) ([]compute.VirtualMachineScaleSet, error) {
	vms, err := getVirtualMachinesScaleSets(subscriptionID, resourceGroup)
	if err != nil {
		return nil, err
	}
	filterVirtualMachineScaleSets(&vms, func(vm compute.VirtualMachineScaleSet) bool {
		return strings.HasPrefix(*vm.Name, helpers.GetPrefix(prefix))
	})

	return vms, nil

}

// VirtualMachineScaleSetVMsByLocation returns VMs by location map
func VirtualMachineScaleSetVMsByLocation(vms VMSMap) VMSMap {
	result := make(VMSMap)
	for _, vmList := range vms {
		for _, vm := range vmList {
			result[*vm.Location] = append(result[*vm.Location], vm)
		}
	}
	return result
}

// VirtualMachineScaleSetIPAddressIDsByLocation returns VMs by location map
func VirtualMachineScaleSetIPAddressIDsByLocation(vms []compute.VirtualMachineScaleSet, subscriptionID, resourceGroup string) (map[string][]string, error) {

	ctx := context.Background()

	interfaceClient, err := getInterfaceClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	publicAPIClient, err := getPublicAddressClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	type item struct {
		result struct {
			ip       string
			location string
		}
		err error
	}

	ch := make(chan item)
	var wg sync.WaitGroup

	for _, vm := range vms {

		wg.Add(1)
		go func(ctx context.Context, vm compute.VirtualMachineScaleSet, wg *sync.WaitGroup, out chan item) {

			defer wg.Done()

			ifss, err := getVirtualMachineScaleSetInterfaces(ctx, &interfaceClient, subscriptionID, resourceGroup, *vm.Name)

			if err != nil {
				out <- item{err: err}
				return
			}

			for _, ifc := range ifss {

				if ifc.InterfacePropertiesFormat == nil {
					continue
				}

				ipConfigurations := ifc.InterfacePropertiesFormat.IPConfigurations

				if ipConfigurations == nil {
					continue
				}

				for _, conf := range *ipConfigurations {
					ipAddress, err := getIPAddressFromID(ctx, &publicAPIClient, *conf.PublicIPAddress.ID)
					if err != nil {
						out <- item{err: err}
						return
					}
					out <- item{struct {
						ip       string
						location string
					}{ip: ipAddress, location: *vm.Location}, nil}
				}
			}
		}(ctx, vm, &wg, ch)
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	ips := make(map[string][]string, len(vms))
	var errs *multierror.Error

	for it := range ch {
		if it.err != nil {
			errs = multierror.Append(errs, it.err)
			continue
		}
		ips[it.result.location] = append(ips[it.result.location], it.result.ip)
	}

	return ips, errs.ErrorOrNil()

}

func getIPAddressFromID(ctx context.Context, client *network.PublicIPAddressesClient, ipAddressID string) (string, error) {

	ipAddr := IPAddressFromString(ipAddressID)

	result, err := client.GetVirtualMachineScaleSetPublicIPAddress(
		ctx,
		ipAddr.ResourceGroup,
		ipAddr.VMSSName,
		ipAddr.VMIndex,
		ipAddr.IFName,
		ipAddr.IPConfigurationName,
		ipAddr.PublicAdddressName,
		"",
	)

	if err != nil {
		return "", fmt.Errorf("Cannot get public IP address: %w", err)
	}

	return *result.IPAddress, nil

}
