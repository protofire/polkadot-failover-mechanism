package validate

import (
	"fmt"
)

// SliceInt validates slice of integers
func SliceInt(i interface{}, k string) ([]string, []error) {
	values, ok := i.([]interface{})
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be slice", k)}
	}

	for idx, val := range values {
		if _, ok := val.(int); !ok {
			return nil, []error{fmt.Errorf("expected type of %q %d position to be int", k, idx)}
		}
	}
	return nil, nil
}

// SliceString validates slice of strings
func SliceString(i interface{}, k string) ([]string, []error) {
	values, ok := i.([]interface{})
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be slice", k)}
	}

	for idx, val := range values {
		if str, ok := val.(string); !ok {
			return nil, []error{fmt.Errorf("expected type of %q %d position to be int", k, idx)}
		} else if len(str) == 0 {
			return nil, []error{fmt.Errorf("expected length of %q %d position to be more than 0", k, idx)}
		}
	}
	return nil, nil
}
