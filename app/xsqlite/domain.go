package xsqlite

import (
	"github.com/5vnetwork/vx-core/common/geo"
	"github.com/golang/protobuf/proto"
)

func (d *Database) AddGeoDomain(domain string, domainSetName string) error {
	geoDomain := &geo.Domain{
		Value: domain,
		Type:  geo.Domain_Full,
	}
	geoDomainBytes, err := proto.Marshal(geoDomain)
	if err != nil {
		return err
	}
	err = d.DB.Create(&GeoDomain{
		GeoDomain:     geoDomainBytes,
		DomainSetName: domainSetName,
	}).Error
	if err != nil {
		return err
	}
	return nil
}
