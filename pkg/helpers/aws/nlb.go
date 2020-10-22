package aws

// This file contains all the supplementary functions that are required to query Load Balancer API V2

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	taws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
)

// getHealthStatusSliceByLBsARN External function that returns a map of target groups and their health statuses
func getHealthStatusSliceByLBsARN(t *testing.T, awsRegion string, arn string) map[string]string {
	result := make(map[string]string)

	TGSSlice := getTGsbyLBsARN(t, awsRegion, arn)
	for _, tg := range TGSSlice.TargetGroups {
		TGHealth := getHealthStatusOfTG(t, awsRegion, tg.TargetGroupArn)

		for _, instance := range TGHealth.TargetHealthDescriptions {
			result[*tg.TargetGroupArn] = *instance.TargetHealth.State
		}
	}

	return result
}

// Function that receives health status of the given target group
func getHealthStatusOfTG(t *testing.T, awsRegion string, tg *string) *elbv2.DescribeTargetHealthOutput {
	rules, err := getHealthStatusOfTGE(t, awsRegion, tg)
	require.NoError(t, err)
	return rules
}

func getHealthStatusOfTGE(t *testing.T, awsRegion string, tg *string) (*elbv2.DescribeTargetHealthOutput, error) {
	nlb := newNLBClient(t, awsRegion)

	var input = &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: tg,
	}

	return nlb.DescribeTargetHealth(input)
}

// Function that receives all the target groups for the given load balancer
func getTGsbyLBsARN(t *testing.T, awsRegion string, arn string) *elbv2.DescribeTargetGroupsOutput {
	rules, err := getTGsbyLBsARNE(t, awsRegion, arn)
	require.NoError(t, err)
	return rules
}

func getTGsbyLBsARNE(t *testing.T, awsRegion string, arn string) (*elbv2.DescribeTargetGroupsOutput, error) {
	nlb := newNLBClient(t, awsRegion)

	var input = &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: &arn,
	}
	return nlb.DescribeTargetGroups(input)
}

// newNLBClient creates a SSM client.
func newNLBClient(t *testing.T, region string) *elbv2.ELBV2 {
	client, err := newNLBClientE(region)
	require.NoError(t, err)
	return client
}

func newNLBClientE(region string) (*elbv2.ELBV2, error) {
	sess, err := taws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return elbv2.New(sess), nil
}

// NLBCheck cjecks NLB
func NLBCheck(t *testing.T, lbs []string, awsRegions []string) bool {
	var err bool = false
	for i, lb := range lbs {
		var errLocal bool = false

		resultMap := getHealthStatusSliceByLBsARN(t, awsRegions[i], lb)
		lenResultMap := len(resultMap)

		// Check that there exactly 6 TargetGroup were created
		if lenResultMap != 6 {
			t.Errorf("ERROR! Expected 6 TGs at LoadBalancer %s got %d", lb, lenResultMap)
			err = true
		} else {
			t.Logf("INFO. There are exactly 6 TGs at LoadBalancer %s", lb)
		}

		// Check that TG reports healthy status
		for TG, result := range resultMap {
			if result != "healthy" {
				t.Error("DEBUG. The LB " + lb + " contains TG " + TG + " with not healthy instances. Instance health status is " + result)
				err = true
				errLocal = true
			}
		}
		if errLocal {
			t.Log("ERROR! The LB " + lb + " contains some TG that are not healthy")
		} else {
			t.Log("All TGs in LB " + lb + " contains only healthy instances.")
		}
	}
	return !err
}
