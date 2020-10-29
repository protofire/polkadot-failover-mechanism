package aws

// This file contains all the supplementary functions that are required to query SSM API

import (
	"testing"

	saws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	taws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
)

// getParameterTypeAndValue retrieves the latest version of SSM Parameter and it's type with decryption
func getParameterTypeAndValue(t *testing.T, awsRegion string, keyName string) (string, string) {
	keyType, keyValue, err := getParameterTypeAndValueE(t, awsRegion, keyName)
	require.NoError(t, err)
	return keyType, keyValue
}

func getParameterTypeAndValueE(t *testing.T, awsRegion string, keyName string) (string, string, error) {
	ssmClient := newSsmClient(t, awsRegion)

	resp, err := ssmClient.GetParameter(&ssm.GetParameterInput{Name: saws.String(keyName), WithDecryption: saws.Bool(true)})
	if err != nil {
		return "", "", err
	}

	parameter := *resp.Parameter
	return *parameter.Type, *parameter.Value, nil
}

// newSsmClient creates a SSM client.
func newSsmClient(t *testing.T, region string) *ssm.SSM {
	client, err := newSsmClientE(region)
	require.NoError(t, err)
	return client
}

func newSsmClientE(region string) (*ssm.SSM, error) {
	sess, err := taws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return ssm.New(sess), nil
}

// typeAndValueComparator Supplementary function: Checks that given parameter in each parameter exists and has the right type
// (e.g. all the encrypted parameters has the SecureString type)
func typeAndValueComparator(t *testing.T, relativePath string, expectedType string, expectedValue string, awsRegions []string, prefix string) int {

	for _, region := range awsRegions {
		ssmType, ssmValue := getParameterTypeAndValue(t, region, "/polkadot/validator-failover/"+prefix+"/"+relativePath)
		if ssmType == expectedType && ssmValue == expectedValue {
			t.Log("INFO. SSM Parameter " + relativePath + " of type " + ssmType + " and value " + ssmValue + " at region " + region + " matched prefedined value.")
		} else {
			t.Error("ERROR! No match for SSM parameter " + relativePath + " at region " + region + ". Expected type: " + expectedType + ", expected value: " + expectedValue + ". Actual type: " + ssmType + ", actual value: " + ssmValue)
			return 0
		}
	}

	return 1
}

// SSMCheck checks ssm parameters
func SSMCheck(t *testing.T, awsRegions []string, prefix string) bool {

	result := typeAndValueComparator(t, "cpu_limit", "String", "1", awsRegions, prefix) *
		typeAndValueComparator(t, "ram_limit", "String", "1", awsRegions, prefix) *
		typeAndValueComparator(t, "name", "String", "test", awsRegions, prefix) *
		typeAndValueComparator(t, "keys/key1/type", "String", "gran", awsRegions, prefix) *
		typeAndValueComparator(t, "keys/key1/seed", "SecureString", "favorite liar zebra assume hurt cage any damp inherit rescue delay panic", awsRegions, prefix) *
		typeAndValueComparator(t, "keys/key1/key", "String", "0x6ce96ae5c300096b09dbd4567b0574f6a1281ae0e5cfe4f6b0233d1821f6206b", awsRegions, prefix) *
		typeAndValueComparator(t, "keys/key2/type", "String", "aura", awsRegions, prefix) *
		typeAndValueComparator(t, "keys/key2/seed", "SecureString", "expire stage crawl shell boss any story swamp skull yellow bamboo copy", awsRegions, prefix) *
		typeAndValueComparator(t, "keys/key2/key", "String", "0x3ff0766f9ebbbceee6c2f40d9323164d07e70c70994c9d00a9512be6680c2394", awsRegions, prefix)

	return result == 1

}
