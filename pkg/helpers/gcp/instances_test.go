package gcp

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/api/compute/v1"
)

func TestDeleteManagementInstances(t *testing.T) {

	names := []string{"z", "x", "y", "y"}
	instancesCount := []int{2, 3, 2, 2}
	groups := []InstanceGroupManager{
		{
			Instances: []*compute.ManagedInstance{
				{
					Instance: "x/y/z",
				},
				{
					Instance: "x/y/zz",
				},
				{
					Instance: "x/y/zzz",
				},
			},
		},
		{
			Instances: []*compute.ManagedInstance{
				{
					Instance: "x/y/z",
				},
				{
					Instance: "x/y/zz",
				},
				{
					Instance: "x/y/zzz",
				},
			},
		},
		{
			Instances: []*compute.ManagedInstance{
				{
					Instance: "x/y/yyy",
				},
				{
					Instance: "x/y/y",
				},
				{
					Instance: "x/y/yyy",
				},
			},
		},
		{
			Instances: []*compute.ManagedInstance{
				{
					Instance: "x/y/y",
				},
				{
					Instance: "x/y/y",
				},
				{
					Instance: "x/y/y",
				},
			},
		},
	}

	for idx, group := range groups {
		group.SearchAndRemoveInstanceByName(names[idx])
		require.Len(t, group.Instances, instancesCount[idx])
		for _, instance := range group.Instances {
			require.NotNil(t, instance)
		}
	}

}
