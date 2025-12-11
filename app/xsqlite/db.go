package xsqlite

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

func (d *Database) Close() error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *Database) GetHandler(id int) *OutboundHandler {
	var handler *OutboundHandler
	err := d.DB.Where("id = ?", id).First(&handler).Error
	if err != nil {
		log.Error().Err(err).Stack().Int("id", id).Msg("GetHandler")
		return nil
	}
	return handler
}

// tag is string of id
func (d *Database) GetHandlerByTag(tag string) *OutboundHandler {
	id, err := strconv.Atoi(tag)
	if err != nil {
		return nil
	}
	return d.GetHandler(id)
}

func (d *Database) UpdateHandler(id int, m map[string]interface{}) error {
	var h *OutboundHandler
	err := d.DB.Where("id = ?", id).First(&h).Error
	if err != nil {
		return err
	}
	err = d.DB.Model(&h).Updates(m).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetHighestSpeedHandler() *OutboundHandler {
	var handler OutboundHandler
	d.DB.Order("speed DESC").First(&handler)
	return &handler
}

func (d *Database) GetAllHandlers() ([]*OutboundHandler, error) {
	var handlers []*OutboundHandler
	if err := d.DB.Find(&handlers).Error; err != nil {
		return nil, err
	}
	return handlers, nil
}

func (d *Database) GetHandlersByGroup(group string) ([]*OutboundHandler, error) {
	var hs []*OutboundHandler
	err := d.DB.Joins("JOIN outbound_handler_group_relations ON outbound_handlers.id = outbound_handler_group_relations.handler_id").
		Where("outbound_handler_group_relations.group_name = ?", group).
		Find(&hs).Error
	if err != nil {
		return nil, fmt.Errorf("get handlers by group: %w", err)
	}
	return hs, nil
}

func (d *Database) GetBatchedHandlers(batchSize int, offset int) ([]*OutboundHandler, error) {
	var hs []*OutboundHandler
	err := d.DB.Order("id ASC").Limit(batchSize).Offset(offset).Find(&hs).Error
	if err != nil {
		return nil, fmt.Errorf("get batched handlers: %w", err)
	}
	return hs, nil
}
