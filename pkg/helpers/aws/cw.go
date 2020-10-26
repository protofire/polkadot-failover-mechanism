package aws

// This file contains all the supplementary functions that are required to query Cloud Watch (AWS)

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	taws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/require"
)

// getAlarmsNamesAndStatesByPrefix External function that receives prefix as argument and returns all alarms with that prefix in the given region
func getAlarmsNamesAndStatesByPrefix(t *testing.T, awsRegion string, prefix string) map[string]string {
	out, err := getAlarmsNamesAndStatesByPrefixE(t, awsRegion, prefix)
	if err != nil {
		t.Error(err)
		return make(map[string]string)
	}
	return out
}

func getAlarmsNamesAndStatesByPrefixE(t *testing.T, awsRegion string, prefix string) (map[string]string, error) {
	result := make(map[string]string)

	cw := newCWClient(t, awsRegion)

	var input = &cloudwatch.DescribeAlarmsInput{
		AlarmNamePrefix: &prefix,
	}

	output, err := cw.DescribeAlarms(input)

	if err != nil {
		t.Error(err)
		return result, err
	}

	for _, v := range output.MetricAlarms {
		result[*v.AlarmName] = *v.StateValue
	}

	return result, nil

}

// newCWClient Supplementary function that enables communications with CW API
func newCWClient(t *testing.T, region string) *cloudwatch.CloudWatch {
	client, err := newCWClientE(region)
	require.NoError(t, err)
	return client
}

func newCWClientE(region string) (*cloudwatch.CloudWatch, error) {
	sess, err := taws.NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return cloudwatch.New(sess), nil
}

// CloudWatchCheck checks cloud watch
func CloudWatchCheck(t *testing.T, awsRegions []string, prefix string, expectedAlertsCount int) bool {

	errors := 0
	attempts := 0
	maxAttempts := 100
	for _, region := range awsRegions {
		for {
			insufficientDataFlag := false
			attempts++
			alarms := getAlarmsNamesAndStatesByPrefix(t, region, prefix)
			alarmsCount := len(alarms)

			// Check that there are exactly 4 CloudWatch alarms (should be changed here if new alarms added)
			if alarmsCount != expectedAlertsCount {
				t.Errorf("ERROR! It is expected to have %d CloudWatch Alarms in total, got %d", expectedAlertsCount, alarmsCount)
				continue
			}
			t.Logf("INFO. CloudWatch Alarms number matches the predefined value of %d", expectedAlertsCount)

			// If alarm still has "INSUFFICIENT DATA" status - we need to wait until alarm either triggers or move into "OK" state.
		loop:
			for k, v := range alarms {
				switch v {
				case "OK":
					t.Logf("INFO. The CloudWatch Alarm %s in region %s has the state OK!", k, region)
					continue
				case "INSUFFICIENT_DATA":
					t.Logf("INFO. The CloudWatch Alarm %s in region %s has insufficient data right now.", k, region)
					insufficientDataFlag = true
					break loop
				default:
					t.Errorf("ERROR! The CloudWatch Alarm %s in region %s has the state %s, which is not OK", k, region, v)
					errors++
				}
			}

			if attempts >= maxAttempts {
				t.Errorf("Max attempts waiting metrics status")
				return false
			}
			// If some of the alarms has insufficient data state - rerun all the checks once again.
			if !insufficientDataFlag {
				break
			} else {
				t.Log("Sleeping 10 seconds before retrying...")
				time.Sleep(10 * time.Second)
			}
		}

	}

	return errors == 0
}
