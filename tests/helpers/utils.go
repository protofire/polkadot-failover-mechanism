package helpers

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

// BuildRegionsParam build strings from regions slice
func BuildRegionsParam(regions ...string) string {
	var res []string
	for _, region := range regions {
		res = append(res, fmt.Sprintf(`"%s"`, region))
	}
	return fmt.Sprintf("[%s]", strings.Join(res, ", "))
}

// SetPostCleanUp schedule final terraform clean up
func SetPostCleanUp(t *testing.T, opts *terraform.Options) {
	if _, ok := os.LookupEnv("POLKADOT_TEST_NO_POST_CLEANUP"); !ok {
		t.Log("Setting terrafrom deferred cleanup...")
		t.Cleanup(func() {
			terraform.Destroy(t, opts)
		})
	} else {
		t.Log("Skipping terrafrom deferred cleanup...")
	}
}

// SetInitialCleanUp schedule initial terraform clean up
func SetInitialCleanUp(t *testing.T, opts *terraform.Options) {
	if _, ok := os.LookupEnv("POLKADOT_TEST_NO_INITIAL_CLEANUP"); !ok {
		t.Log("Starting terrafrom cleanup...")
		terraform.Destroy(t, opts)
	} else {
		t.Log("Skipping terrafrom cleanup...")
	}
}
