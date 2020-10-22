package google

import (
	"fmt"
	"regexp"
	"time"

	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/validate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// Copied from the official Google Cloud auto-generated client.
	ProjectRegex         = "(?:(?:[-a-z0-9]{1,63}\\.)*(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?):)?(?:[0-9]{1,19}|(?:[a-z0-9](?:[-a-z0-9]{0,61}[a-z0-9])?))"
	ProjectRegexWildCard = "(?:(?:[-a-z0-9]{1,63}\\.)*(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?):)?(?:[0-9]{1,19}|(?:[a-z0-9](?:[-a-z0-9]{0,61}[a-z0-9])?)|-)"
	RegionRegex          = "[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?"
	SubnetworkRegex      = "[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?"
)

// nolint
func validateRegexp(re string) schema.SchemaValidateDiagFunc {
	return validate.DiagFunc(func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)
		if !regexp.MustCompile(re).MatchString(value) {
			errors = append(errors, fmt.Errorf(
				"%q (%q) doesn't match regexp %q", k, value, re))
		}

		return
	})
}

// nolint
func validateNonNegativeDuration() schema.SchemaValidateDiagFunc {
	return validate.DiagFunc(func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		dur, err := time.ParseDuration(v)
		if err != nil {
			es = append(es, fmt.Errorf("expected %s to be a duration, but parsing gave an error: %s", k, err.Error()))
			return
		}

		if dur < 0 {
			es = append(es, fmt.Errorf("duration %v must be a non-negative duration", dur))
			return
		}

		return
	})
}
