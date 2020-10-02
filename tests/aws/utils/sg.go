package utils

// This file contains all the supplementary functions that are required to query EC2's Security groups API

import (
	"testing"

	saws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/go-cmp/cmp"
	taws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
)

// getSGRulesMapByTag Function that returns a set of Security group permissions for particular prefix
func getSGRulesMapByTag(t *testing.T, awsRegion string, tag string, value string) []*ec2.IpPermission {
	rules, err := getSGRulesMapByTagE(t, awsRegion, tag, value)
	require.NoError(t, err)
	return rules
}

func getSGRulesMapByTagE(t *testing.T, awsRegion string, tag string, value string) ([]*ec2.IpPermission, error) {
	asg := newSGClient(t, awsRegion)

	ec2FilterList := []*ec2.Filter{
		{
			Name:   saws.String("tag:" + tag),
			Values: saws.StringSlice([]string{value}),
		},
	}
	input := &ec2.DescribeSecurityGroupsInput{Filters: ec2FilterList}
	result, err := asg.DescribeSecurityGroups(input)

	if err != nil {
		return nil, err
	}

	var rules []*ec2.IpPermission

	for _, value := range result.SecurityGroups {
		rules = append(rules, value.IpPermissions...)
	}

	return rules, nil
}

func newSGClient(t *testing.T, region string) *ec2.EC2 {
	client, err := newSGClientE(t, region)
	require.NoError(t, err)
	return client
}

func newSGClientE(t *testing.T, region string) (*ec2.EC2, error) {
	sess, err := taws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return ec2.New(sess), nil
}

// SGCheck checks security groups
func SGCheck(t *testing.T, awsRegions []string, prefix string) bool {

	// A set of predefined security rules to compare existing rules with.
	fromPorts := []int64{30333, 22, 8301, 8600, 8500, 8300}
	toPorts := []int64{30333, 22, 8302, 8600, 8500}
	ipProtocols := []string{"tcp", "udp"}
	cidrIPs := []string{"0.0.0.0/0", "10.2.0.0/16", "10.1.0.0/16", "10.0.0.0/16"}

	var rules = []*ec2.IpPermission{
		{
			FromPort:   &fromPorts[0],
			IpProtocol: &ipProtocols[0],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[0],
				},
			},
			ToPort: &toPorts[0],
		},
		{
			FromPort:   &fromPorts[1],
			IpProtocol: &ipProtocols[0],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[0],
				},
			},
			ToPort: &toPorts[1],
		},
		{
			FromPort:   &fromPorts[2],
			IpProtocol: &ipProtocols[1],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[1],
				},
				{
					CidrIp: &cidrIPs[2],
				},
				{
					CidrIp: &cidrIPs[3],
				},
			},
			ToPort: &toPorts[2],
		},
		{
			FromPort:   &fromPorts[3],
			IpProtocol: &ipProtocols[1],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[2],
				},
				{
					CidrIp: &cidrIPs[3],
				},
				{
					CidrIp: &cidrIPs[1],
				},
			},
			ToPort: &toPorts[3],
		},
		{
			FromPort:   &fromPorts[4],
			IpProtocol: &ipProtocols[1],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[1],
				},
				{
					CidrIp: &cidrIPs[2],
				},
				{
					CidrIp: &cidrIPs[3],
				},
			},
			ToPort: &toPorts[4],
		},
		{
			FromPort:   &fromPorts[0],
			IpProtocol: &ipProtocols[1],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[0],
				},
			},
			ToPort: &toPorts[0],
		},
		{
			FromPort:   &fromPorts[4],
			IpProtocol: &ipProtocols[0],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[3],
				},
				{
					CidrIp: &cidrIPs[2],
				},
				{
					CidrIp: &cidrIPs[1],
				},
			},
			ToPort: &toPorts[4],
		},
		{
			FromPort:   &fromPorts[5],
			IpProtocol: &ipProtocols[0],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[3],
				},
				{
					CidrIp: &cidrIPs[2],
				},
				{
					CidrIp: &cidrIPs[1],
				},
			},
			ToPort: &toPorts[2],
		},
		{
			FromPort:   &fromPorts[3],
			IpProtocol: &ipProtocols[0],
			IpRanges: []*ec2.IpRange{
				{
					CidrIp: &cidrIPs[1],
				},
				{
					CidrIp: &cidrIPs[3],
				},
				{
					CidrIp: &cidrIPs[2],
				},
			},
			ToPort: &toPorts[3],
		},
	}

	// For each region fetch all the security groups prefixed with predefined prefix and compare it one by one with a list of predefined groups
	for _, region := range awsRegions {

		t.Logf("INFO. Checking matching of SG rules. Region: %s...", region)

		ruleSlice := getSGRulesMapByTag(t, region, "prefix", prefix)
		lenRuleSlice := len(ruleSlice)

		if lenRuleSlice != 9 {
			t.Errorf("ERROR! Expecting to get 9 security groups, got %d", lenRuleSlice)
			return false
		}

		for _, ruleSet := range ruleSlice {
			found := 0
			for _, ruleExpect := range rules {
				if cmp.Equal(ruleSet, ruleExpect) {
					found = 1
					break
				}
			}
			if found != 1 {
				t.Errorf("ERROR! No match were found for current record: %#v. Region: %s", ruleSet, region)
				return false
			}
			t.Logf("INFO. The following record matches one of the predefined security rules: %#v. Region: %s", ruleSet, region)
		}
	}

	return true
}
