package google

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/transport"
)

type providerMeta struct {
	ModuleName string `cty:"module_name"`
}

// Config is the configuration structure used to instantiate the Google
// provider.
type Config struct {
	AccessToken                        string
	Credentials                        string
	ImpersonateServiceAccount          string
	ImpersonateServiceAccountDelegates []string
	Project                            string
	Region                             string
	BillingProject                     string
	Zone                               string
	Scopes                             []string
	BatchingConfig                     *batchingConfig
	UserProjectOverride                bool
	RequestTimeout                     time.Duration
	// PollInterval is passed to resource.StateChangeConf in common_operation.go
	// It controls the interval at which we poll for successful operations
	PollInterval time.Duration

	client    *http.Client
	context   context.Context
	userAgent string

	tokenSource oauth2.TokenSource

	AccessApprovalBasePath       string
	AccessContextManagerBasePath string
	ActiveDirectoryBasePath      string
	AppEngineBasePath            string
	BigQueryBasePath             string
	BigqueryDataTransferBasePath string
	BigtableBasePath             string
	BinaryAuthorizationBasePath  string
	CloudAssetBasePath           string
	CloudBuildBasePath           string
	CloudFunctionsBasePath       string
	CloudIotBasePath             string
	CloudRunBasePath             string
	CloudSchedulerBasePath       string
	CloudTasksBasePath           string
	ComputeBasePath              string
	ContainerAnalysisBasePath    string
	DataCatalogBasePath          string
	DataLossPreventionBasePath   string
	DataprocBasePath             string
	DatastoreBasePath            string
	DeploymentManagerBasePath    string
	DialogflowBasePath           string
	DNSBasePath                  string
	FilestoreBasePath            string
	FirestoreBasePath            string
	GameServicesBasePath         string
	HealthcareBasePath           string
	IapBasePath                  string
	IdentityPlatformBasePath     string
	KMSBasePath                  string
	LoggingBasePath              string
	MLEngineBasePath             string
	MonitoringBasePath           string
	NetworkManagementBasePath    string
	OSConfigBasePath             string
	OSLoginBasePath              string
	PubsubBasePath               string
	RedisBasePath                string
	ResourceManagerBasePath      string
	RuntimeConfigBasePath        string
	SecretManagerBasePath        string
	SecurityCenterBasePath       string
	ServiceManagementBasePath    string
	ServiceUsageBasePath         string
	SourceRepoBasePath           string
	SpannerBasePath              string
	SQLBasePath                  string
	StorageBasePath              string
	TPUBasePath                  string
	VPCAccessBasePath            string

	CloudBillingBasePath  string
	ComposerBasePath      string
	ComputeBetaBasePath   string
	ContainerBasePath     string
	ContainerBetaBasePath string
	DataprocBetaBasePath  string
	DataflowBasePath      string
	// nolint
	DnsBetaBasePath                string
	IamCredentialsBasePath         string
	ResourceManagerV2Beta1BasePath string
	IAMBasePath                    string
	CloudIoTBasePath               string
	ServiceNetworkingBasePath      string
	StorageTransferBasePath        string
	BigtableAdminBasePath          string

	requestBatcherServiceUsage   *RequestBatcher
	requestBatcherIam            *RequestBatcher
	DeleteVmsWithAPIInSingleMode bool
}

// Generated product base paths
// nolint
var AccessApprovalDefaultBasePath = "https://accessapproval.googleapis.com/v1/"
var AccessContextManagerDefaultBasePath = "https://accesscontextmanager.googleapis.com/v1/"
var ActiveDirectoryDefaultBasePath = "https://managedidentities.googleapis.com/v1/"
var AppEngineDefaultBasePath = "https://appengine.googleapis.com/v1/"
var BigQueryDefaultBasePath = "https://bigquery.googleapis.com/bigquery/v2/"
var BigqueryDataTransferDefaultBasePath = "https://bigquerydatatransfer.googleapis.com/v1/"
var BigtableDefaultBasePath = "https://bigtableadmin.googleapis.com/v2/"
var BinaryAuthorizationDefaultBasePath = "https://binaryauthorization.googleapis.com/v1/"
var CloudAssetDefaultBasePath = "https://cloudasset.googleapis.com/v1/"
var CloudBuildDefaultBasePath = "https://cloudbuild.googleapis.com/v1/"
var CloudFunctionsDefaultBasePath = "https://cloudfunctions.googleapis.com/v1/"
var CloudIotDefaultBasePath = "https://cloudiot.googleapis.com/v1/"
var CloudRunDefaultBasePath = "https://{{location}}-run.googleapis.com/"
var CloudSchedulerDefaultBasePath = "https://cloudscheduler.googleapis.com/v1/"
var CloudTasksDefaultBasePath = "https://cloudtasks.googleapis.com/v2/"
var ComputeDefaultBasePath = "https://compute.googleapis.com/compute/v1/"
var ContainerAnalysisDefaultBasePath = "https://containeranalysis.googleapis.com/v1/"
var DataCatalogDefaultBasePath = "https://datacatalog.googleapis.com/v1/"
var DataLossPreventionDefaultBasePath = "https://dlp.googleapis.com/v2/"
var DataprocDefaultBasePath = "https://dataproc.googleapis.com/v1/"
var DatastoreDefaultBasePath = "https://datastore.googleapis.com/v1/"
var DeploymentManagerDefaultBasePath = "https://www.googleapis.com/deploymentmanager/v2/"
var DialogflowDefaultBasePath = "https://dialogflow.googleapis.com/v2/"
var DNSDefaultBasePath = "https://dns.googleapis.com/dns/v1/"
var FilestoreDefaultBasePath = "https://file.googleapis.com/v1/"
var FirestoreDefaultBasePath = "https://firestore.googleapis.com/v1/"
var GameServicesDefaultBasePath = "https://gameservices.googleapis.com/v1/"
var HealthcareDefaultBasePath = "https://healthcare.googleapis.com/v1/"
var IapDefaultBasePath = "https://iap.googleapis.com/v1/"
var IdentityPlatformDefaultBasePath = "https://identitytoolkit.googleapis.com/v2/"
var KMSDefaultBasePath = "https://cloudkms.googleapis.com/v1/"
var LoggingDefaultBasePath = "https://logging.googleapis.com/v2/"
var MLEngineDefaultBasePath = "https://ml.googleapis.com/v1/"
var MonitoringDefaultBasePath = "https://monitoring.googleapis.com/"
var NetworkManagementDefaultBasePath = "https://networkmanagement.googleapis.com/v1/"
var OSConfigDefaultBasePath = "https://osconfig.googleapis.com/v1/"
var OSLoginDefaultBasePath = "https://oslogin.googleapis.com/v1/"
var PubsubDefaultBasePath = "https://pubsub.googleapis.com/v1/"
var RedisDefaultBasePath = "https://redis.googleapis.com/v1/"
var ResourceManagerDefaultBasePath = "https://cloudresourcemanager.googleapis.com/v1/"
var RuntimeConfigDefaultBasePath = "https://runtimeconfig.googleapis.com/v1beta1/"

// nolint
var SecretManagerDefaultBasePath = "https://secretmanager.googleapis.com/v1/"
var SecurityCenterDefaultBasePath = "https://securitycenter.googleapis.com/v1/"
var ServiceManagementDefaultBasePath = "https://servicemanagement.googleapis.com/v1/"
var ServiceUsageDefaultBasePath = "https://serviceusage.googleapis.com/v1/"
var SourceRepoDefaultBasePath = "https://sourcerepo.googleapis.com/v1/"
var SpannerDefaultBasePath = "https://spanner.googleapis.com/v1/"
var SQLDefaultBasePath = "https://sqladmin.googleapis.com/sql/v1beta4/"
var StorageDefaultBasePath = "https://storage.googleapis.com/storage/v1/"
var TPUDefaultBasePath = "https://tpu.googleapis.com/v1/"
var VPCAccessDefaultBasePath = "https://vpcaccess.googleapis.com/v1/"

var DefaultClientScopes = []string{
	"https://www.googleapis.com/auth/compute",
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/cloud-identity",
	"https://www.googleapis.com/auth/ndev.clouddns.readwrite",
	"https://www.googleapis.com/auth/devstorage.full_control",
	"https://www.googleapis.com/auth/userinfo.email",
}

func (c *Config) LoadAndValidate(ctx context.Context) error {
	if len(c.Scopes) == 0 {
		c.Scopes = DefaultClientScopes
	}

	c.context = ctx

	tokenSource, err := c.getTokenSource(c.Scopes)
	if err != nil {
		return err
	}
	c.tokenSource = tokenSource

	cleanCtx := context.WithValue(ctx, oauth2.HTTPClient, cleanhttp.DefaultClient())

	// 1. OAUTH2 TRANSPORT/CLIENT - sets up proper auth headers
	client := oauth2.NewClient(cleanCtx, tokenSource)

	// 2. Logging Transport - ensure we log HTTP requests to GCP APIs.
	loggingTransport := logging.NewTransport("Google", client.Transport)

	// 3. Retry Transport - retries common temporary errors
	// Keep order for wrapping logging so we log each retried request as well.
	// This value should be used if needed to create shallow copies with additional retry predicates.
	// See ClientWithAdditionalRetries
	retryTransport := NewTransportWithDefaultRetries(loggingTransport)

	// Set final transport value.
	client.Transport = retryTransport

	// This timeout is a timeout per HTTP request, not per logical operation.
	client.Timeout = c.synchronousTimeout()

	c.client = client
	c.context = ctx

	c.Region = GetRegionFromRegionSelfLink(c.Region)

	c.requestBatcherServiceUsage = NewRequestBatcher("Service Usage", ctx, c.BatchingConfig)
	c.requestBatcherIam = NewRequestBatcher("IAM", ctx, c.BatchingConfig)

	c.PollInterval = 10 * time.Second

	return nil
}

func expandProviderBatchingConfig(v interface{}) (*batchingConfig, error) {
	config := &batchingConfig{
		sendAfter:      time.Second * defaultBatchSendIntervalSec,
		enableBatching: true,
	}

	if v == nil {
		return config, nil
	}
	ls := v.([]interface{})
	if len(ls) == 0 || ls[0] == nil {
		return config, nil
	}

	cfgV := ls[0].(map[string]interface{})
	if sendAfterV, ok := cfgV["send_after"]; ok {
		sendAfter, err := time.ParseDuration(sendAfterV.(string))
		if err != nil {
			return nil, fmt.Errorf("unable to parse duration from 'send_after' value %q", sendAfterV)
		}
		config.sendAfter = sendAfter
	}

	if enable, ok := cfgV["enable_batching"]; ok {
		config.enableBatching = enable.(bool)
	}

	return config, nil
}

func (c *Config) synchronousTimeout() time.Duration {
	if c.RequestTimeout == 0 {
		return 30 * time.Second
	}
	return c.RequestTimeout
}

func (c *Config) getTokenSource(clientScopes []string) (oauth2.TokenSource, error) {
	creds, err := c.GetCredentials(clientScopes)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	return creds.TokenSource, nil
}

// Methods to create new services from config
// Some base paths below need the version and possibly more of the path
// set on them. The client libraries are inconsistent about which values they need;
// while most only want the host URL, some older ones also want the version and some
// of those "projects" as well. You can find out if this is required by looking at
// the basePath value in the client library file.
func (c *Config) NewComputeClient(userAgent string) *compute.Service {
	computeClientBasePath := c.ComputeBasePath + "projects/"
	log.Printf("[INFO] Instantiating GCE client for path %s", computeClientBasePath)
	clientCompute, err := compute.NewService(c.context, option.WithHTTPClient(c.client))
	if err != nil {
		log.Printf("[WARN] Error creating client compute: %s", err)
		return nil
	}
	clientCompute.UserAgent = userAgent
	clientCompute.BasePath = computeClientBasePath

	return clientCompute
}

func (c *Config) NewComputeBetaClient(userAgent string) *computeBeta.Service {
	computeBetaClientBasePath := c.ComputeBetaBasePath + "projects/"
	log.Printf("[INFO] Instantiating GCE Beta client for path %s", computeBetaClientBasePath)
	clientComputeBeta, err := computeBeta.NewService(c.context, option.WithHTTPClient(c.client))
	if err != nil {
		log.Printf("[WARN] Error creating client compute beta: %s", err)
		return nil
	}
	clientComputeBeta.UserAgent = userAgent
	clientComputeBeta.BasePath = computeBetaClientBasePath

	return clientComputeBeta
}

func (c *Config) NewMetricsClient(userAgent string) *monitoring.MetricClient {
	log.Print("[INFO] Instantiating GCP metrics client")
	ctx := context.Background()
	clientMetrics, err := monitoring.NewMetricClient(
		ctx,
		option.WithTokenSource(c.tokenSource),
		option.WithUserAgent(userAgent),
	)
	if err != nil {
		log.Printf("[ERROR] Error creating metrics client: %s", err)
		return nil
	}
	return clientMetrics
}

// staticTokenSource is used to be able to identify static token sources without reflection.
type staticTokenSource struct {
	oauth2.TokenSource
}

func (c *Config) GetCredentials(clientScopes []string) (googleoauth.Credentials, error) {

	if c.AccessToken != "" {
		contents, _, err := pathOrContents(c.AccessToken)
		if err != nil {
			return googleoauth.Credentials{}, fmt.Errorf("Error loading access token: %s", err)
		}
		token := &oauth2.Token{AccessToken: contents}

		if c.ImpersonateServiceAccount != "" {
			opts := []option.ClientOption{option.WithTokenSource(oauth2.StaticTokenSource(token)), option.ImpersonateCredentials(c.ImpersonateServiceAccount, c.ImpersonateServiceAccountDelegates...), option.WithScopes(clientScopes...)}
			creds, err := transport.Creds(context.TODO(), opts...)
			if err != nil {
				return googleoauth.Credentials{}, err
			}
			return *creds, nil
		}

		log.Printf("[INFO] Authenticating using configured Google JSON 'access_token'...")
		log.Printf("[INFO]   -- Scopes: %s", clientScopes)

		return googleoauth.Credentials{
			TokenSource: staticTokenSource{oauth2.StaticTokenSource(token)},
		}, nil
	}

	if c.Credentials != "" {
		contents, _, err := pathOrContents(c.Credentials)
		if err != nil {
			return googleoauth.Credentials{}, fmt.Errorf("error loading credentials: %s", err)
		}
		if c.ImpersonateServiceAccount != "" {
			opts := []option.ClientOption{option.WithCredentialsJSON([]byte(contents)), option.ImpersonateCredentials(c.ImpersonateServiceAccount, c.ImpersonateServiceAccountDelegates...), option.WithScopes(clientScopes...)}
			creds, err := transport.Creds(context.TODO(), opts...)
			if err != nil {
				return googleoauth.Credentials{}, err
			}
			return *creds, nil
		}
		creds, err := googleoauth.CredentialsFromJSON(c.context, []byte(contents), clientScopes...)
		if err != nil {
			return googleoauth.Credentials{}, fmt.Errorf("unable to parse credentials from '%s': %s", contents, err)
		}

		log.Printf("[INFO] Authenticating using configured Google JSON 'credentials'...")
		log.Printf("[INFO]   -- Scopes: %s", clientScopes)
		return *creds, nil
	}

	if c.ImpersonateServiceAccount != "" {
		opts := option.ImpersonateCredentials(c.ImpersonateServiceAccount, c.ImpersonateServiceAccountDelegates...)
		creds, err := transport.Creds(context.TODO(), opts, option.WithScopes(clientScopes...))
		if err != nil {
			return googleoauth.Credentials{}, err
		}
		return *creds, nil

	}

	log.Printf("[INFO] Authenticating using DefaultClient...")
	log.Printf("[INFO]   -- Scopes: %s", clientScopes)

	defaultTS, err := googleoauth.DefaultTokenSource(context.Background(), clientScopes...)
	if err != nil {
		return googleoauth.Credentials{}, fmt.Errorf("Attempted to load application default credentials since neither `credentials` nor `access_token` was set in the provider block.  No credentials loaded. To use your gcloud credentials, run 'gcloud auth application-default login'.  Original error: %w", err)
	}
	return googleoauth.Credentials{
		TokenSource: defaultTS,
	}, err
}

// For a consumer of config.go that isn't a full fledged provider and doesn't
// have its own endpoint mechanism such as sweepers, init {{service}}BasePath
// values to a default. After using this, you should call config.LoadAndValidate.
func ConfigureBasePaths(c *Config) {
	// Generated Products
	c.AccessApprovalBasePath = AccessApprovalDefaultBasePath
	c.AccessContextManagerBasePath = AccessContextManagerDefaultBasePath
	c.ActiveDirectoryBasePath = ActiveDirectoryDefaultBasePath
	c.AppEngineBasePath = AppEngineDefaultBasePath
	c.BigQueryBasePath = BigQueryDefaultBasePath
	c.BigqueryDataTransferBasePath = BigqueryDataTransferDefaultBasePath
	c.BigtableBasePath = BigtableDefaultBasePath
	c.BinaryAuthorizationBasePath = BinaryAuthorizationDefaultBasePath
	c.CloudAssetBasePath = CloudAssetDefaultBasePath
	c.CloudBuildBasePath = CloudBuildDefaultBasePath
	c.CloudFunctionsBasePath = CloudFunctionsDefaultBasePath
	c.CloudIotBasePath = CloudIotDefaultBasePath
	c.CloudRunBasePath = CloudRunDefaultBasePath
	c.CloudSchedulerBasePath = CloudSchedulerDefaultBasePath
	c.CloudTasksBasePath = CloudTasksDefaultBasePath
	c.ComputeBasePath = ComputeDefaultBasePath
	c.ContainerAnalysisBasePath = ContainerAnalysisDefaultBasePath
	c.DataCatalogBasePath = DataCatalogDefaultBasePath
	c.DataLossPreventionBasePath = DataLossPreventionDefaultBasePath
	c.DataprocBasePath = DataprocDefaultBasePath
	c.DatastoreBasePath = DatastoreDefaultBasePath
	c.DeploymentManagerBasePath = DeploymentManagerDefaultBasePath
	c.DialogflowBasePath = DialogflowDefaultBasePath
	c.DNSBasePath = DNSDefaultBasePath
	c.FilestoreBasePath = FilestoreDefaultBasePath
	c.FirestoreBasePath = FirestoreDefaultBasePath
	c.GameServicesBasePath = GameServicesDefaultBasePath
	c.HealthcareBasePath = HealthcareDefaultBasePath
	c.IapBasePath = IapDefaultBasePath
	c.IdentityPlatformBasePath = IdentityPlatformDefaultBasePath
	c.KMSBasePath = KMSDefaultBasePath
	c.LoggingBasePath = LoggingDefaultBasePath
	c.MLEngineBasePath = MLEngineDefaultBasePath
	c.MonitoringBasePath = MonitoringDefaultBasePath
	c.NetworkManagementBasePath = NetworkManagementDefaultBasePath
	c.OSConfigBasePath = OSConfigDefaultBasePath
	c.OSLoginBasePath = OSLoginDefaultBasePath
	c.PubsubBasePath = PubsubDefaultBasePath
	c.RedisBasePath = RedisDefaultBasePath
	c.ResourceManagerBasePath = ResourceManagerDefaultBasePath
	c.RuntimeConfigBasePath = RuntimeConfigDefaultBasePath
	c.SecretManagerBasePath = SecretManagerDefaultBasePath
	c.SecurityCenterBasePath = SecurityCenterDefaultBasePath
	c.ServiceManagementBasePath = ServiceManagementDefaultBasePath
	c.ServiceUsageBasePath = ServiceUsageDefaultBasePath
	c.SourceRepoBasePath = SourceRepoDefaultBasePath
	c.SpannerBasePath = SpannerDefaultBasePath
	c.SQLBasePath = SQLDefaultBasePath
	c.StorageBasePath = StorageDefaultBasePath
	c.TPUBasePath = TPUDefaultBasePath
	c.VPCAccessBasePath = VPCAccessDefaultBasePath

	// Handwritten Products / Versioned / Atypical Entries
	c.CloudBillingBasePath = CloudBillingDefaultBasePath
	c.ComposerBasePath = ComposerDefaultBasePath
	c.ComputeBetaBasePath = ComputeBetaDefaultBasePath
	c.ContainerBasePath = ContainerDefaultBasePath
	c.ContainerBetaBasePath = ContainerBetaDefaultBasePath
	c.DataprocBasePath = DataprocDefaultBasePath
	c.DataflowBasePath = DataflowDefaultBasePath
	c.DnsBetaBasePath = DnsBetaDefaultBasePath
	c.IamCredentialsBasePath = IamCredentialsDefaultBasePath
	c.ResourceManagerV2Beta1BasePath = ResourceManagerV2Beta1DefaultBasePath
	c.IAMBasePath = IAMDefaultBasePath
	c.ServiceNetworkingBasePath = ServiceNetworkingDefaultBasePath
	c.BigQueryBasePath = BigQueryDefaultBasePath
	c.StorageTransferBasePath = StorageTransferDefaultBasePath
	c.BigtableAdminBasePath = BigtableAdminDefaultBasePath
}
