// Contains functions that don't really belong anywhere else.

package google

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TerraformResourceData interface {
	HasChange(string) bool
	GetOkExists(string) (interface{}, bool)
	GetOk(string) (interface{}, bool)
	Get(string) interface{}
	Set(string, interface{}) error
	SetId(string)
	Id() string
}

type TerraformResourceDiff interface {
	HasChange(string) bool
	GetChange(string) (interface{}, interface{})
	Get(string) interface{}
	Clear(string) error
	ForceNew(string) error
}

// getProject reads the "project" field from the given resource data and falls
// back to the provider's value if not given. If the provider's value is not
// given, an error is returned.
func getProject(d TerraformResourceData, config *Config) (string, error) {
	return getProjectFromSchema("project", d, config)
}

func mergeResourceMaps(ms ...map[string]*schema.Resource) (map[string]*schema.Resource, error) {
	merged := make(map[string]*schema.Resource)
	var duplicates []string

	for _, m := range ms {
		for k, v := range m {
			if _, ok := merged[k]; ok {
				duplicates = append(duplicates, k)
			}

			merged[k] = v
		}
	}

	var err error
	if len(duplicates) > 0 {
		err = fmt.Errorf("saw duplicates in mergeResourceMaps: %v", duplicates)
	}

	return merged, err
}

func generateUserAgentString(d *schema.ResourceData, currentUserAgent string) (string, error) {
	var m providerMeta

	err := d.GetProviderMeta(&m)
	if err != nil {
		return currentUserAgent, err
	}

	if m.ModuleName != "" {
		return strings.Join([]string{currentUserAgent, m.ModuleName}, " "), nil
	}

	return currentUserAgent, nil
}
