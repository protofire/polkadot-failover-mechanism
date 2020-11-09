package azure

import "strings"

// Normalize transforms the human readable Azure Region/Location names (e.g. `West US`)
// into the canonical value to allow comparisons between user-code and API Responses
func Normalize(input string) string {
	return strings.Replace(strings.ToLower(input), " ", "", -1)
}

func NormalizeSlice(input []string) []string {
	result := make([]string, 0, len(input))
	for _, value := range input {
		result = append(result, Normalize(value))
	}
	return result
}
