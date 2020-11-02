package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/common"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/features"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/location"
)

//ClientBuilder builder
type ClientBuilder struct {
	AuthConfig                   *authentication.Config
	DisableCorrelationRequestID  bool
	DisableTerraformPartnerID    bool
	PartnerID                    string
	SkipProviderRegistration     bool
	StorageUseAzureAD            bool
	TerraformVersion             string
	Features                     features.UserFeatures
	DeleteVmsWithAPIInSingleMode bool
}

const azureStackEnvironmentError = `
The AzureRM Provider supports the different Azure Public Clouds - including China, Germany,
Public and US Government - however it does not support Azure Stack due to differences in
API and feature availability.

Terraform instead offers a separate "azurestack" provider which supports the functionality
and API's available in Azure Stack via Azure Stack Profiles.
`

// Build builds Client
func Build(ctx context.Context, builder ClientBuilder) (*Client, error) {
	// point folks towards the separate Azure Stack Provider when using Azure Stack
	if strings.EqualFold(builder.AuthConfig.Environment, "AZURESTACKCLOUD") {
		return nil, fmt.Errorf(azureStackEnvironmentError)
	}

	isAzureStack, err := authentication.IsEnvironmentAzureStack(ctx, builder.AuthConfig.MetadataHost, builder.AuthConfig.Environment)
	if err != nil {
		return nil, err
	}
	if isAzureStack {
		return nil, fmt.Errorf(azureStackEnvironmentError)
	}

	env, err := authentication.AzureEnvironmentByNameFromEndpoint(ctx, builder.AuthConfig.MetadataHost, builder.AuthConfig.Environment)
	if err != nil {
		return nil, err
	}

	if features.EnhancedValidationEnabled() {
		location.CacheSupportedLocations(ctx, env)
	}

	// client declarations:
	account, err := NewResourceManagerAccount(ctx, *builder.AuthConfig, *env)
	if err != nil {
		return nil, fmt.Errorf("Error building account: %+v", err)
	}

	client := Client{
		Account: account,
	}

	oauthConfig, err := builder.AuthConfig.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}

	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, fmt.Errorf("Unable to configure OAuthConfig for tenant %s", builder.AuthConfig.TenantID)
	}

	senderImpl := sender.BuildSender("AzureRM")

	// Resource Manager endpoints
	endpoint := env.ResourceManagerEndpoint
	auth, err := builder.AuthConfig.GetAuthorizationToken(senderImpl, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, err
	}

	// Graph Endpoints
	graphEndpoint := env.GraphEndpoint
	graphAuth, err := builder.AuthConfig.GetAuthorizationToken(senderImpl, oauthConfig, graphEndpoint)
	if err != nil {
		return nil, err
	}

	// Storage Endpoints
	storageAuth, err := builder.AuthConfig.GetAuthorizationToken(senderImpl, oauthConfig, env.ResourceIdentifiers.Storage)
	if err != nil {
		return nil, err
	}

	// Key Vault Endpoints
	keyVaultAuth := builder.AuthConfig.BearerAuthorizerCallback(senderImpl, oauthConfig)

	o := &common.ClientOptions{
		SubscriptionID:               builder.AuthConfig.SubscriptionID,
		TenantID:                     builder.AuthConfig.TenantID,
		PartnerID:                    builder.PartnerID,
		TerraformVersion:             builder.TerraformVersion,
		GraphAuthorizer:              graphAuth,
		GraphEndpoint:                graphEndpoint,
		KeyVaultAuthorizer:           keyVaultAuth,
		ResourceManagerAuthorizer:    auth,
		ResourceManagerEndpoint:      endpoint,
		StorageAuthorizer:            storageAuth,
		SkipProviderReg:              builder.SkipProviderRegistration,
		DisableCorrelationRequestID:  builder.DisableCorrelationRequestID,
		DisableTerraformPartnerID:    builder.DisableTerraformPartnerID,
		Environment:                  *env,
		Features:                     builder.Features,
		StorageUseAzureAD:            builder.StorageUseAzureAD,
		DeleteVmsWithAPIInSingleMode: builder.DeleteVmsWithAPIInSingleMode,
	}

	if err := client.build(o); err != nil {
		return nil, fmt.Errorf("error building Client: %+v", err)
	}

	return &client, nil
}
