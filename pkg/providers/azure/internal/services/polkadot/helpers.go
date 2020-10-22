package polkadot

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-30/compute"
)

func getVmsToDelete(
	vmScaleSetVMs azure.VMSMap,
	locations []string,
	instanceCounts []int,
	validatorScaleSetName,
	validatorHostname string,
) map[string][]string {

	results := make(map[string][]string)

	for vmssName, vms := range vmScaleSetVMs {
		if len(vms) == 0 {
			continue
		}
		loc := *vms[0].Location
		idx := helpers.FindStrIndex(loc, locations)
		requiredVMsCount := instanceCounts[idx]

		if requiredVMsCount >= len(vms) {
			continue
		}

		if vmssName != validatorScaleSetName {
			// just mark for deleting any first
			for _, vm := range vms[:len(vms)-requiredVMsCount] {
				results[vmssName] = append(results[vmssName], path.Base(*vm.ID))
			}
		} else {
			for _, vm := range vms {
				vmHostname := vm.OsProfile.ComputerName
				if requiredVMsCount != 0 {
					// trying not delete validator instances
					if vmHostname == nil || *vmHostname != validatorHostname {
						results[vmssName] = append(results[vmssName], path.Base(*vm.ID))
					}
				} else {
					results[vmssName] = append(results[vmssName], path.Base(*vm.ID))
				}
			}
		}
	}

	return results
}

func getValidatorLocations(vmScaleSetVMs azure.VMSMap, locations []string, validatorScaleSetName string) (string, int) {

	validatorLocation, locationIdx := "", -1

	for _, vm := range vmScaleSetVMs[validatorScaleSetName] {
		validatorLocation = *vm.Location
		locationIdx = helpers.FindStrIndex(validatorLocation, locations)
		if locationIdx != -1 {
			break
		}
	}

	return validatorLocation, locationIdx
}

func deleteVMs(
	ctx context.Context,
	client *compute.VirtualMachineScaleSetsClient,
	resourceGroup,
	validatorScaleSetName string,
	vmScaleSetVMIDsToDelete []string,
) error {

	future, err := client.DeleteInstances(
		ctx,
		resourceGroup,
		validatorScaleSetName,
		compute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIds: &vmScaleSetVMIDsToDelete},
	)
	if err != nil {
		return fmt.Errorf(
			"error deleting Virtual Machines %q from Scale Set %q (Resource Group %q): %w",
			strings.Join(vmScaleSetVMIDsToDelete, ", "),
			validatorScaleSetName,
			resourceGroup,
			err,
		)
	}

	log.Printf(
		"[DEBUG] Waiting for Virtual Machines %q from Scale Set %q (Resource Group %q) to be deleted..",
		strings.Join(vmScaleSetVMIDsToDelete, ", "),
		validatorScaleSetName,
		resourceGroup,
	)
	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf(
			"error waiting to deleting Virtual Machines %q from Scale Set %q (Resource Group %q): %w",
			strings.Join(vmScaleSetVMIDsToDelete, ", "),
			validatorScaleSetName,
			resourceGroup,
			err,
		)
	}
	log.Printf("[DEBUG] Virtual Machines from Scale Set %q (Resource Group %q) was deleted", validatorScaleSetName, resourceGroup)
	return nil
}
