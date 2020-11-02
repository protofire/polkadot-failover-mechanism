package azure

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-30/compute"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"
)

func checkDisks(prefix, vmName string, disks []compute.DiskInstanceView) error {

	if len(disks) != 2 {
		return fmt.Errorf("VM %q has %d disk attached. Reqired: 2. Disks: %#v", vmName, len(disks), disks)
	}

	found := false
	fullPrefix := helpers.GetPrefix(prefix)
	for _, disk := range disks {
		diskName := *disk.Name
		if strings.HasPrefix(diskName, fullPrefix) {
			found = true
			for _, status := range *disk.Statuses {
				if path.Base(*status.Code) != "succeeded" {
					return fmt.Errorf("VM %q disk %q status is not healthy: %q", vmName, *disk.Name, *status.Code)
				}
			}
		}
	}
	if !found {
		return fmt.Errorf("Cannot get VM %q disks with prefix: %q", vmName, fullPrefix)
	}
	return nil
}

// VolumesCheck checks virtual machine scale set health status
func VolumesCheck(prefix, subscriptionID, resourceGroup string, vms VMSMap) error {

	if len(vms) == 0 {
		return fmt.Errorf("Cannot get health status for no virtual machines")
	}

	client, err := getVMScaleSetVMsClient(subscriptionID)

	if err != nil {
		return nil
	}

	ctx := context.Background()

	for vmssName, vmList := range vms {
		for _, vm := range vmList {
			view, err := client.GetInstanceView(ctx, resourceGroup, vmssName, path.Base(*vm.ID))
			if err != nil {
				return err
			}
			disks := *view.Disks

			if err := checkDisks(prefix, *vm.Name, disks); err != nil {
				return err
			}

		}
	}
	return nil
}
