package helpers

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/go-multierror"
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
	t.Log("Setting terrafrom deferred cleanup...")
	t.Cleanup(func() {
		cleanup()
	})
}

// SetInitialTFCleanUp schedule initial terraform clean up
func SetInitialTFCleanUp(t *testing.T, opts *terraform.Options) {
	if _, ok := os.LookupEnv("POLKADOT_TEST_INITIAL_TF_CLEANUP"); ok {
		t.Log("Starting terrafrom cleanup...")
		terraform.Destroy(t, opts)
	} else {
		t.Log("Skipping initial terrafrom cleanup...")
	}
}

// WaitOnErrorChannel waits till channel closed
func WaitOnErrorChannel(ch chan error, wg *sync.WaitGroup) error {

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	var result *multierror.Error

	for err := range ch {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()

}

// GetPrefix returns prefix for resources
func GetPrefix(prefix string) string {
	return prefix + "-"
}

// LastPartOnSplit splits string on delimiter and returns last part
func LastPartOnSplit(s, delimiter string) string {
	return s[strings.LastIndex(s, delimiter)+1:]
}

// FilterStrings filters in place strings slice
func FilterStrings(items *[]string, handler func(item string) bool) {

	start := 0
	for i := start; i < len(*items); i++ {
		if !handler((*items)[i]) {
			// vm will be deleted
			continue
		}
		if i != start {
			(*items)[start], (*items)[i] = (*items)[i], (*items)[start]
		}
		start++
	}

	*items = (*items)[:start]

}
