package provider

import (
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/services/common"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/services/polkadot"
)

// SupportedServices list all cloud services
func SupportedServices() []common.ServiceRegistration {
	return []common.ServiceRegistration{
		polkadot.Registration{},
	}
}
