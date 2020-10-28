package helpers

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

// BuildRegionParams build strings from regions slice
func BuildRegionParams(regions ...string) string {
	var res []string
	for _, region := range regions {
		res = append(res, fmt.Sprintf(`"%s"`, region))
	}
	return fmt.Sprintf("[%s]", strings.Join(res, ", "))
}

// SetPostTFCleanUp schedule final terraform clean up
func SetPostTFCleanUp(t *testing.T, cleanup func()) {
	t.Log("Setting terraform deferred cleanup...")
	t.Cleanup(func() {
		cleanup()
	})
}

// SetInitialTFCleanUp schedule initial terraform clean up
func SetInitialTFCleanUp(t *testing.T, opts *terraform.Options) {
	if _, ok := os.LookupEnv("POLKADOT_TEST_INITIAL_TF_CLEANUP"); ok {
		t.Log("Starting terraform cleanup...")
		terraform.Destroy(t, opts)
	} else {
		t.Log("Skipping initial terraform cleanup...")
	}
}
