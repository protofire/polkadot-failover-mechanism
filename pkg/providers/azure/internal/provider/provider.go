package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	polkadotClient "github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/services/polkadot/client"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/internal/clients"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/azure/utils"
)

// AzurePolkadotProvider constructor
func AzurePolkadotProvider() *schema.Provider {
	return azurePolkadotProvider(false)
}

// TestAzurePolkadotProvider constructor
func TestAzurePolkadotProvider() *schema.Provider {
	return azurePolkadotProvider(true)
}

func azurePolkadotProvider(testProvider bool) *schema.Provider {
	// avoids this showing up in test output
	var debugLog = func(f string, v ...interface{}) {
		if os.Getenv("TF_LOG") == "" {
			return
		}

		if os.Getenv("TF_ACC") != "" {
			return
		}

		log.Printf(f, v...)
	}

	dataSources := make(map[string]*schema.Resource)
	resources := make(map[string]*schema.Resource)
	for _, service := range SupportedServices() {
		debugLog("[DEBUG] Registering Data Sources for %q..", service.Name())
		for k, v := range service.SupportedDataSources() {
			if existing := dataSources[k]; existing != nil {
				panic(fmt.Sprintf("An existing Data Source exists for %q", k))
			}

			dataSources[k] = v
		}

		debugLog("[DEBUG] Registering Resources for %q..", service.Name())
		for k, v := range service.SupportedResources() {
			if existing := resources[k]; existing != nil {
				panic(fmt.Sprintf("An existing Resource exists for %q", k))
			}

			resources[k] = v
		}
	}

	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"subscription_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
				Description: "The Subscription ID which should be used.",
			},

			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
				Description: "The Client ID which should be used.",
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
				Description: "The Tenant ID which should be used.",
			},

			"auxiliary_tenant_ids": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
				Description: "The Cloud Environment which should be used. Possible values are public, usgovernment, german, and china. Defaults to public.",
			},

			"metadata_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_METADATA_HOSTNAME", ""),
				Description: "The Hostname which should be used for the Azure Metadata Service.",
			},

			// Client Certificate specific fields
			"client_certificate_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PATH", ""),
				Description: "The path to the Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate.",
			},

			"client_certificate_password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PASSWORD", ""),
				Description: "The password associated with the Client Certificate. For use when authenticating as a Service Principal using a Client Certificate",
			},

			// Client Secret specific fields
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
				Description: "The Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret.",
			},

			// Managed Service Identity specific fields
			"use_msi": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", false),
				Description: "Allowed Managed Service Identity be used for Authentication.",
			},
			"msi_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_MSI_ENDPOINT", ""),
				Description: "The path to a custom endpoint for Managed Service Identity - in most circumstances this should be detected automatically. ",
			},

			// Managed Tracking GUID for User-agent
			"partner_id": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validate.DiagFunc(validation.Any(validation.IsUUID, validation.StringIsEmpty)),
				DefaultFunc:      schema.EnvDefaultFunc("ARM_PARTNER_ID", ""),
				Description:      "A GUID/UUID that is registered with Microsoft to facilitate partner resource usage attribution.",
			},

			"disable_correlation_request_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_DISABLE_CORRELATION_REQUEST_ID", false),
				Description: "This will disable the x-ms-correlation-request-id header.",
			},

			"disable_terraform_partner_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_DISABLE_TERRAFORM_PARTNER_ID", false),
				Description: "This will disable the Terraform Partner ID which is used if a custom `partner_id` isn't specified.",
			},

			"features": schemaFeatures(),

			// Advanced feature flags
			"skip_credentials_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_CREDENTIALS_VALIDATION", false),
				Description: "This will cause the AzureRM Provider to skip verifying the credentials being used are valid.",
			},

			"skip_provider_registration": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SKIP_PROVIDER_REGISTRATION", false),
				Description: "Should the AzureRM Provider skip registering all of the Resource Providers that it supports, if they're not already registered?",
			},

			"storage_use_azuread": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_STORAGE_USE_AZUREAD", false),
				Description: "Should the AzureRM Provider use AzureAD to access the Storage Data Plane Apis?",
			},

			"delete_vms_with_api_in_single_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("POLKADOT_FAILOVER_DELETE_VMS_WITH_API", false),
				Description: "Delete vms in single mode with API call preserving current active validator",
			},
		},

		DataSourcesMap: dataSources,
		ResourcesMap:   resources,
	}

	if testProvider {
		p.ConfigureContextFunc = providerTestConfigure(p)
	} else {
		p.ConfigureContextFunc = providerConfigure(p)
	}

	return p
}

func providerTestConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		client := clients.Client{
			Account:  nil,
			Features: expandFeatures(d.Get("features").([]interface{})),
			Polkadot: &polkadotClient.Client{},
		}
		client.Features.PolkadotFailOverFeature.DeleteVmsWithAPIInSingleMode = d.Get("delete_vms_with_api_in_single_mode").(bool)
		return &client, nil
	}
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var auxTenants []string
		if v, ok := d.Get("auxiliary_tenant_ids").([]interface{}); ok && len(v) > 0 {
			auxTenants = *utils.ExpandStringSlice(v)
		} else if v := os.Getenv("ARM_AUXILIARY_TENANT_IDS"); v != "" {
			auxTenants = strings.Split(v, ";")
		}

		if len(auxTenants) > 3 {
			return nil, diag.Errorf("the provider only supports 3 auxiliary tenant IDs")
		}

		metadataHost := d.Get("metadata_host").(string)

		builder := &authentication.Builder{
			SubscriptionID:     d.Get("subscription_id").(string),
			ClientID:           d.Get("client_id").(string),
			ClientSecret:       d.Get("client_secret").(string),
			TenantID:           d.Get("tenant_id").(string),
			AuxiliaryTenantIDs: auxTenants,
			Environment:        d.Get("environment").(string),
			MetadataURL:        metadataHost, // TODO: rename this in Helpers too
			MsiEndpoint:        d.Get("msi_endpoint").(string),
			ClientCertPassword: d.Get("client_certificate_password").(string),
			ClientCertPath:     d.Get("client_certificate_path").(string),

			// Feature Toggles
			SupportsClientCertAuth:         true,
			SupportsClientSecretAuth:       true,
			SupportsManagedServiceIdentity: d.Get("use_msi").(bool),
			SupportsAzureCliToken:          true,
			SupportsAuxiliaryTenants:       len(auxTenants) > 0,

			// Doc Links
			ClientSecretDocsLink: "https://www.terraform.io/docs/providers/azurerm/guides/service_principal_client_secret.html",
		}

		config, err := builder.Build()
		if err != nil {
			return nil, diag.Errorf("Error building AzureRM Client: %+v", err)
		}

		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}

		skipProviderRegistration := d.Get("skip_provider_registration").(bool)
		clientBuilder := clients.ClientBuilder{
			AuthConfig:                   config,
			SkipProviderRegistration:     skipProviderRegistration,
			TerraformVersion:             terraformVersion,
			PartnerID:                    d.Get("partner_id").(string),
			DisableCorrelationRequestID:  d.Get("disable_correlation_request_id").(bool),
			DisableTerraformPartnerID:    d.Get("disable_terraform_partner_id").(bool),
			Features:                     expandFeatures(d.Get("features").([]interface{})),
			StorageUseAzureAD:            d.Get("storage_use_azuread").(bool),
			DeleteVmsWithAPIInSingleMode: d.Get("delete_vms_with_api_in_single_mode").(bool),
		}

		client, err := clients.Build(ctx, clientBuilder)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return client, nil
	}
}
