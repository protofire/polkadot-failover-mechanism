package client

import (
	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/insights"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-30/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-06-01/network"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/common"
)

type Client struct {
	VMScaleSetsClient       *compute.VirtualMachineScaleSetsClient
	VMScaleSetVMsClient     *compute.VirtualMachineScaleSetVMsClient
	VMClient                *compute.VirtualMachinesClient
	InterfacesClient        *network.InterfacesClient
	PublicIPAddressesClient *network.PublicIPAddressesClient
	MetricsClient           *insights.MetricsClient
}

func NewClient(o *common.ClientOptions) *Client {
	vmScaleSetsClient := compute.NewVirtualMachineScaleSetsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionID)
	o.ConfigureClient(&vmScaleSetsClient.Client, o.ResourceManagerAuthorizer)

	vmScaleSetVMsClient := compute.NewVirtualMachineScaleSetVMsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionID)
	o.ConfigureClient(&vmScaleSetVMsClient.Client, o.ResourceManagerAuthorizer)

	vmClient := compute.NewVirtualMachinesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionID)
	o.ConfigureClient(&vmClient.Client, o.ResourceManagerAuthorizer)

	interfacesClient := network.NewInterfacesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionID)
	o.ConfigureClient(&interfacesClient.Client, o.ResourceManagerAuthorizer)

	publicIPAddressClient := network.NewPublicIPAddressesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionID)
	o.ConfigureClient(&publicIPAddressClient.Client, o.ResourceManagerAuthorizer)

	metricsClient := insights.NewMetricsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionID)
	o.ConfigureClient(&metricsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		VMScaleSetsClient:       &vmScaleSetsClient,
		VMScaleSetVMsClient:     &vmScaleSetVMsClient,
		VMClient:                &vmClient,
		InterfacesClient:        &interfacesClient,
		PublicIPAddressesClient: &publicIPAddressClient,
		MetricsClient:           &metricsClient,
	}
}
