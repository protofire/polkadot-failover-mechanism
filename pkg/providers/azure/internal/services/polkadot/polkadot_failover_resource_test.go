package polkadot

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
)

func TestGetVMsToDelete(t *testing.T) {

	id1, id2, id3 := "id1", "id2", "id3"
	vmSSName1, vmSSName2 := "vmSS1", "vmSS2"
	locationName1, locationName2 := "centralus", "westus"
	hostname1, hostname2 := "hostname1", "hostname2"

	validatorHostname := hostname1

	vms := map[string][]compute.VirtualMachineScaleSetVM{
		vmSSName1: {
			{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					OsProfile: &compute.OSProfile{
						ComputerName: &hostname1,
					},
				},
				ID:       &id1,
				Location: &locationName1,
			},
			{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					OsProfile: &compute.OSProfile{
						ComputerName: &hostname2,
					},
				},
				ID:       &id2,
				Location: &locationName1,
			},
		},
		vmSSName2: {
			{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					OsProfile: &compute.OSProfile{
						ComputerName: &hostname2,
					},
				},
				ID:       &id3,
				Location: &locationName2,
			},
		},
	}

	result := getVmsToDelete(vms, validatorHostname)
	require.Len(t, result, 2)

	for _, vmss := range result {
		if vmss.vmssName == vmSSName1 {
			require.Equal(t, vmss.vmsIDs, []string{id2})
		}
		if vmss.vmssName == vmSSName2 {
			require.Equal(t, vmss.vmsIDs, []string{id3})
		}
	}

	result = getVmsToDelete(vms, "")
	require.Len(t, result, 2)
	for _, vmss := range result {
		if vmss.vmssName == vmSSName1 {
			require.Equal(t, vmss.vmsIDs, []string{id1, id2})
		}
		if vmss.vmssName == vmSSName2 {
			require.Equal(t, vmss.vmsIDs, []string{id3})
		}
	}
}
