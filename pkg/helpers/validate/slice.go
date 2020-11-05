package validate

import (
	"fmt"
)

// SliceString validates slice of strings
func SliceString(i interface{}, k string) ([]string, []error) {
	values, ok := i.([]interface{})
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be slice", k)}
	}

	for idx, val := range values {
		if str, ok := val.(string); !ok {
			return nil, []error{fmt.Errorf("expected type of %q %d position to be string", k, idx)}
		} else if len(str) == 0 {
			return nil, []error{fmt.Errorf("expected length of %q %d position to be more than 0", k, idx)}
		}
	}
	return nil, nil
}
