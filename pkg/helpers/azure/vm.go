package azure

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/fanout"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
)

// VMSMap map location to list of virtual machines
type VMSMap map[string][]compute.VirtualMachineScaleSetVM

func (vmss VMSMap) Size() int {
	s := 0
	for _, vms := range vmss {
		s += len(vms)
	}
	return s
}

// IPAddress represents virtual machine scale set IP public address configuration
type IPAddress struct {
	// /subscriptions/6ad71a09-e4a3-44e0-8e5f-df997c709a74/resourceGroups/814_Protofire_Web3/providers/Microsoft.Compute/virtualMachineScaleSets/test-instance-primary/virtualMachines/0/networkInterfaces/test-ni-primary/ipConfigurations/primary-primary/publicIPAddresses/public-primary
	SubscriptionID      string
	ResourceGroup       string
	VMSSName            string
	VMIndex             string
	IFName              string
	IPConfigurationName string
	PublicAddressName   string
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
		PublicAddressName:   parts[15],
	}
}

var retryableServiceErrorCodes = []string{"NetworkingInternalOperationError"}

type future interface {
	WaitForCompletionRef(ctx context.Context, client autorest.Client) (err error)
}

func waitForFuture(ctx context.Context, future future, client *compute.VirtualMachineScaleSetsClient) error {
	for {
		if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
			sErr := &azure.ServiceError{}
			if errors.As(err, sErr) && helpers.StringsContainsBool(sErr.Code, retryableServiceErrorCodes) {
				continue
			}
			return err
		}
		return nil
	}
}

func GetVMScaleSetClient(subscriptionID string) (compute.VirtualMachineScaleSetsClient, error) {

	client := compute.NewVirtualMachineScaleSetsClient(subscriptionID)

	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getVMScaleSetVMsClient(subscriptionID string) (compute.VirtualMachineScaleSetVMsClient, error) {

	client := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getInterfaceClient(subscriptionID string) (network.InterfacesClient, error) {

	client := network.NewInterfacesClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("cannot get authorizer: %w", err)
	}
	client.Authorizer = auth

	return client, nil

}

func getPublicAddressClient(subscriptionID string) (network.PublicIPAddressesClient, error) {

	client := network.NewPublicIPAddressesClient(subscriptionID)
	auth, err := getAuthorizer()
	if err != nil {
		return client, fmt.Errorf("cannot get authorizer: %w", err)
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

func getVMInstancesFromVMScaleSets(
	ctx context.Context,
	client *compute.VirtualMachineScaleSetVMsClient,
	resourceGroup string,
	vmScaleSets []compute.VirtualMachineScaleSet,
) (VMSMap, error) {

	type vmsItem struct {
		vmScaleSetName string
		vms            []compute.VirtualMachineScaleSetVM
	}

	var names []interface{}

	for _, vmScaleSet := range vmScaleSets {
		names = append(names, *vmScaleSet.Name)
	}

	out := fanout.ConcurrentResponseItems(ctx, func(ctx context.Context, value interface{}) (interface{}, error) {
		name := value.(string)
		vmScaleSetVMs, err := getVMInstancesFromVMScaleSet(ctx, client, resourceGroup, name)
		if err != nil {
			return nil, err
		}
		return vmsItem{
			vmScaleSetName: name,
			vms:            vmScaleSetVMs,
		}, nil
	}, names...)

	result := make(VMSMap)

	items, err := fanout.ReadItemChannel(out)

	if err != nil {
		return result, err
	}

	for _, item := range items {
		vmsItem := item.(vmsItem)
		result[vmsItem.vmScaleSetName] = vmsItem.vms
	}

	return result, nil

}

func getVirtualMachinesScaleSetVMs(
	ctx context.Context,
	vmScaleSetClient *compute.VirtualMachineScaleSetsClient,
	vmScaleSetVMsClient *compute.VirtualMachineScaleSetVMsClient,
	resourceGroup string,
) (VMSMap, error) {

	result, err := vmScaleSetClient.List(ctx, resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("cannot get list of VMs in resource group: %w", err)
	}

	vms, err := getVMInstancesFromVMScaleSets(ctx, vmScaleSetVMsClient, resourceGroup, result.Values())

	if err != nil {
		return nil, fmt.Errorf("cannot get list of VMs in resource group: %w", err)
	}

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		pageVms, err := getVMInstancesFromVMScaleSets(ctx, vmScaleSetVMsClient, resourceGroup, result.Values())
		if err != nil {
			return nil, err
		}
		for key, value := range pageVms {
			vms[key] = value
		}

	}

	return vms, nil

}

func getVirtualMachinesScaleSets(ctx context.Context, vmScaleSetClient *compute.VirtualMachineScaleSetsClient, resourceGroup string) ([]compute.VirtualMachineScaleSet, error) {

	result, err := vmScaleSetClient.List(ctx, resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("cannot get list of VMs in resource group: %w", err)
	}

	vmScaleSets := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		vmScaleSets = append(vmScaleSets, result.Values()...)
	}

	return vmScaleSets, nil

}

func getVirtualMachineScaleSetInterfaces(ctx context.Context, client *network.InterfacesClient, resourceGroup string, vmScaleSet string) ([]network.Interface, error) {

	result, err := client.ListVirtualMachineScaleSetNetworkInterfaces(ctx, resourceGroup, vmScaleSet)

	if err != nil {
		return nil, fmt.Errorf("cannot get virtual machine stateful set %q network interfaces: %w", vmScaleSet, err)
	}

	ifss := result.Values()

	for err = result.NextWithContext(ctx); err != nil; err = result.NextWithContext(ctx) {
		ifss = append(ifss, result.Values()...)
	}

	return ifss, nil

}

func FilterVirtualMachineScaleSets(vms *[]compute.VirtualMachineScaleSet, handler func(vm compute.VirtualMachineScaleSet) bool) {

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

// GetVirtualMachineScaleSetVMs gets all virtual machines
func GetVirtualMachineScaleSetVMs(prefix, subscriptionID, resourceGroup string) (VMSMap, error) {

	ctx := context.Background()

	vmScaleSetClient, err := GetVMScaleSetClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	vmScaleSetClientVMs, err := getVMScaleSetVMsClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	return GetVirtualMachineScaleSetVMsWithClient(ctx, &vmScaleSetClient, &vmScaleSetClientVMs, prefix, resourceGroup)
}

// GetVirtualMachineScaleSetVMsWithClient gets all virtual machines
func GetVirtualMachineScaleSetVMsWithClient(
	ctx context.Context,
	vmScaleSetClient *compute.VirtualMachineScaleSetsClient,
	vmScaleSetClientVMs *compute.VirtualMachineScaleSetVMsClient,
	prefix,
	resourceGroup string,
	locations ...string,
) (VMSMap, error) {

	vms, err := getVirtualMachinesScaleSetVMs(ctx, vmScaleSetClient, vmScaleSetClientVMs, resourceGroup)
	if err != nil {
		return nil, err
	}

	for name := range vms {
		if !strings.HasPrefix(name, helpers.GetPrefix(prefix)) {
			delete(vms, name)
		}
	}

	if len(locations) > 0 {
		for name, vmsList := range vms {
			for _, vm := range vmsList {
				if helpers.FindStrIndex(*vm.Location, locations) != -1 {
					delete(vms, name)
					break
				}
			}
		}
	}

	return vms, nil

}

func CheckVirtualMachineScaleSetVMsWithClient(
	ctx context.Context,
	vmScaleSetClient *compute.VirtualMachineScaleSetsClient,
	vmScaleSetClientVMs *compute.VirtualMachineScaleSetVMsClient,
	resourceGroup string,
	vmssNames ...string,
) ([]string, error) {

	vms, err := getVirtualMachinesScaleSetVMs(ctx, vmScaleSetClient, vmScaleSetClientVMs, resourceGroup)
	if err != nil {
		return nil, err
	}

	var vmssNamesOutput []string

	for vmssName, vms := range vms {
		if helpers.StringsContainsBool(vmssName, vmssNames) && len(vms) > 0 {
			vmssNamesOutput = append(vmssNamesOutput, vmssName)
		}
	}

	return vmssNamesOutput, nil

}

func WaitForVirtualMachineScaleSetVMsWithClient(
	ctx context.Context,
	vmScaleSetClient *compute.VirtualMachineScaleSetsClient,
	vmScaleSetClientVMs *compute.VirtualMachineScaleSetVMsClient,
	prefix,
	resourceGroup string,
	size int,
	period int,
	locations ...string,
) (VMSMap, error) {

	ticker := time.NewTicker(time.Duration(period) * time.Second)
	tickerChan := ticker.C

	defer ticker.Stop()

	var err error
	var vms VMSMap

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("period waiting for validator. Context has been cancelled. Last error: %w", err)
		case <-tickerChan:
			vms, err = GetVirtualMachineScaleSetVMsWithClient(
				ctx,
				vmScaleSetClient,
				vmScaleSetClientVMs,
				prefix,
				resourceGroup,
				locations...,
			)
			if err == nil && vms.Size() == size {
				return vms, nil
			}
			if err != nil {
				log.Printf("[ERROR] failover: Got error while was waiting for VM scale sets: %v", err)
			} else {
				log.Printf("[DEBUG] failover: Got %d virtual machines instead %d", vms.Size(), size)
			}
		}
	}

}

// GetVirtualMachineScaleSets gets all test virtual machines
func GetVirtualMachineScaleSets(prefix, subscriptionID, resourceGroup string) ([]compute.VirtualMachineScaleSet, error) {

	ctx := context.Background()

	client, err := GetVMScaleSetClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	return GetVirtualMachineScaleSetsWithClient(ctx, &client, prefix, resourceGroup)

}

// GetVirtualMachineScaleSetsWithClient gets all test virtual machines
func GetVirtualMachineScaleSetsWithClient(
	ctx context.Context,
	client *compute.VirtualMachineScaleSetsClient,
	prefix, resourceGroup string,
	locations ...string,
) ([]compute.VirtualMachineScaleSet, error) {

	vms, err := getVirtualMachinesScaleSets(ctx, client, resourceGroup)
	if err != nil {
		return nil, err
	}
	FilterVirtualMachineScaleSets(&vms, func(vm compute.VirtualMachineScaleSet) bool {
		return strings.HasPrefix(*vm.Name, helpers.GetPrefix(prefix))
	})

	if len(locations) > 0 {
		FilterVirtualMachineScaleSets(&vms, func(vm compute.VirtualMachineScaleSet) bool {
			return helpers.FindStrIndex(*vm.Location, locations) != -1
		})
	}

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
func VirtualMachineScaleSetIPAddressIDsByLocation(vmScaleSets []compute.VirtualMachineScaleSet, subscriptionID, resourceGroup string) (map[string][]string, error) {

	ctx := context.Background()

	interfaceClient, err := getInterfaceClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	publicAPIClient, err := getPublicAddressClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	type ipItem struct {
		ip       string
		location string
	}

	var vmScaleSetInterfaces []interface{}

	for _, vmScaleSet := range vmScaleSets {
		vmScaleSetInterfaces = append(vmScaleSetInterfaces, vmScaleSet)
	}

	out := fanout.ConcurrentResponseItems(ctx, func(ctx context.Context, value interface{}) (interface{}, error) {

		vm := value.(compute.VirtualMachineScaleSet)

		ifss, err := getVirtualMachineScaleSetInterfaces(ctx, &interfaceClient, resourceGroup, *vm.Name)

		if err != nil {
			return nil, err
		}

		var ipItems []ipItem

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
					return nil, err
				}
				ipItems = append(ipItems, ipItem{ip: ipAddress, location: *vm.Location})
			}

		}

		return ipItems, nil

	}, vmScaleSetInterfaces...)

	ips := make(map[string][]string, len(vmScaleSets))

	items, err := fanout.ReadItemChannel(out)

	if err != nil {
		return ips, nil
	}

	for _, item := range items {
		ipItems := item.([]ipItem)
		for _, ipItem := range ipItems {
			ips[ipItem.location] = append(ips[ipItem.location], ipItem.ip)
		}
	}

	return ips, nil

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
		ipAddr.PublicAddressName,
		"",
	)

	if err != nil {
		return "", fmt.Errorf("cannot get public IP address: %w", err)
	}

	return *result.IPAddress, nil

}

func GetVMScaleSetNames(ctx context.Context, client *compute.VirtualMachineScaleSetsClient, resourceGroup, prefix string) ([]string, error) {

	vmScaleSets, err := GetVirtualMachineScaleSetsWithClient(
		ctx,
		client,
		prefix,
		resourceGroup,
	)

	if err != nil {
		return nil, fmt.Errorf("[ERROR]. Cannot get VM scale sets: %w", err)
	}

	vmScaleSetsNames := make([]string, 0, len(vmScaleSets))

	for _, vmScaleSet := range vmScaleSets {
		vmScaleSetsNames = append(vmScaleSetsNames, *vmScaleSet.Name)
	}

	return vmScaleSetsNames, nil

}

func GetVMScaleSetNamesWithInstances(ctx context.Context, client *compute.VirtualMachineScaleSetsClient, resourceGroup, prefix string) ([]string, error) {

	vmScaleSets, err := GetVirtualMachineScaleSetsWithClient(
		ctx,
		client,
		prefix,
		resourceGroup,
	)

	if err != nil {
		return nil, fmt.Errorf("[ERROR]. Cannot get VM scale sets: %w", err)
	}

	vmScaleSetsNames := make([]string, 0, len(vmScaleSets))

	for _, vmScaleSet := range vmScaleSets {
		if *vmScaleSet.Sku.Capacity > 0 {
			vmScaleSetsNames = append(vmScaleSetsNames, *vmScaleSet.Name)
		}
	}

	return vmScaleSetsNames, nil

}

func GetVMScaleSetInstancesCount(ctx context.Context, client *compute.VirtualMachineScaleSetsClient, resourceGroup, prefix string, locations ...string) (map[string]int, error) {

	vmScaleSets, err := GetVirtualMachineScaleSetsWithClient(
		ctx,
		client,
		prefix,
		resourceGroup,
	)

	if err != nil {
		return nil, fmt.Errorf("[ERROR]. Cannot get VM scale sets: %w", err)
	}

	vmScaleSetsNamesToInstancesCount := make(map[string]int)

	for _, vmScaleSet := range vmScaleSets {
		if _, ok := helpers.StringsContains(*vmScaleSet.Location, locations); ok {
			vmScaleSetsNamesToInstancesCount[*vmScaleSet.Name] = int(*vmScaleSet.Sku.Capacity)
		}
	}

	return vmScaleSetsNamesToInstancesCount, nil

}

func setVMssCapacity(
	ctx context.Context,
	client *compute.VirtualMachineScaleSetsClient,
	resourceGroup,
	vmScaleSetName string,
	removeVmsCount int) error {

	vmss, err := client.Get(ctx, resourceGroup, vmScaleSetName)

	if err != nil {
		return fmt.Errorf("cannot get vm scale set %q: %w", vmScaleSetName, err)
	}

	sku := vmss.Sku
	*sku.Capacity = *sku.Capacity - int64(removeVmsCount)

	if *sku.Capacity < 0 {
		*sku.Capacity = 0
	}

	log.Printf("[DEBUG] failover: updating vm scale set %q to capacity %d...", vmScaleSetName, *sku.Capacity)

	updateFuture, err := client.Update(ctx, resourceGroup, vmScaleSetName, compute.VirtualMachineScaleSetUpdate{
		Sku: sku,
	})

	if err != nil {
		return fmt.Errorf("cannot update vm scale set %q: %w", vmScaleSetName, err)
	}

	if err := waitForFuture(ctx, &updateFuture, client); err != nil {
		return fmt.Errorf(
			"error waiting for VM Scale Set %q (Resource Group %q) is being updated. Error type: %T: %w",
			vmScaleSetName,
			resourceGroup,
			err,
			err,
		)
	}

	log.Printf("[DEBUG] failover: updated vm scale set %q to capacity %d", vmScaleSetName, *sku.Capacity)

	return nil

}

func DeleteVMs(
	ctx context.Context,
	client *compute.VirtualMachineScaleSetsClient,
	resourceGroup,
	vmScaleSetName string,
	vmScaleSetVMIDsToDelete []string,
	updateVMssCapacity bool,
) error {

	log.Printf("[DEBUG] failover: deleting vm scale set %q instances %s", vmScaleSetName, vmScaleSetVMIDsToDelete)

	deleteFuture, err := client.DeleteInstances(
		ctx,
		resourceGroup,
		vmScaleSetName,
		compute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIds: &vmScaleSetVMIDsToDelete},
	)
	if err != nil {
		return fmt.Errorf(
			"error deleting Virtual Machines %q from Scale Set %q (Resource Group %q): %w",
			strings.Join(vmScaleSetVMIDsToDelete, ", "),
			vmScaleSetName,
			resourceGroup,
			err,
		)
	}

	log.Printf(
		"[DEBUG] failover: Waiting for Virtual Machines %q from Scale Set %q (Resource Group %q) to be deleted..",
		strings.Join(vmScaleSetVMIDsToDelete, ", "),
		vmScaleSetName,
		resourceGroup,
	)

	if err := waitForFuture(ctx, &deleteFuture, client); err != nil {
		return fmt.Errorf(
			"error waiting for Virtual Machines %q from Scale Set %q (Resource Group %q) is being deleted. Error type: %T: %w",
			strings.Join(vmScaleSetVMIDsToDelete, ", "),
			vmScaleSetName,
			resourceGroup,
			err,
			err,
		)
	}

	log.Printf("[DEBUG] failover: Virtual Machines from Scale Set %q (Resource Group %q) was deleted", vmScaleSetName, resourceGroup)

	if updateVMssCapacity {
		return setVMssCapacity(ctx, client, resourceGroup, vmScaleSetName, len(vmScaleSetVMIDsToDelete))
	}

	return nil
}
