package clients

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

type ResourceManagerAccount struct {
	AuthenticatedAsAServicePrincipal bool
	ClientID                         string
	Environment                      azure.Environment
	ObjectID                         string
	SubscriptionID                   string
	TenantID                         string
}

func NewResourceManagerAccount(ctx context.Context, config authentication.Config, env azure.Environment) (*ResourceManagerAccount, error) {
	objectID := ""

	// TODO remove this when we confirm that MSI no longer returns nil with getAuthenticatedObjectID
	if getAuthenticatedObjectID := config.GetAuthenticatedObjectID; getAuthenticatedObjectID != nil {
		v, err := getAuthenticatedObjectID(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting authenticated object ID: %v", err)
		}
		objectID = v
	}

	account := ResourceManagerAccount{
		AuthenticatedAsAServicePrincipal: config.AuthenticatedAsAServicePrincipal,
		ClientID:                         config.ClientID,
		Environment:                      env,
		ObjectID:                         objectID,
		TenantID:                         config.TenantID,
		SubscriptionID:                   config.SubscriptionID,
	}
	return &account, nil
}
