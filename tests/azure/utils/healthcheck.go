package utils

import (
	"context"
	"fmt"
	"path"
)

// HealthStatusCheck checks virtual machine scale set health status
func HealthStatusCheck(subscriptionID, resourceGroup string, vms VMSMap) error {

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
			status := view.VMHealth.Status.Code
			if path.Base(*status) != "healthy" {
				return fmt.Errorf("VM %q status is not healthy: %q", *vm.Name, *status)
			}
		}
	}
	return nil
}
