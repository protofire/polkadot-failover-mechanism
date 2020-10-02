package utils

// This file contains all the supplementary functions that are required to query EC2 API

import (
	"fmt"
	"testing"

	taws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
)

// GetHealthyEc2InstanceIdsByTag External function that returns a list of instance IDs that are running in given region
func GetHealthyEc2InstanceIdsByTag(t *testing.T, region string, tagName string, tagValue string) []string {
	out, err := getHealthyEc2InstanceIdsByTagE(t, region, tagName, tagValue)
	require.NoError(t, err)
	return out
}

func getHealthyEc2InstanceIdsByTagE(t *testing.T, region string, tagName string, tagValue string) ([]string, error) {
	ec2Filters := map[string][]string{
		"instance-state-name":          {"running"},
		fmt.Sprintf("tag:%s", tagName): {tagValue},
	}
	return taws.GetEc2InstanceIdsByFiltersE(t, region, ec2Filters)
}
