package resource

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SetSchemaValues(d *schema.ResourceData, diagnostics diag.Diagnostics, primaryCount, secondaryCount, tertiaryCount int) diag.Diagnostics {

	if diagnostics == nil {
		diagnostics = make(diag.Diagnostics, 0)
	}

	if err := d.Set("primary_count", primaryCount); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	if err := d.Set("secondary_count", secondaryCount); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	if err := d.Set("tertiary_count", tertiaryCount); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	if err := d.Set("failover_instances", []int{primaryCount, secondaryCount, tertiaryCount}); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	return diagnostics
}

func ExpandInt(values []interface{}) []int {
	results := make([]int, 0, len(values))
	for _, value := range values {
		results = append(results, value.(int))
	}
	return results
}

func ExpandString(values []interface{}) []string {
	results := make([]string, 0, len(values))
	for _, value := range values {
		results = append(results, value.(string))
	}
	return results
}

func PrepareID(values ...string) string {
	return strings.Join(values, "_")
}
