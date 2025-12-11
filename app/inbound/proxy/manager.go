package proxy

import (
	"errors"
	"fmt"
	"sync"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type Inbound interface {
	Tag() string
	common.Runnable
	UserManage
}

type UserManage interface {
	AddUser(user i.User)
	RemoveUser(uid, secret string)
	WithOnUnauthorizedRequest(f i.UnauthorizedReport)
}

type InboundManager struct {
	access   sync.RWMutex
	handlers map[string]Inbound
	running  bool
}

func NewManager() *InboundManager {
	return &InboundManager{
		handlers: make(map[string]Inbound),
	}
}

func (m *InboundManager) Start() error {
	m.access.Lock()
	defer m.access.Unlock()

	started := make([]Inbound, 0, len(m.handlers))
	for _, inbound := range m.handlers {
		err := inbound.Start()
		if err != nil {
			log.Error().Err(err).Str("tag", inbound.Tag()).
				Msg("failed to start inbound handler")
			common.CloseAll(started)
			return err
		}
		started = append(started, inbound)
	}
	m.running = true
	return nil
}

func (m *InboundManager) Close() error {
	m.access.Lock()
	defer m.access.Unlock()
	m.running = false

	var errorList []error
	for _, handler := range m.handlers {
		if err := handler.Close(); err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) > 0 {
		return fmt.Errorf("failed to close all handlers: %w", errors.Join(errorList...))
	}
	return nil
}

func (m *InboundManager) AddInbound(handler Inbound) error {
	m.access.Lock()
	defer m.access.Unlock()

	if m.running {
		err := handler.Start()
		if err != nil {
			return err
		}
	}

	tag := handler.Tag()
	if existing, ok := m.handlers[tag]; ok {
		err := existing.Close()
		if err != nil {
			log.Warn().Err(err).Str("tag", tag).Msg("failed to close existing inbound")
		}
	}
	m.handlers[tag] = handler

	return nil
}

func (m *InboundManager) GetInbound(tag string) (Inbound, error) {
	m.access.RLock()
	defer m.access.RUnlock()

	handler, found := m.handlers[tag]
	if !found {
		return nil, errors.New("handler not found: " + tag)
	}
	return handler, nil
}

// RemoveInbound implements inbound.Manager.
func (m *InboundManager) RemoveInbound(tag string) error {
	m.access.Lock()
	defer m.access.Unlock()

	if handler, found := m.handlers[tag]; found {
		if err := handler.Close(); err != nil {
			log.Error().Err(err).Str("tag", tag).Msg("failed to close handler")
		}
		delete(m.handlers, tag)
		return nil
	}

	return errors.New("handler not found")
}

func (m *InboundManager) GetInbounds() []Inbound {
	m.access.RLock()
	defer m.access.RUnlock()

	var handlers []Inbound
	for _, handler := range m.handlers {
		handlers = append(handlers, handler)
	}
	return handlers
}
