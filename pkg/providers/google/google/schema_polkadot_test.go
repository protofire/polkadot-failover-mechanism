package google

import (
	"testing"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/stretchr/testify/require"
)

func TestGCPFailoverID(t *testing.T) {
	failoverOrig := &GCPFailover{
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
		Project: "test",
	}

	id, err := failoverOrig.ID()
	require.NoError(t, err)

	failoverUnpack := &GCPFailover{}
	err = failoverUnpack.FromID(id)
	require.NoError(t, err)
	require.Equal(t, failoverOrig, failoverUnpack)

}

func TestGCPFailoverIDOrSchema(t *testing.T) {
	failoverOrig := &GCPFailover{
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
		Project: "test",
	}

	id, err := failoverOrig.ID()
	require.NoError(t, err)

	d := &schema.ResourceData{}
	d.SetId(id)

	failoverUnpack := &GCPFailover{}
	err = failoverUnpack.FromIDOrSchema(d)
	require.NoError(t, err)
	require.Equal(t, failoverOrig, failoverUnpack)

}

func TestGCPFailoverSetCount(t *testing.T) {
	failover := &GCPFailover{}

	failover.SetCounts(1, 2, 3)
	require.Equal(t, 1, failover.PrimaryCount)
	require.Equal(t, 2, failover.SecondaryCount)
	require.Equal(t, 3, failover.TertiaryCount)
}
