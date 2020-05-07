package test

import (
	"testing"

	taws "github.com/gruntwork-io/terratest/modules/aws"
        "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/stretchr/testify/require"
)

func GetAlarmsNamesAndStatesByPrefix(t *testing.T, awsRegion string, prefix string) map[string]string {
	out, err := GetAlarmsNamesAndStatesByPrefixE(t, awsRegion, prefix)
	if err != nil {
		t.Error(err)
                return make(map[string]string)
	}
	return out
}

func GetAlarmsNamesAndStatesByPrefixE(t *testing.T, awsRegion string, prefix string) (map[string]string, error) {
	result := make(map[string]string)

        cw := NewCWClient(t, awsRegion)

	var input = &cloudwatch.DescribeAlarmsInput {
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

func NewCWClient(t *testing.T, region string) *cloudwatch.CloudWatch {
        client, err := NewCWClientE(t, region)
        require.NoError(t, err)
        return client
}

func NewCWClientE(t *testing.T, region string) (*cloudwatch.CloudWatch, error) {
        sess, err := taws.NewAuthenticatedSession(region)
        if err != nil {
                return nil, err
        }

        return cloudwatch.New(sess), nil
}
