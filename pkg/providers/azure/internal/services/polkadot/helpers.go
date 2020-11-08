package polkadot

import (
	"path"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
)

func getVmsToDelete(vmScaleSetVMs azure.VMSMap, validatorHostname string) map[string][]string {

	results := make(map[string][]string)

	for vmssName, vms := range vmScaleSetVMs {
		for _, vm := range vms {
			vmHostname := vm.OsProfile.ComputerName
			if vmHostname == nil || *vmHostname != validatorHostname {
				results[vmssName] = append(results[vmssName], path.Base(*vm.ID))
			}
		}
	}

	return results
}

func getValidatorLocation(vmScaleSetVMs azure.VMSMap, locations []string, validatorScaleSetName string) int {

	if validatorScaleSetName == "" {
		return -1
	}

	for _, vm := range vmScaleSetVMs[validatorScaleSetName] {
		validatorLocation := *vm.Location
		if locationIdx := helpers.FindStrIndex(validatorLocation, locations); locationIdx != -1 {
			return locationIdx
		}
	}

	return -1
}
