package test

// This file contains all the supplementary functions that are required to query EC2 API

import (
        "fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
        "github.com/stretchr/testify/require"
)


// External function that returns a list of instance IDs that are running in given region
func GetHealthyEc2InstanceIdsByTag(t *testing.T, region string, tagName string, tagValue string) []string {
        out, err := GetHealthyEc2InstanceIdsByTagE(t, region, tagName, tagValue)
        require.NoError(t, err)
        return out
}

func GetHealthyEc2InstanceIdsByTagE(t *testing.T, region string, tagName string, tagValue string) ([]string, error) {
        ec2Filters := map[string][]string{
                "instance-state-name": {"running"},
                fmt.Sprintf("tag:%s", tagName): {tagValue},
        }
        return aws.GetEc2InstanceIdsByFiltersE(t, region, ec2Filters)
}
