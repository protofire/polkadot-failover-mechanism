package failover

// CalculateInstancesForSingleFailOverMode calculates per environment count
// in case failover mode is single and there are no any instances deployed yet
func CalculateInstancesForSingleFailOverMode(counts []int) []int {

	maxIdx := 0
	maxCount := counts[0]
	for idx, i := range counts {
		if i > maxCount {
			maxIdx = idx
		}
	}

	result := make([]int, len(counts))
	result[maxIdx] = 1

	return result

}

// CalculateInstanceCountPerRegion returns number of instances per location
func CalculateInstanceCountPerRegion(instanceLocations []string, location string, normFunc func(s string) string) []int {

	counts := make([]int, len(instanceLocations))

	if location == "" {
		return []int{1, 0, 0}
	}

	for i, loc := range instanceLocations {
		if normFunc(loc) == location {
			counts[i] = 1
		} else {
			counts[i] = 0
		}
	}
	return counts
}
