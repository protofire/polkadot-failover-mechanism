package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/protofire/polkadot-failover-mechanism/pkg/helpers/resource"
)

type Failover struct {
	resource.Failover
	Project string
}

func (f *Failover) FromIDOrSchema(d *schema.ResourceData) error {
	if id := d.Id(); id != "" {
		err := resource.BsonUnPack(f, id)
		if err != nil {
			return err
		}
		f.Source = resource.FailoverSourceID
		return nil
	}
	return f.FromSchema(d)

}

func (f *Failover) FromID(id string) error {
	return resource.BsonUnPack(f, id)
}

func (f *Failover) ID() (string, error) {
	return resource.BsonPack(f)
}
