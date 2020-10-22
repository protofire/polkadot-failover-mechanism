package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortByPosition(t *testing.T) {
	values := []int{1, 10, 15}
	names := []string{"1", "2", "3"}
	positions := []string{"3", "2", "1"}

	values = SortIntSPosition(names, positions, values)
	require.Equal(t, values, []int{15, 10, 1})

	values = []int{1, 2, 3, 4}
	names = []string{"1", "2", "3", "4"}
	positions = []string{"3", "2", "4", "1"}

	values = SortIntSPosition(names, positions, values)
	require.Equal(t, values, []int{3, 2, 4, 1})

}
