package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/stretchr/testify/require"
)

func TestVMSMap_Size(t *testing.T) {
	vmScaleSetVMs := VMSMap{
		"1": []compute.VirtualMachineScaleSetVM{
			{},
			{},
			{},
			{},
		},
		"2": []compute.VirtualMachineScaleSetVM{
			{},
			{},
			{},
			{},
		},
		"3": []compute.VirtualMachineScaleSetVM{
			{},
			{},
			{},
			{},
		},
	}

	require.Equal(t, 12, vmScaleSetVMs.Size())
	require.Len(t, vmScaleSetVMs, 3)

}
