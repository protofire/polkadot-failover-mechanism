package clients

import (
	polkadotClient "github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/services/polkadot/client"

	"github.com/Azure/go-autorest/autorest"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/common"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/features"
)

// Client instance
type Client struct {
	Account  *ResourceManagerAccount
	Features features.UserFeatures
	Polkadot *polkadotClient.Client
}

// nolint
func (client *Client) build(o *common.ClientOptions) error {
	autorest.Count429AsRetry = false
	client.Features = o.Features
	client.Features.PolkadotFailOverFeature.DeleteVmsWithAPIInSingleMode = o.DeleteVmsWithAPIInSingleMode
	client.Polkadot = polkadotClient.NewClient(o)
	return nil
}
