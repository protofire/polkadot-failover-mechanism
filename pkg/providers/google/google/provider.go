package google

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protofire/polkadot-failover-mechanism/pkg/providers/google/version"

	googleoauth "golang.org/x/oauth2/google"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	return provider(false)
}

func TestProvider() *schema.Provider {
	return provider(true)
}

func provider(testing bool) *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CREDENTIALS",
					"GOOGLE_CLOUD_KEYFILE_JSON",
					"GCLOUD_KEYFILE_JSON",
				}, nil),
				ValidateDiagFunc: validate.DiagFunc(validateCredentials),
			},

			"access_token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_OAUTH_ACCESS_TOKEN",
				}, nil),
				ConflictsWith: []string{"credentials"},
			},
			"impersonate_service_account": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_IMPERSONATE_SERVICE_ACCOUNT",
				}, nil),
			},

			"impersonate_service_account_delegates": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"project": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_PROJECT",
					"GOOGLE_CLOUD_PROJECT",
					"GCLOUD_PROJECT",
					"CLOUDSDK_CORE_PROJECT",
				}, nil),
			},

			"billing_project": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_BILLING_PROJECT",
				}, nil),
			},

			"region": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_REGION",
					"GCLOUD_REGION",
					"CLOUDSDK_COMPUTE_REGION",
				}, nil),
			},

			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_ZONE",
					"GCLOUD_ZONE",
					"CLOUDSDK_COMPUTE_ZONE",
				}, nil),
			},

			"scopes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"delete_vms_with_api_in_single_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("POLKADOT_FAILOVER_DELETE_VMS_WITH_API", false),
				Description: "Delete vms in single mode with API call preserving current active validator",
				Default:     true,
			},

			"batching": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"send_after": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "10s",
							ValidateDiagFunc: validateNonNegativeDuration(),
						},
						"enable_batching": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},

			"user_project_override": {
				Type:     schema.TypeBool,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"USER_PROJECT_OVERRIDE",
				}, nil),
			},

			"request_timeout": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Generated Products
			"access_approval_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_ACCESS_APPROVAL_CUSTOM_ENDPOINT",
				}, AccessApprovalDefaultBasePath),
			},
			"access_context_manager_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_ACCESS_CONTEXT_MANAGER_CUSTOM_ENDPOINT",
				}, AccessContextManagerDefaultBasePath),
			},
			"active_directory_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_ACTIVE_DIRECTORY_CUSTOM_ENDPOINT",
				}, ActiveDirectoryDefaultBasePath),
			},
			"app_engine_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_APP_ENGINE_CUSTOM_ENDPOINT",
				}, AppEngineDefaultBasePath),
			},
			"big_query_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_BIG_QUERY_CUSTOM_ENDPOINT",
				}, BigQueryDefaultBasePath),
			},
			"bigquery_data_transfer_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_BIGQUERY_DATA_TRANSFER_CUSTOM_ENDPOINT",
				}, BigqueryDataTransferDefaultBasePath),
			},
			"bigtable_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_BIGTABLE_CUSTOM_ENDPOINT",
				}, BigtableDefaultBasePath),
			},
			"binary_authorization_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_BINARY_AUTHORIZATION_CUSTOM_ENDPOINT",
				}, BinaryAuthorizationDefaultBasePath),
			},
			"cloud_asset_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_ASSET_CUSTOM_ENDPOINT",
				}, CloudAssetDefaultBasePath),
			},
			"cloud_build_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_BUILD_CUSTOM_ENDPOINT",
				}, CloudBuildDefaultBasePath),
			},
			"cloud_functions_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_FUNCTIONS_CUSTOM_ENDPOINT",
				}, CloudFunctionsDefaultBasePath),
			},
			"cloud_iot_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_IOT_CUSTOM_ENDPOINT",
				}, CloudIotDefaultBasePath),
			},
			"cloud_run_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_RUN_CUSTOM_ENDPOINT",
				}, CloudRunDefaultBasePath),
			},
			"cloud_scheduler_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_SCHEDULER_CUSTOM_ENDPOINT",
				}, CloudSchedulerDefaultBasePath),
			},
			"cloud_tasks_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CLOUD_TASKS_CUSTOM_ENDPOINT",
				}, CloudTasksDefaultBasePath),
			},
			"compute_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_COMPUTE_CUSTOM_ENDPOINT",
				}, ComputeDefaultBasePath),
			},
			"container_analysis_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_CONTAINER_ANALYSIS_CUSTOM_ENDPOINT",
				}, ContainerAnalysisDefaultBasePath),
			},
			"data_catalog_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DATA_CATALOG_CUSTOM_ENDPOINT",
				}, DataCatalogDefaultBasePath),
			},
			"data_loss_prevention_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DATA_LOSS_PREVENTION_CUSTOM_ENDPOINT",
				}, DataLossPreventionDefaultBasePath),
			},
			"dataproc_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DATAPROC_CUSTOM_ENDPOINT",
				}, DataprocDefaultBasePath),
			},
			"datastore_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DATASTORE_CUSTOM_ENDPOINT",
				}, DatastoreDefaultBasePath),
			},
			"deployment_manager_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DEPLOYMENT_MANAGER_CUSTOM_ENDPOINT",
				}, DeploymentManagerDefaultBasePath),
			},
			"dialogflow_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DIALOGFLOW_CUSTOM_ENDPOINT",
				}, DialogflowDefaultBasePath),
			},
			"dns_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_DNS_CUSTOM_ENDPOINT",
				}, DNSDefaultBasePath),
			},
			"filestore_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_FILESTORE_CUSTOM_ENDPOINT",
				}, FilestoreDefaultBasePath),
			},
			"firestore_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_FIRESTORE_CUSTOM_ENDPOINT",
				}, FirestoreDefaultBasePath),
			},
			"game_services_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_GAME_SERVICES_CUSTOM_ENDPOINT",
				}, GameServicesDefaultBasePath),
			},
			"healthcare_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_HEALTHCARE_CUSTOM_ENDPOINT",
				}, HealthcareDefaultBasePath),
			},
			"iap_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_IAP_CUSTOM_ENDPOINT",
				}, IapDefaultBasePath),
			},
			"identity_platform_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_IDENTITY_PLATFORM_CUSTOM_ENDPOINT",
				}, IdentityPlatformDefaultBasePath),
			},
			"kms_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_KMS_CUSTOM_ENDPOINT",
				}, KMSDefaultBasePath),
			},
			"logging_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_LOGGING_CUSTOM_ENDPOINT",
				}, LoggingDefaultBasePath),
			},
			"ml_engine_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_ML_ENGINE_CUSTOM_ENDPOINT",
				}, MLEngineDefaultBasePath),
			},
			"monitoring_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_MONITORING_CUSTOM_ENDPOINT",
				}, MonitoringDefaultBasePath),
			},
			"network_management_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_NETWORK_MANAGEMENT_CUSTOM_ENDPOINT",
				}, NetworkManagementDefaultBasePath),
			},
			"os_config_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_OS_CONFIG_CUSTOM_ENDPOINT",
				}, OSConfigDefaultBasePath),
			},
			"os_login_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_OS_LOGIN_CUSTOM_ENDPOINT",
				}, OSLoginDefaultBasePath),
			},
			"pubsub_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_PUBSUB_CUSTOM_ENDPOINT",
				}, PubsubDefaultBasePath),
			},
			"redis_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_REDIS_CUSTOM_ENDPOINT",
				}, RedisDefaultBasePath),
			},
			"resource_manager_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_RESOURCE_MANAGER_CUSTOM_ENDPOINT",
				}, ResourceManagerDefaultBasePath),
			},
			"runtime_config_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_RUNTIME_CONFIG_CUSTOM_ENDPOINT",
				}, RuntimeConfigDefaultBasePath),
			},
			"secret_manager_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SECRET_MANAGER_CUSTOM_ENDPOINT",
				}, SecretManagerDefaultBasePath),
			},
			"security_center_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SECURITY_CENTER_CUSTOM_ENDPOINT",
				}, SecurityCenterDefaultBasePath),
			},
			"service_management_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SERVICE_MANAGEMENT_CUSTOM_ENDPOINT",
				}, ServiceManagementDefaultBasePath),
			},
			"service_usage_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SERVICE_USAGE_CUSTOM_ENDPOINT",
				}, ServiceUsageDefaultBasePath),
			},
			"source_repo_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SOURCE_REPO_CUSTOM_ENDPOINT",
				}, SourceRepoDefaultBasePath),
			},
			"spanner_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SPANNER_CUSTOM_ENDPOINT",
				}, SpannerDefaultBasePath),
			},
			"sql_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_SQL_CUSTOM_ENDPOINT",
				}, SQLDefaultBasePath),
			},
			"storage_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_STORAGE_CUSTOM_ENDPOINT",
				}, StorageDefaultBasePath),
			},
			"tpu_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_TPU_CUSTOM_ENDPOINT",
				}, TPUDefaultBasePath),
			},
			"vpc_access_custom_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateCustomEndpoint(),
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"GOOGLE_VPC_ACCESS_CUSTOM_ENDPOINT",
				}, VPCAccessDefaultBasePath),
			},

			// Handwritten Products / Versioned / Atypical Entries
			CloudBillingCustomEndpointEntryKey:           CloudBillingCustomEndpointEntry,
			ComposerCustomEndpointEntryKey:               ComposerCustomEndpointEntry,
			ComputeBetaCustomEndpointEntryKey:            ComputeBetaCustomEndpointEntry,
			ContainerCustomEndpointEntryKey:              ContainerCustomEndpointEntry,
			ContainerBetaCustomEndpointEntryKey:          ContainerBetaCustomEndpointEntry,
			DataprocBetaCustomEndpointEntryKey:           DataprocBetaCustomEndpointEntry,
			DataflowCustomEndpointEntryKey:               DataflowCustomEndpointEntry,
			DnsBetaCustomEndpointEntryKey:                DnsBetaCustomEndpointEntry,
			IamCredentialsCustomEndpointEntryKey:         IamCredentialsCustomEndpointEntry,
			ResourceManagerV2Beta1CustomEndpointEntryKey: ResourceManagerV2Beta1CustomEndpointEntry,
			RuntimeConfigCustomEndpointEntryKey:          RuntimeConfigCustomEndpointEntry,
			IAMCustomEndpointEntryKey:                    IAMCustomEndpointEntry,
			ServiceNetworkingCustomEndpointEntryKey:      ServiceNetworkingCustomEndpointEntry,
			ServiceUsageCustomEndpointEntryKey:           ServiceUsageCustomEndpointEntry,
			StorageTransferCustomEndpointEntryKey:        StorageTransferCustomEndpointEntry,
			BigtableAdminCustomEndpointEntryKey:          BigtableAdminCustomEndpointEntry,
		},

		ProviderMetaSchema: map[string]*schema.Schema{
			"module_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"polkadot_failover": resourcePolkadotFailover(),
		},

		DataSourcesMap: ResourceMap(),
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(ctx, d, provider, testing)
	}

	return provider
}

func ResourceMap() map[string]*schema.Resource {
	resourceMap, _ := ResourceMapWithErrors()
	return resourceMap
}

func ResourceMapWithErrors() (map[string]*schema.Resource, error) {
	return mergeResourceMaps()
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, p *schema.Provider, testing bool) (interface{}, diag.Diagnostics) {
	config := Config{
		Project:                      d.Get("project").(string),
		Region:                       d.Get("region").(string),
		Zone:                         d.Get("zone").(string),
		UserProjectOverride:          d.Get("user_project_override").(bool),
		BillingProject:               d.Get("billing_project").(string),
		DeleteVmsWithAPIInSingleMode: d.Get("delete_vms_with_api_in_single_mode").(bool),
		userAgent:                    p.UserAgent("terraform-provider-google", version.ProviderVersion),
	}

	if v, ok := d.GetOk("request_timeout"); ok {
		var err error
		config.RequestTimeout, err = time.ParseDuration(v.(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}
	// Add credential source
	if v, ok := d.GetOk("access_token"); ok {
		config.AccessToken = v.(string)
	} else if v, ok := d.GetOk("credentials"); ok {
		config.Credentials = v.(string)
	}
	if v, ok := d.GetOk("impersonate_service_account"); ok {
		config.ImpersonateServiceAccount = v.(string)
	}

	scopes := d.Get("scopes").([]interface{})
	if len(scopes) > 0 {
		config.Scopes = make([]string, len(scopes))
	}
	for i, scope := range scopes {
		config.Scopes[i] = scope.(string)
	}

	delegates := d.Get("impersonate_service_account_delegates").([]interface{})
	if len(delegates) > 0 {
		config.ImpersonateServiceAccountDelegates = make([]string, len(delegates))
	}
	for i, delegate := range delegates {
		config.ImpersonateServiceAccountDelegates[i] = delegate.(string)
	}

	batchCfg, err := expandProviderBatchingConfig(d.Get("batching"))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	config.BatchingConfig = batchCfg

	// Generated products
	config.AccessApprovalBasePath = d.Get("access_approval_custom_endpoint").(string)
	config.AccessContextManagerBasePath = d.Get("access_context_manager_custom_endpoint").(string)
	config.ActiveDirectoryBasePath = d.Get("active_directory_custom_endpoint").(string)
	config.AppEngineBasePath = d.Get("app_engine_custom_endpoint").(string)
	config.BigQueryBasePath = d.Get("big_query_custom_endpoint").(string)
	config.BigqueryDataTransferBasePath = d.Get("bigquery_data_transfer_custom_endpoint").(string)
	config.BigtableBasePath = d.Get("bigtable_custom_endpoint").(string)
	config.BinaryAuthorizationBasePath = d.Get("binary_authorization_custom_endpoint").(string)
	config.CloudAssetBasePath = d.Get("cloud_asset_custom_endpoint").(string)
	config.CloudBuildBasePath = d.Get("cloud_build_custom_endpoint").(string)
	config.CloudFunctionsBasePath = d.Get("cloud_functions_custom_endpoint").(string)
	config.CloudIotBasePath = d.Get("cloud_iot_custom_endpoint").(string)
	config.CloudRunBasePath = d.Get("cloud_run_custom_endpoint").(string)
	config.CloudSchedulerBasePath = d.Get("cloud_scheduler_custom_endpoint").(string)
	config.CloudTasksBasePath = d.Get("cloud_tasks_custom_endpoint").(string)
	config.ComputeBasePath = d.Get("compute_custom_endpoint").(string)
	config.ContainerAnalysisBasePath = d.Get("container_analysis_custom_endpoint").(string)
	config.DataCatalogBasePath = d.Get("data_catalog_custom_endpoint").(string)
	config.DataLossPreventionBasePath = d.Get("data_loss_prevention_custom_endpoint").(string)
	config.DataprocBasePath = d.Get("dataproc_custom_endpoint").(string)
	config.DatastoreBasePath = d.Get("datastore_custom_endpoint").(string)
	config.DeploymentManagerBasePath = d.Get("deployment_manager_custom_endpoint").(string)
	config.DialogflowBasePath = d.Get("dialogflow_custom_endpoint").(string)
	config.DNSBasePath = d.Get("dns_custom_endpoint").(string)
	config.FilestoreBasePath = d.Get("filestore_custom_endpoint").(string)
	config.FirestoreBasePath = d.Get("firestore_custom_endpoint").(string)
	config.GameServicesBasePath = d.Get("game_services_custom_endpoint").(string)
	config.HealthcareBasePath = d.Get("healthcare_custom_endpoint").(string)
	config.IapBasePath = d.Get("iap_custom_endpoint").(string)
	config.IdentityPlatformBasePath = d.Get("identity_platform_custom_endpoint").(string)
	config.KMSBasePath = d.Get("kms_custom_endpoint").(string)
	config.LoggingBasePath = d.Get("logging_custom_endpoint").(string)
	config.MLEngineBasePath = d.Get("ml_engine_custom_endpoint").(string)
	config.MonitoringBasePath = d.Get("monitoring_custom_endpoint").(string)
	config.NetworkManagementBasePath = d.Get("network_management_custom_endpoint").(string)
	config.OSConfigBasePath = d.Get("os_config_custom_endpoint").(string)
	config.OSLoginBasePath = d.Get("os_login_custom_endpoint").(string)
	config.PubsubBasePath = d.Get("pubsub_custom_endpoint").(string)
	config.RedisBasePath = d.Get("redis_custom_endpoint").(string)
	config.ResourceManagerBasePath = d.Get("resource_manager_custom_endpoint").(string)
	config.RuntimeConfigBasePath = d.Get("runtime_config_custom_endpoint").(string)
	config.SecretManagerBasePath = d.Get("secret_manager_custom_endpoint").(string)
	config.SecurityCenterBasePath = d.Get("security_center_custom_endpoint").(string)
	config.ServiceManagementBasePath = d.Get("service_management_custom_endpoint").(string)
	config.ServiceUsageBasePath = d.Get("service_usage_custom_endpoint").(string)
	config.SourceRepoBasePath = d.Get("source_repo_custom_endpoint").(string)
	config.SpannerBasePath = d.Get("spanner_custom_endpoint").(string)
	config.SQLBasePath = d.Get("sql_custom_endpoint").(string)
	config.StorageBasePath = d.Get("storage_custom_endpoint").(string)
	config.TPUBasePath = d.Get("tpu_custom_endpoint").(string)
	config.VPCAccessBasePath = d.Get("vpc_access_custom_endpoint").(string)

	// Handwritten Products / Versioned / Atypical Entries

	config.CloudBillingBasePath = d.Get(CloudBillingCustomEndpointEntryKey).(string)
	config.ComposerBasePath = d.Get(ComposerCustomEndpointEntryKey).(string)
	config.ComputeBetaBasePath = d.Get(ComputeBetaCustomEndpointEntryKey).(string)
	config.ContainerBasePath = d.Get(ContainerCustomEndpointEntryKey).(string)
	config.ContainerBetaBasePath = d.Get(ContainerBetaCustomEndpointEntryKey).(string)
	config.DataprocBetaBasePath = d.Get(DataprocBetaCustomEndpointEntryKey).(string)
	config.DataflowBasePath = d.Get(DataflowCustomEndpointEntryKey).(string)
	config.DnsBetaBasePath = d.Get(DnsBetaCustomEndpointEntryKey).(string)
	config.IamCredentialsBasePath = d.Get(IamCredentialsCustomEndpointEntryKey).(string)
	config.ResourceManagerV2Beta1BasePath = d.Get(ResourceManagerV2Beta1CustomEndpointEntryKey).(string)
	config.RuntimeConfigBasePath = d.Get(RuntimeConfigCustomEndpointEntryKey).(string)
	config.IAMBasePath = d.Get(IAMCustomEndpointEntryKey).(string)
	config.ServiceNetworkingBasePath = d.Get(ServiceNetworkingCustomEndpointEntryKey).(string)
	config.ServiceUsageBasePath = d.Get(ServiceUsageCustomEndpointEntryKey).(string)
	config.StorageTransferBasePath = d.Get(StorageTransferCustomEndpointEntryKey).(string)
	config.BigtableAdminBasePath = d.Get(BigtableAdminCustomEndpointEntryKey).(string)

	if !testing {
		if err := config.LoadAndValidate(ctx); err != nil {
			return nil, diag.FromErr(err)
		}
	}

	return &config, nil
}

func validateCredentials(v interface{}, _ string) (warnings []string, errors []error) {
	if v == nil || v.(string) == "" {
		return
	}
	creds := v.(string)
	// if this is a path and we can stat it, assume it's ok
	if _, err := os.Stat(creds); err == nil {
		return
	}
	if _, err := googleoauth.CredentialsFromJSON(context.Background(), []byte(creds)); err != nil {
		errors = append(errors,
			fmt.Errorf("JSON credentials in %q are not valid: %s", creds, err))
	}

	return
}
