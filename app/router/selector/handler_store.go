// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package selector

import (
	"sort"
	"strconv"
	"sync"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/xsqlite"
)

// HandlerStore holds all xsqlite.OutboundHandler rows for use when the
// remote DB is unavailable.
type HandlerStore struct {
	mu       sync.RWMutex
	handlers map[int]*xsqlite.OutboundHandler
	groups   map[int][]string
}

func NewHandlerStore() *HandlerStore {
	return &HandlerStore{
		handlers: make(map[int]*xsqlite.OutboundHandler),
		groups:   make(map[int][]string),
	}
}

func (s *HandlerStore) GetHandler(id int) *xsqlite.OutboundHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handlers[id]
}

func (s *HandlerStore) GetAllHandlers() ([]*xsqlite.OutboundHandler, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*xsqlite.OutboundHandler, 0, len(s.handlers))
	for _, h := range s.handlers {
		out = append(out, h)
	}
	return out, nil
}

func (s *HandlerStore) GetHandlersByGroup(group string) ([]*xsqlite.OutboundHandler, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*xsqlite.OutboundHandler
	for id, tags := range s.groups {
		for _, tag := range tags {
			if tag == group {
				if h := s.handlers[id]; h != nil {
					out = append(out, h)
				}
				break
			}
		}
	}
	return out, nil
}

func (s *HandlerStore) GetBatchedHandlers(batchSize int, offset int) ([]*xsqlite.OutboundHandler, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]int, 0, len(s.handlers))
	for id := range s.handlers {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	if offset >= len(ids) {
		return nil, nil
	}
	end := offset + batchSize
	if end > len(ids) {
		end = len(ids)
	}
	out := make([]*xsqlite.OutboundHandler, 0, end-offset)
	for _, id := range ids[offset:end] {
		out = append(out, s.handlers[id])
	}
	return out, nil
}

func (s *HandlerStore) Set(handlers []*xsqlite.OutboundHandler, groups map[int][]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, h := range handlers {
		if h == nil {
			continue
		}
		s.handlers[h.ID] = h
		if tags, ok := groups[h.ID]; ok {
			s.groups[h.ID] = append([]string(nil), tags...)
		}
	}
}

func (s *HandlerStore) Remove(ids []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range ids {
		delete(s.handlers, id)
		delete(s.groups, id)
	}
}

func (s *HandlerStore) Replace(handlers []*xsqlite.OutboundHandler, groups map[int][]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = make(map[int]*xsqlite.OutboundHandler, len(handlers))
	s.groups = make(map[int][]string, len(groups))
	for _, h := range handlers {
		if h == nil {
			continue
		}
		s.handlers[h.ID] = h
	}
	for id, tags := range groups {
		s.groups[id] = append([]string(nil), tags...)
	}
}

func (s *HandlerStore) ReplaceFromProtos(rows []*configs.OutboundHandler) {
	handlers, groups := fromProtos(rows)
	s.Replace(handlers, groups)
}

func (s *HandlerStore) SetFromProtos(rows []*configs.OutboundHandler) {
	handlers, groups := fromProtos(rows)
	s.Set(handlers, groups)
}

func (s *HandlerStore) RemoveTags(tags []string) {
	ids := make([]int, 0, len(tags))
	for _, tag := range tags {
		id, err := strconv.Atoi(tag)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	s.Remove(ids)
}

func fromProtos(rows []*configs.OutboundHandler) ([]*xsqlite.OutboundHandler, map[int][]string) {
	handlers := make([]*xsqlite.OutboundHandler, 0, len(rows))
	groups := make(map[int][]string)
	for _, row := range rows {
		h := protoToXsqlite(row)
		if h == nil {
			continue
		}
		handlers = append(handlers, h)
		if tags := row.GetGroupTags(); len(tags) > 0 {
			groups[h.ID] = append([]string(nil), tags...)
		}
	}
	return handlers, groups
}

func protoToXsqlite(row *configs.OutboundHandler) *xsqlite.OutboundHandler {
	if row == nil || row.GetId() == 0 {
		return nil
	}
	h := &xsqlite.OutboundHandler{
		ID:               int(row.GetId()),
		Selected:         row.GetSelected(),
		CountryCode:      row.GetCountryCode(),
		Ok:               int(row.GetOk()),
		Speed:            row.GetSpeed(),
		SpeedTestTime:    int(row.GetSpeedTestTime()),
		Ping:             int(row.GetPing()),
		PingTestTime:     int(row.GetPingTestTime()),
		Config:           row.GetConfig(),
		Sni:              row.GetSni(),
		ServerIp:         row.GetServerIp(),
		Support6:         int(row.GetSupport6()),
		Support6TestTime: int(row.GetSupport6TestTime()),
	}
	if row.SubId != nil {
		sid := int(row.GetSubId())
		h.SubId = &sid
	}
	return h
}
