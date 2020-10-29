package failover

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateInstanceCountPerRegion(t *testing.T) {
	counts := CalculateInstanceCountPerRegion([]string{"Central US", "East US", "West US"}, "centralus", func(s string) string {
		return strings.Replace(strings.ToLower(s), " ", "", -1)
	})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstanceCountPerRegion([]string{"centralus", "East US", "West US"}, "centralus", func(s string) string {
		return strings.Replace(strings.ToLower(s), " ", "", -1)
	})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstanceCountPerRegion([]string{"centralus", "East US", "West US"}, "", func(s string) string {
		return strings.Replace(strings.ToLower(s), " ", "", -1)
	})
	require.Equal(t, []int{1, 0, 0}, counts)
}

func TestCalculateInstancesForSingleFailOverMode(t *testing.T) {
	counts := CalculateInstancesForSingleFailOverMode([]int{1, 0, 0})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstancesForSingleFailOverMode([]int{2, 0, 0})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstancesForSingleFailOverMode([]int{0, 0, 0})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstancesForSingleFailOverMode([]int{0, 0, 1})
	require.Equal(t, []int{0, 0, 1}, counts)
	counts = CalculateInstancesForSingleFailOverMode([]int{1, 1, 1})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstancesForSingleFailOverMode([]int{10, 10, 10})
	require.Equal(t, []int{1, 0, 0}, counts)
	counts = CalculateInstancesForSingleFailOverMode([]int{10, 10, 11})
	require.Equal(t, []int{0, 0, 1}, counts)
}
