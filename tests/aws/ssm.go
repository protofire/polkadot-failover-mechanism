package test

// This file contains all the supplementary functions that are required to query SSM API

import (
	"testing"

	taws "github.com/gruntwork-io/terratest/modules/aws"
	aws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/require"
)

// GetParameter retrieves the latest version of SSM Parameter and it's type with decryption
func GetParameterTypeAndValue(t *testing.T, awsRegion string, keyName string) (string, string) {
	keyType, keyValue, err := GetParameterTypeAndValueE(t, awsRegion, keyName)
	require.NoError(t, err)
	return keyType, keyValue
}

func GetParameterTypeAndValueE(t *testing.T, awsRegion string, keyName string) (string, string, error) {
	ssmClient := NewSsmClient(t, awsRegion)

	resp, err := ssmClient.GetParameter(&ssm.GetParameterInput{Name: aws.String(keyName), WithDecryption: aws.Bool(true)})
	if err != nil {
		return "", "", err
	}

	parameter := *resp.Parameter
	return *parameter.Type, *parameter.Value, nil
}

// NewSsmClient creates a SSM client.
func NewSsmClient(t *testing.T, region string) *ssm.SSM {
	client, err := NewSsmClientE(t, region)
	require.NoError(t, err)
	return client
}

func NewSsmClientE(t *testing.T, region string) (*ssm.SSM, error) {
	sess, err := taws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return ssm.New(sess), nil
}
