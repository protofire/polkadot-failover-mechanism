package azure

import (
	"context"
	"fmt"
	"path"
)

// HealthStatusCheck checks virtual machine scale set health status
func HealthStatusCheck(subscriptionID, resourceGroup string, vms VMSMap) error {

	if len(vms) == 0 {
		return fmt.Errorf("cannot get health status for no virtual machines")
	}

	client, err := getVMScaleSetVMsClient(subscriptionID)

	if err != nil {
		return nil
	}

	ctx := context.Background()

	for vmScaleSetName, vmScaleSetVMList := range vms {
		for _, vm := range vmScaleSetVMList {
			view, err := client.GetInstanceView(ctx, resourceGroup, vmScaleSetName, path.Base(*vm.ID))
			if err != nil {
				return err
			}
			if view.VMHealth == nil {
				return fmt.Errorf("instance view is null for VM %q", *vm.Name)
			}
			if view.VMHealth.Status == nil {
				return fmt.Errorf("instance view health status is null for VM %q", *vm.Name)
			}
			status := view.VMHealth.Status.Code
			if path.Base(*status) != "healthy" {
				return fmt.Errorf("VM %q status is not healthy: %q", *vm.Name, *status)
			}
		}
	}
	return nil
}
