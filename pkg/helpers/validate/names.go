package validate

import (
	"fmt"
	"strings"
)

// Prefix validates prefix
func Prefix(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected %q to be a string but it wasn't", k))
		return
	}

	// The value must not be empty.
	if strings.TrimSpace(v) == "" {
		errors = append(errors, fmt.Errorf("%q must not be empty", k))
		return
	}

	if strings.HasPrefix(v, "_") {
		errors = append(errors, fmt.Errorf("%q cannot begin with an underscore", k))
	}

	if strings.HasSuffix(v, ".") || strings.HasSuffix(v, "-") {
		errors = append(errors, fmt.Errorf("%q cannot end with an period or dash", k))
	}

	specialCharacters := `\/"[]:|<>+=;,?*@&~!#$%^()_{}'`
	if strings.ContainsAny(v, specialCharacters) {
		errors = append(errors, fmt.Errorf("%q cannot contain the special characters: `%s`", k, specialCharacters))
	}

	return warnings, errors
}
