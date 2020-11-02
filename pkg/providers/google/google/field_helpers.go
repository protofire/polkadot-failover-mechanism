package google

import (
	"fmt"
	"regexp"
)

const (
	globalLinkTemplate       = "projects/%s/global/%s/%s"
	globalLinkBasePattern    = "projects/(.+)/global/%s/(.+)"
	zonalLinkTemplate        = "projects/%s/zones/%s/%s/%s"
	regionalLinkTemplate     = "projects/%s/regions/%s/%s/%s"
	projectLinkTemplate      = "projects/%s/%s/%s"
	organizationLinkTemplate = "organizations/%s/%s/%s"
)

// ------------------------------------------------------------
// Field helpers
// ------------------------------------------------------------

func ParseNetworkFieldValue(network string, d TerraformResourceData, config *Config) (*GlobalFieldValue, error) {
	return parseGlobalFieldValue("networks", network, "project", d, config, true)
}

// ------------------------------------------------------------
// Base helpers used to create helpers for specific fields.
// ------------------------------------------------------------

type GlobalFieldValue struct {
	Project string
	Name    string

	resourceType string
}

func (f GlobalFieldValue) RelativeLink() string {
	if len(f.Name) == 0 {
		return ""
	}

	return fmt.Sprintf(globalLinkTemplate, f.Project, f.resourceType, f.Name)
}

// Parses a global field supporting 5 different formats:
// - https://www.googleapis.com/compute/ANY_VERSION/projects/{my_project}/global/{resource_type}/{resource_name}
// - projects/{my_project}/global/{resource_type}/{resource_name}
// - global/{resource_type}/{resource_name}
// - resource_name
// - "" (empty string). RelativeLink() returns empty if isEmptyValid is true.
//
// If the project is not specified, it first tries to get the project from the `projectSchemaField` and then fallback on the default project.
func parseGlobalFieldValue(resourceType, fieldValue, projectSchemaField string, d TerraformResourceData, config *Config, isEmptyValid bool) (*GlobalFieldValue, error) {
	if len(fieldValue) == 0 {
		if isEmptyValid {
			return &GlobalFieldValue{resourceType: resourceType}, nil
		}
		return nil, fmt.Errorf("the global field for resource %s cannot be empty", resourceType)
	}

	r := regexp.MustCompile(fmt.Sprintf(globalLinkBasePattern, resourceType))
	if parts := r.FindStringSubmatch(fieldValue); parts != nil {
		return &GlobalFieldValue{
			Project: parts[1],
			Name:    parts[2],

			resourceType: resourceType,
		}, nil
	}

	project, err := getProjectFromSchema(projectSchemaField, d, config)
	if err != nil {
		return nil, err
	}

	return &GlobalFieldValue{
		Project: project,
		Name:    GetResourceNameFromSelfLink(fieldValue),

		resourceType: resourceType,
	}, nil
}

type ZonalFieldValue struct {
	Project string
	Zone    string
	Name    string

	resourceType string
}

func (f ZonalFieldValue) RelativeLink() string {
	if len(f.Name) == 0 {
		return ""
	}

	return fmt.Sprintf(zonalLinkTemplate, f.Project, f.Zone, f.resourceType, f.Name)
}

func getProjectFromSchema(projectSchemaField string, d TerraformResourceData, config *Config) (string, error) {
	res, ok := d.GetOk(projectSchemaField)
	if ok && projectSchemaField != "" {
		return res.(string), nil
	}
	if config.Project != "" {
		return config.Project, nil
	}
	return "", fmt.Errorf("%s: required field is not set", projectSchemaField)
}

type OrganizationFieldValue struct {
	// nolint
	OrgId string
	Name  string

	resourceType string
}

func (f OrganizationFieldValue) RelativeLink() string {
	if len(f.Name) == 0 {
		return ""
	}

	return fmt.Sprintf(organizationLinkTemplate, f.OrgId, f.resourceType, f.Name)
}

type RegionalFieldValue struct {
	Project string
	Region  string
	Name    string

	resourceType string
}

func (f RegionalFieldValue) RelativeLink() string {
	if len(f.Name) == 0 {
		return ""
	}

	return fmt.Sprintf(regionalLinkTemplate, f.Project, f.Region, f.resourceType, f.Name)
}

type ProjectFieldValue struct {
	Project string
	Name    string

	resourceType string
}

func (f ProjectFieldValue) RelativeLink() string {
	if len(f.Name) == 0 {
		return ""
	}

	return fmt.Sprintf(projectLinkTemplate, f.Project, f.resourceType, f.Name)
}
