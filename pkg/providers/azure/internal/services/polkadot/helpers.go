package polkadot

import (
	"fmt"
	"path"
	"strings"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
)

type vmssWithInstances struct {
	vmssName string
	vmsIDs   []string
}

type vmssWithInstancesList []vmssWithInstances

func (v vmssWithInstancesList) String() string {
	var pairs []string
	for _, vmss := range v {
		pairs = append(pairs, fmt.Sprintf("%s => %s", vmss.vmssName, strings.Join(vmss.vmsIDs, ", ")))
	}
	return strings.Join(pairs, ". ")
}

func getVmsToDelete(vmScaleSetVMs azure.VMSMap, validatorHostname string) vmssWithInstancesList {

	var results vmssWithInstancesList

	for vmssName, vms := range vmScaleSetVMs {
		vmss := vmssWithInstances{vmssName: vmssName}
		for _, vm := range vms {
			vmHostname := vm.OsProfile.ComputerName
			if vmHostname == nil || *vmHostname != validatorHostname {
				vmss.vmsIDs = append(vmss.vmsIDs, path.Base(*vm.ID))
			}
		}
		if len(vmss.vmsIDs) > 0 {
			results = append(results, vmss)
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
