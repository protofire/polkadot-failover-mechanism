package test

// This file contains all the supplementary functions that are required to query EC2's Security groups API

import (
	"testing"

	taws "github.com/gruntwork-io/terratest/modules/aws"
    aws "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/stretchr/testify/require"
)

// Function that returns a set of Security group permissions for particular prefix
func GetSGRulesMapByTag(t *testing.T, awsRegion string, tag string, value string) []*ec2.IpPermission {
	rules, err := GetSGRulesMapByTagE(t, awsRegion, tag, value)
        require.NoError(t, err)
        return rules
}

func GetSGRulesMapByTagE(t *testing.T, awsRegion string, tag string, value string) ([]*ec2.IpPermission, error) {
	asg := NewSGClient(t, awsRegion)

	ec2FilterList := []*ec2.Filter{
		{
			Name: aws.String("tag:" + tag),
			Values: aws.StringSlice([]string{value}),
		},
	}
	input := &ec2.DescribeSecurityGroupsInput{Filters: ec2FilterList}
	result, err := asg.DescribeSecurityGroups(input)

	if err != nil {
		return nil, err
	}

	var rules []*ec2.IpPermission

	for _, value := range result.SecurityGroups {
		rules = append(rules,value.IpPermissions...)
	}

	return rules,nil
}

// NewSsmClient creates a SSM client.
func NewSGClient(t *testing.T, region string) *ec2.EC2 {
        client, err := NewSGClientE(t, region)
        require.NoError(t, err)
        return client
}

func NewSGClientE(t *testing.T, region string) (*ec2.EC2, error) {
        sess, err := taws.NewAuthenticatedSession(region)
        if err != nil {
                return nil, err
        }

        return ec2.New(sess), nil
}
