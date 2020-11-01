package resource

import (
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/failover"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type FailoverSource int

// FailOverMode enumerates the values for upgrade mode.
type FailOverMode string

const (
	FailoverSourceID FailoverSource = iota + 1
	FailoverSourceSchema

	// FailOverModeDistributed ...
	FailOverModeDistributed FailOverMode = "distributed"
	// FailOverModeSingle ...
	FailOverModeSingle FailOverMode = "single"

	TagsFieldName              = "tags"
	InstancesFieldName         = "instances"
	LocationsFieldName         = "locations"
	PrimaryCountFieldName      = "primary_count"
	SecondaryCountFieldName    = "secondary_count"
	TertiaryCountFieldName     = "tertiary_count"
	FailoverInstancesFieldName = "failover_instances"
	FailoverModeFieldName      = "failover_mode"
	PrefixFieldName            = "prefix"
	MetricNameFieldName        = "metric_name"
	MetricNamespaceFieldName   = "metric_namespace"
)

type Failover struct {
	Prefix            string
	FailoverMode      FailOverMode
	MetricName        string
	MetricNameSpace   string
	Instances         []int
	Locations         []string
	PrimaryCount      int
	SecondaryCount    int
	TertiaryCount     int
	FailoverInstances []int
	Source            FailoverSource
}

func (f *Failover) SetPrimaryCount(n int) {
	f.PrimaryCount = n
	f.FailoverInstances[0] = n
}

func (f *Failover) SetSecondaryCount(n int) {
	f.SecondaryCount = n
	f.FailoverInstances[1] = n
}

func (f *Failover) SetTertiaryCount(n int) {
	f.TertiaryCount = n
	f.FailoverInstances[2] = n
}

func (f Failover) IsNotSet() bool {
	return f.PrimaryCount == 0 && f.SecondaryCount == 0 && f.TertiaryCount == 0
}

func (f Failover) Initialized() bool {
	return len(f.Locations) != 0 && f.MetricName != "" && f.MetricNameSpace != ""
}

func (f *Failover) FillDefaultCountsIfNotSet() {
	if f.IsNotSet() {
		if f.IsSingleMode() {
			// get first location for validator
			counts := failover.CalculateInstancesForSingleFailOverMode(f.Instances)
			f.SetCounts(counts...)
		} else {
			f.SetCounts(f.Instances...)
		}
	}
}

func (f Failover) IsSingleMode() bool {
	return f.FailoverMode == FailOverModeSingle
}

func (f Failover) IsDistributedMode() bool {
	return f.FailoverMode == FailOverModeDistributed
}

func (f *Failover) SetCounts(values ...int) {
	switch len(values) {
	case 0:
		f.PrimaryCount = f.Instances[0]
		f.SecondaryCount = f.Instances[1]
		f.TertiaryCount = f.Instances[2]
	case 1:
		f.PrimaryCount = values[0]
		f.SecondaryCount = f.Instances[1]
		f.TertiaryCount = f.Instances[2]
	case 2:
		f.PrimaryCount = values[0]
		f.SecondaryCount = values[1]
		f.TertiaryCount = f.Instances[2]
	default:
		f.PrimaryCount = values[0]
		f.SecondaryCount = values[1]
		f.TertiaryCount = values[2]
	}
	f.FailoverInstances = []int{f.PrimaryCount, f.SecondaryCount, f.TertiaryCount}
}

func (f Failover) InstancesCount() int {
	return f.PrimaryCount + f.SecondaryCount + f.TertiaryCount
}

func (f Failover) SetSchemaValues(d *schema.ResourceData) error {
	if err := d.Set(PrimaryCountFieldName, f.PrimaryCount); err != nil {
		return err
	}
	if err := d.Set(SecondaryCountFieldName, f.SecondaryCount); err != nil {
		return err
	}
	if err := d.Set(TertiaryCountFieldName, f.TertiaryCount); err != nil {
		return err
	}
	if err := d.Set(FailoverInstancesFieldName, f.FailoverInstances); err != nil {
		return err
	}
	return nil
}

func (f Failover) SetSchemaValuesDiag(d *schema.ResourceData) diag.Diagnostics {
	if err := f.SetSchemaValues(d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func (f *Failover) FromSchema(d *schema.ResourceData) error {

	f.FailoverMode = FailOverMode(d.Get(FailoverModeFieldName).(string))

	instanceLocationsRaw := d.Get(LocationsFieldName).([]interface{})
	f.Locations = ExpandString(instanceLocationsRaw)

	instancesRaw := d.Get(InstancesFieldName).([]interface{})
	f.Instances = ExpandInt(instancesRaw)

	f.Prefix = d.Get(PrefixFieldName).(string)
	f.MetricName = d.Get(MetricNameFieldName).(string)
	f.MetricNameSpace = d.Get(MetricNamespaceFieldName).(string)

	f.PrimaryCount = d.Get(PrimaryCountFieldName).(int)
	f.SecondaryCount = d.Get(SecondaryCountFieldName).(int)
	f.TertiaryCount = d.Get(TertiaryCountFieldName).(int)

	f.Source = FailoverSourceSchema

	return nil

}
