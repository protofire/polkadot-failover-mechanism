package location

import (
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Schema returns the default Schema which should be used for Location fields
// where these are Required and Cannot be Changed
func Schema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateFunc:     EnhancedValidate,
		StateFunc:        StateFunc,
		DiffSuppressFunc: DiffSuppressFunc,
	}
}

// SchemaOptional returns the Schema for a Location field where this can be optionally specified
func SchemaOptional() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		StateFunc:        StateFunc,
		DiffSuppressFunc: DiffSuppressFunc,
	}
}

func SchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}

// Schema returns the Schema which should be used for Location fields
// where these are Required and can be changed
func SchemaWithoutForceNew() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		ValidateFunc:     EnhancedValidate,
		StateFunc:        StateFunc,
		DiffSuppressFunc: DiffSuppressFunc,
	}
}

func DiffSuppressFunc(_, old, new string, _ *schema.ResourceData) bool {
	return Normalize(old) == Normalize(new)
}

func str(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func HashCode(location interface{}) int {
	loc := location.(string)
	return str(Normalize(loc))
}

func StateFunc(location interface{}) string {
	input := location.(string)
	return Normalize(input)
}
