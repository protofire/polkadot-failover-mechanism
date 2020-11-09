package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/autoscaling"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

type Config struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Regions       []string
	MaxRetries    int

	AssumeRoleARN               string
	AssumeRoleDurationSeconds   int
	AssumeRoleExternalID        string
	AssumeRolePolicy            string
	AssumeRolePolicyARNs        []string
	AssumeRoleSessionName       string
	AssumeRoleTags              map[string]string
	AssumeRoleTransitiveTagKeys []string

	AllowedAccountIds   []string
	ForbiddenAccountIds []string

	Endpoints map[string]string
	Insecure  bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountID bool
	SkipMetadataAPICheck    bool
	S3ForcePathStyle        bool

	terraformVersion string
}

type Client struct {
	cloudwatchconn     *cloudwatch.CloudWatch
	autoscalingconn    *autoscaling.AutoScaling
	ec2conn            *ec2.EC2
	dnsSuffix          string
	supportedplatforms []string
	region             string
}

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (client *Client) PartitionHostname(prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, client.dnsSuffix)
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (client *Client) RegionalHostname(prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, client.region, client.dnsSuffix)
}

func configureRegionClient(c *Config, regionIdx int) (*Client, error) {
	region := c.Regions[regionIdx]

	if !c.SkipRegionValidation {
		if err := awsbase.ValidateRegion(c.Regions[regionIdx]); err != nil {
			return nil, err
		}
	}

	awsbaseConfig := &awsbase.Config{
		AccessKey:                   c.AccessKey,
		AssumeRoleARN:               c.AssumeRoleARN,
		AssumeRoleDurationSeconds:   c.AssumeRoleDurationSeconds,
		AssumeRoleExternalID:        c.AssumeRoleExternalID,
		AssumeRolePolicy:            c.AssumeRolePolicy,
		AssumeRolePolicyARNs:        c.AssumeRolePolicyARNs,
		AssumeRoleSessionName:       c.AssumeRoleSessionName,
		AssumeRoleTags:              c.AssumeRoleTags,
		AssumeRoleTransitiveTagKeys: c.AssumeRoleTransitiveTagKeys,
		CallerDocumentationURL:      "https://registry.terraform.io/providers/hashicorp/aws",
		CallerName:                  "Terraform AWS Provider",
		CredsFilename:               c.CredsFilename,
		DebugLogging:                logging.IsDebugOrHigher(),
		IamEndpoint:                 c.Endpoints["iam"],
		Insecure:                    c.Insecure,
		MaxRetries:                  c.MaxRetries,
		Profile:                     c.Profile,
		Region:                      region,
		SecretKey:                   c.SecretKey,
		SkipCredsValidation:         c.SkipCredsValidation,
		SkipMetadataApiCheck:        c.SkipMetadataAPICheck,
		SkipRequestingAccountId:     c.SkipRequestingAccountID,
		StsEndpoint:                 c.Endpoints["sts"],
		Token:                       c.Token,
		UserAgentProducts: []*awsbase.UserAgentProduct{
			{Name: "APN", Version: "1.0"},
			{Name: "HashiCorp", Version: "1.0"},
			{Name: "Terraform", Version: c.terraformVersion,
				Extra: []string{"+https://www.terraform.io"}},
		},
	}

	sess, accountID, partition, err := awsbase.GetSessionWithAccountIDAndPartition(awsbaseConfig)
	if err != nil {
		return nil, fmt.Errorf("error configuring Terraform AWS Provider: %w", err)
	}

	if accountID == "" {
		log.Printf("[WARN] AWS account ID not found for provider. See https://www.terraform.io/docs/providers/aws/index.html#skip_requesting_account_id for implications.")
	}

	if err := awsbase.ValidateAccountID(accountID, c.AllowedAccountIds, c.ForbiddenAccountIds); err != nil {
		return nil, err
	}

	dnsSuffix := "amazonaws.com"
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		dnsSuffix = p.DNSSuffix()
	}

	client := &Client{
		cloudwatchconn:  cloudwatch.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["cloudwatch"])})),
		ec2conn:         ec2.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["ec2"])})),
		autoscalingconn: autoscaling.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["autoscaling"])})),
		dnsSuffix:       dnsSuffix,
		region:          region,
	}

	// "Global" services that require customizations
	globalAcceleratorConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["globalaccelerator"]),
	}
	route53Config := &aws.Config{
		Endpoint: aws.String(c.Endpoints["route53"]),
	}
	shieldConfig := &aws.Config{
		Endpoint: aws.String(c.Endpoints["shield"]),
	}

	// Force "global" services to correct regions
	switch partition {
	case endpoints.AwsPartitionID:
		globalAcceleratorConfig.Region = aws.String(endpoints.UsWest2RegionID)
		route53Config.Region = aws.String(endpoints.UsEast1RegionID)
		shieldConfig.Region = aws.String(endpoints.UsEast1RegionID)
	case endpoints.AwsCnPartitionID:
		// The AWS Go SDK is missing endpoint information for Route 53 in the AWS China partition.
		// This can likely be removed in the future.
		if aws.StringValue(route53Config.Endpoint) == "" {
			route53Config.Endpoint = aws.String("https://api.route53.cn")
		}
		route53Config.Region = aws.String(endpoints.CnNorthwest1RegionID)
	case endpoints.AwsUsGovPartitionID:
		route53Config.Region = aws.String(endpoints.UsGovWest1RegionID)
	}

	client.ec2conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if r.Operation.Name == "CreateClientVpnEndpoint" {
			if isAWSErr(r.Error, "OperationNotPermitted", "Endpoint cannot be created while another endpoint is being created") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "CreateVpnConnection" {
			if isAWSErr(r.Error, "VpnConnectionLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "CreateVpnGateway" {
			if isAWSErr(r.Error, "VpnGatewayLimitExceeded", "maximum number of mutating objects has been reached") {
				r.Retryable = aws.Bool(true)
			}
		}

		if r.Operation.Name == "AttachVpnGateway" || r.Operation.Name == "DetachVpnGateway" {
			if isAWSErr(r.Error, "InvalidParameterValue", "This call cannot be completed because there are pending VPNs or Virtual Interfaces") {
				r.Retryable = aws.Bool(true)
			}
		}
	})

	if !c.SkipGetEC2Platforms {
		supportedPlatforms, err := GetSupportedEC2Platforms(client.ec2conn)
		if err != nil {
			// We intentionally fail *silently* because there's a chance
			// user just doesn't have ec2:DescribeAccountAttributes permissions
			log.Printf("[WARN] Unable to get supported EC2 platforms: %s", err)
		} else {
			client.supportedplatforms = supportedPlatforms
		}
	}

	return client, nil

}

// Client configures and returns a fully initialized Client
func (c *Config) Client() (interface{}, diag.Diagnostics) {
	// Get the auth and region. This can fail if keys/regions were not
	// specified and we're attempting to use the environment.
	if !c.SkipRegionValidation {
		for _, region := range c.Regions {
			if err := awsbase.ValidateRegion(region); err != nil {
				return nil, diag.FromErr(err)
			}
		}
	}

	var clients []*Client

	for idx := range c.Regions {
		client, err := configureRegionClient(c, idx)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		clients = append(clients, client)

	}

	return clients, nil

}

func GetSupportedEC2Platforms(conn *ec2.EC2) ([]string, error) {
	attrName := "supported-platforms"

	input := ec2.DescribeAccountAttributesInput{
		AttributeNames: []*string{aws.String(attrName)},
	}
	attributes, err := conn.DescribeAccountAttributes(&input)
	if err != nil {
		return nil, err
	}

	var platforms []string
	for _, attr := range attributes.AccountAttributes {
		if *attr.AttributeName == attrName {
			for _, v := range attr.AttributeValues {
				platforms = append(platforms, *v.AttributeValue)
			}
			break
		}
	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("No EC2 platforms detected")
	}

	return platforms, nil
}
