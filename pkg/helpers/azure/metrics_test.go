package azure

import (
	"encoding/json"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/stretchr/testify/require"
)

var metricsSuccessResponse = `
{
  "cost": 0,
  "timespan": "2020-10-17T18:15:39Z/2020-10-17T19:15:39Z",
  "interval": "PT1H",
  "value": [
    {
      "id": "/subscriptions/6ad71a09-e4a3-44e0-8e5f-df997c709a74/resourceGroups/814_Protofire_Web3/providers/Microsoft.Compute/virtualMachineScaleSets/qfbm-instance-primary/providers/Microsoft.Insights/metrics/Value",
      "type": "Microsoft.Insights/metrics",
      "name": {
        "value": "Value",
        "localizedValue": "Value"
      },
      "unit": "Unspecified",
      "timeseries": [
        {
          "metadatavalues": [
            {
              "name": {
                "value": "host",
                "localizedValue": "host"
              },
              "value": "primary000002"
            }
          ],
          "data": [
            {
              "timeStamp": "2020-10-17T18:15:00Z",
              "maximum": 1
            }
          ]
        }
      ],
      "errorCode": "Success"
    }
  ],
  "namespace": "qfbm/validator",
  "resourceregion": "centralus"
}
`

var metricResponse = `
{
    "id": "/subscriptions/6ad71a09-e4a3-44e0-8e5f-df997c709a74/resourceGroups/814_Protofire_Web3/providers/Microsoft.Compute/virtualMachineScaleSets/vzsm-instance-primary/providers/Microsoft.Insights/metrics/value",
    "type": "Microsoft.Insights/metrics",
    "name": {
        "value": "value",
        "localizedValue": "value"
    },
    "unit": "Unspecified",
    "timeseries": [
        {
            "metadatavalues": [
                {
                    "name": {
                        "value": "host",
                        "localizedValue": "host"
                    },
                    "value": "primary000000"
                }
            ],
            "data": [
                {
                    "timeStamp": "2020-10-21T23:38:00Z"
                },
                {
                    "timeStamp": "2020-10-21T23:39:00Z"
                },
                {
                    "timeStamp": "2020-10-21T23:40:00Z"
                },
                {
                    "timeStamp": "2020-10-21T23:41:00Z"
                },
                {
                    "timeStamp": "2020-10-21T23:42:00Z",
                    "maximum": 1
                }
            ]
        }
    ]
}
`

var metricsBlankResponse = `
{
  "cost": 0,
  "timespan": "2020-10-17T18:15:39Z/2020-10-17T19:15:39Z",
  "interval": "PT1H",
  "value": [
    {
      "id": "/subscriptions/6ad71a09-e4a3-44e0-8e5f-df997c709a74/resourceGroups/814_Protofire_Web3/providers/Microsoft.Compute/virtualMachineScaleSets/qfbm-instance-primary/providers/Microsoft.Insights/metrics/Value",
      "type": "Microsoft.Insights/metrics",
      "name": {
        "value": "Value",
        "localizedValue": "Value"
      },
      "unit": "Unspecified",
      "timeseries": [],
      "errorCode": "Success"
    }
  ],
  "namespace": "qfbm/validator",
  "resourceregion": "centralus"
}
`

func marshallResponse(data string) (insights.Response, error) {
	resp := insights.Response{}
	err := json.Unmarshal([]byte(data), &resp)
	return resp, err
}

func marshallMetric(data string) (insights.Metric, error) {
	resp := insights.Metric{}
	err := json.Unmarshal([]byte(data), &resp)
	return resp, err
}

func TestMetricsResponse(t *testing.T) {
	responseSuccess, err := marshallResponse(metricsSuccessResponse)
	require.NoError(t, err)
	responseBlank, err := marshallResponse(metricsBlankResponse)
	require.NoError(t, err)

	vmSSName1 := "test1"
	vmSSName2 := "test2"

	mp := make(map[string]insights.Metric)
	mp[vmSSName1] = (*responseSuccess.Value)[0]
	mp[vmSSName2] = (*responseBlank.Value)[0]

	validator, err := findValidator(mp, insights.Maximum, 1)
	require.NoError(t, err)
	require.Equal(t, vmSSName1, validator.ScaleSetName)
	require.Equal(t, "primary000002", validator.Hostname)
	require.Equal(t, 1, validator.Metric)

}

func TestMetrics(t *testing.T) {
	metric, err := marshallMetric(metricResponse)
	require.NoError(t, err)

	vmSSName1 := "test1"

	mp := make(map[string]insights.Metric)
	mp[vmSSName1] = metric

	validator, err := findValidator(mp, insights.Maximum, 1)
	require.NoError(t, err)
	require.Equal(t, vmSSName1, validator.ScaleSetName)
	require.Equal(t, "primary000000", validator.Hostname)
	require.Equal(t, 1, validator.Metric)

}
