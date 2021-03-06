package gcp

import (
	"github.com/hashicorp/go-multierror"
)

type cleanFunc func(project, prefix string, dryRun bool) error

// CleanResources cleans gcp resources
func CleanResources(project, prefix string, dryRun bool) error {

	result := &multierror.Error{}

	functions := []cleanFunc{
		InstanceGroupsClean,
		SMClean,
		HealthCheckClean,
		SAClean,
		NetworkClean,
		NotificationChannelsClean,
		AlertPolicyClean,
		InstanceTemplatesClean,
	}

	for _, function := range functions {
		err := function(project, prefix, dryRun)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
