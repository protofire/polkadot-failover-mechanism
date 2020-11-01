package polkadot

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/azure"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"
)

const (
	ScaleSetsFieldName        = "scale_sets"
	PrefixFieldName           = "prefix"
	MetricNameFieldName       = "metric_name"
	MetricOutputNameFieldName = "metric_output_name"
	MetricNamespaceFieldName  = "metric_namespace"
	ResourceGroupFieldName    = "resource_group_name"
)

type MetricSource struct {
	ResourceGroup    string
	Prefix           string
	MetricName       string
	MetricOutputName string
	MetricNameSpace  string
	ScaleSets        []string
}

func (m MetricSource) SetSchemaValues(d *schema.ResourceData) error {
	if err := d.Set(MetricOutputNameFieldName, m.MetricOutputName); err != nil {
		return err
	}
	return nil
}

func (m MetricSource) SetSchemaValuesDiag(d *schema.ResourceData) diag.Diagnostics {
	if err := m.SetSchemaValues(d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func (m *MetricSource) FromSchema(d *schema.ResourceData) error {
	scaleSetsRaw := d.Get(ScaleSetsFieldName).([]interface{})
	m.ScaleSets = resource.ExpandString(scaleSetsRaw)
	m.Prefix = d.Get(PrefixFieldName).(string)
	m.MetricName = d.Get(MetricNameFieldName).(string)
	m.MetricNameSpace = d.Get(MetricNamespaceFieldName).(string)
	m.ResourceGroup = d.Get(ResourceGroupFieldName).(string)
	return nil
}

func (m *MetricSource) ID() (string, error) {
	return resource.BsonPack(m)
}

func (m *MetricSource) SetMetric(metric string) {
	m.MetricOutputName = metric
}

type AzureFailover struct {
	resource.Failover
	ResourceGroup string
}

func (f *AzureFailover) FromIDOrSchema(d *schema.ResourceData) error {
	if id := d.Id(); id != "" {
		err := resource.BsonUnPack(f, id)
		if err != nil {
			return err
		}
		f.Source = resource.FailoverSourceID
		return nil
	}
	err := f.FromSchema(d)
	f.Locations = azure.NormalizeSlice(f.Locations)
	f.ResourceGroup = d.Get(ResourceGroupFieldName).(string)
	return err
}

func (f *AzureFailover) FromID(id string) error {
	return resource.BsonUnPack(f, id)
}

func (f *AzureFailover) ID() (string, error) {
	return resource.BsonPack(f)
}
