package selector

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

type Filter interface {
	GetHandlers() ([]outHandler, error)
}

type omFilter struct {
	om           *outbound.Manager
	filterConfig *configs.SelectorConfig_Filter
}

func NewOmFilter(filterConfig *configs.SelectorConfig_Filter, om *outbound.Manager) *omFilter {
	return &omFilter{
		filterConfig: filterConfig,
		om:           om,
	}
}

type handlerWithStats struct {
	i.Outbound
	oStats
}

func (h *handlerWithStats) GetHandler() (i.Outbound, error) {
	return h.Outbound, nil
}

func (h *handlerWithStats) Name() string {
	return h.Outbound.Tag()
}

func (f *omFilter) GetHandlers() ([]outHandler, error) {
	handlers := f.om.GetAllHandlers()
	var ret []outHandler
	for _, h := range handlers {
		if !inSubset(h, f.filterConfig) {
			continue
		}

		o := oStats{}
		if handlerWithSupport6Info, ok := h.(*outbound.HandlerWithSupport6Info); ok {
			if handlerWithSupport6Info.Support6() {
				o.support6 = 1
			} else {
				o.support6 = -1
			}
		}
		ret = append(ret, &handlerWithStats{
			Outbound: h,
			oStats:   o,
		})
	}
	return ret, nil
}

func inSubset(h i.Outbound, filterConfig *configs.SelectorConfig_Filter) bool {
	for _, prefix := range filterConfig.Prefixes {
		if strings.HasPrefix(h.Tag(), prefix) {
			return !filterConfig.Inverse
		}
	}
	for _, t := range filterConfig.Tags {
		if t == h.Tag() {
			return !filterConfig.Inverse
		}
	}
	return filterConfig.Inverse
}

type dbFilter struct {
	db            Db
	filterConfig  *configs.SelectorConfig_Filter
	landHandlers  []*xsqlite.OutboundHandler
	createHandler CreateHandlerFunc
}

func NewDbFilter(db Db, filterConfig *configs.SelectorConfig_Filter,
	landHandlers []*xsqlite.OutboundHandler, createHandler CreateHandlerFunc) *dbFilter {
	return &dbFilter{
		db:            db,
		filterConfig:  filterConfig,
		landHandlers:  landHandlers,
		createHandler: createHandler,
	}
}

func (f *dbFilter) GetHandlers() ([]outHandler, error) {
	handlers, err := f.getHandlersRetry()
	if err != nil {
		return nil, err
	}
	ret := make([]outHandler, 0, len(handlers))
	for _, h := range handlers {
		o := oStats{}
		if len(f.landHandlers) > 0 {
			o.support6 = f.landHandlers[len(f.landHandlers)-1].GetSupport6()
		} else {
			o.support6 = h.GetSupport6()
			o.ok = h.Ok
			o.speed = int(units.MbpsToBytes(h.GetSpeed()))
			o.ping = h.GetPing()
		}
		ret = append(ret, &dbHandler{
			id:            h.ID,
			oStats:        o,
			db:            f.db,
			landHandlers:  f.landHandlers,
			createHandler: f.createHandler,
		})
	}
	return ret, nil
}

type dbHandler struct {
	id int
	oStats
	db            Db
	landHandlers  []*xsqlite.OutboundHandler
	createHandler CreateHandlerFunc
}

func (h *dbHandler) GetHandler() (i.Outbound, error) {
	handler := h.db.GetHandler(h.id)
	if handler == nil {
		return nil, errors.New("handler not found")
	}
	return h.createHandler(handler.ToConfig(), h.landHandlers)
}

func (h *dbHandler) Name() string {
	if len(h.landHandlers) == 0 {
		return strconv.Itoa(h.id)
	} else {
		name := strconv.Itoa(h.id)
		for _, landHandler := range h.landHandlers {
			name = name + "-" + strconv.Itoa(landHandler.ID)
		}
		return name
	}
}

func (h *dbHandler) IsChain() bool {
	return len(h.landHandlers) > 0
}

func (f *dbFilter) getHandlersRetry() ([]*xsqlite.OutboundHandler, error) {
	log.Debug().Msg("GetHandlers")
	for i := 0; i < 3; i++ {
		handlers, err := f.getHandlers()
		if err != nil {
			log.Error().Err(err).Msg("get handlers")
			time.Sleep(time.Millisecond * 100)
			continue
		}
		return handlers, nil
	}
	return nil, errors.New("cannot get handlers")
}

func (f *dbFilter) getHandlers() ([]*xsqlite.OutboundHandler, error) {
	handlers := make(map[int]*xsqlite.OutboundHandler)

	if f.filterConfig.All {
		// var handlers []*xsqlite.OutboundHandler
		// if err := f.db.Find(&handlers).Error; err != nil {
		// 	return nil, fmt.Errorf("get all handlers: %w", err)
		// }
		// return handlers, nil
		return f.db.GetAllHandlers()
	}

	// group
	for _, group := range f.filterConfig.GroupTags {
		// var hs []*xsqlite.OutboundHandler
		// err := f.db.Joins("JOIN outbound_handler_group_relations ON outbound_handlers.id = outbound_handler_group_relations.handler_id").
		// 	Where("outbound_handler_group_relations.group_name = ?", group).
		// 	Find(&hs).Error
		// if err != nil {
		// 	return nil, fmt.Errorf("get handlers by group: %w", err)
		// }
		hs, err := f.db.GetHandlersByGroup(group)
		if err != nil {
			return nil, fmt.Errorf("get handlers by group: %w", err)
		}
		for _, h := range hs {
			handlers[h.ID] = h
		}
	}

	batchSize := 100
	i := 0
	for {
		// var hs []*xsqlite.OutboundHandler
		// err := f.db.Order("id ASC").Limit(batchSize).Offset(i * batchSize).Find(&hs).Error
		// if err != nil {
		// 	return nil, fmt.Errorf("get batch handlers: %w", err)
		// }
		hs, err := f.db.GetBatchedHandlers(batchSize, i*batchSize)
		if err != nil {
			return nil, fmt.Errorf("get batch handlers: %w", err)
		}
		if len(hs) == 0 {
			break
		}
	F:
		for _, h := range hs {
			tag := h.GetTag()
			for _, prefix := range f.filterConfig.GetPrefixes() {
				if strings.HasPrefix(tag, prefix) {
					handlers[h.ID] = h
					continue F
				}
			}
			for _, filterTag := range f.filterConfig.GetTags() {
				if tag == filterTag {
					handlers[h.ID] = h
					continue F
				}
			}
			for _, handlerId := range f.filterConfig.GetHandlerIds() {
				if h.ID == int(handlerId) {
					handlers[h.ID] = h
					continue F
				}
			}
			for _, subId := range f.filterConfig.GetSubIds() {
				if h.SubId != nil && *h.SubId == int(subId) {
					handlers[h.ID] = h
					continue F
				}
			}
			if f.filterConfig.Selected && h.Selected {
				handlers[h.ID] = h
				continue
			}
			// if f.filterConfig.GetInverse() {
			// 	handlers[h.ID] = h
			// }
		}
		i++
		if len(hs) < batchSize {
			break
		}
	}

	handlersRet := make([]*xsqlite.OutboundHandler, 0, len(handlers))
	for _, h := range handlers {
		handlersRet = append(handlersRet, h)
	}
	return handlersRet, nil
}
