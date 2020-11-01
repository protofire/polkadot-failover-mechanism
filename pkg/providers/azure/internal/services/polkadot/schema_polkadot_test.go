package polkadot

import (
	"testing"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"

	"github.com/stretchr/testify/require"
)

func TestAzureFailoverID(t *testing.T) {
	failoverOrig := &AzureFailover{
		Failover: resource.Failover{
			Prefix:            "test",
			FailoverMode:      resource.FailOverModeDistributed,
			MetricName:        "test",
			MetricNameSpace:   "test",
			Instances:         []int{1, 2, 3},
			Locations:         []string{"1", "2", "3"},
			PrimaryCount:      1,
			SecondaryCount:    0,
			TertiaryCount:     0,
			FailoverInstances: []int{1, 0, 0},
			Source:            resource.FailoverSourceID,
		},
		ResourceGroup: "test",
	}

	id, err := failoverOrig.ID()
	require.NoError(t, err)

	failoverUnpack := &AzureFailover{}
	err = failoverUnpack.FromID(id)
	require.NoError(t, err)
	require.Equal(t, failoverOrig, failoverUnpack)

}
