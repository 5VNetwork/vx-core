package outbound

import (
	"sync"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

func NewManager() *Manager {
	return &Manager{
		handlers: make(map[string]i.Outbound),
	}
}

const DirectHandlerTag = "direct"
const DnsHandlerTag = "dns"

// holds all outbound handlers. notify its listeners when a change happens.
type Manager struct {
	sync.RWMutex
	handlers map[string]i.Outbound
	// observers of handler changes: add, delete
	handlerObservers []HandlerObserver
}

type HandlerObserver interface {
	OnHandlerChanged()
}

type OnHandlerChangedFunc func()

func (f OnHandlerChangedFunc) OnHandlerChanged() {
	f()
}

func (m *Manager) GetAllHandlers() []i.Outbound {
	m.RLock()
	defer m.RUnlock()
	all := make([]i.Outbound, 0, len(m.handlers))
	for _, handler := range m.handlers {
		all = append(all, handler)
	}
	return all
}

func (m *Manager) Start() error {
	for _, handler := range m.handlers {
		if err := common.Start(handler); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Close() error {
	m.Lock()
	defer m.Unlock()
	for _, handler := range m.handlers {
		if err := common.Close(handler); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetHandler(tag string) i.Outbound {
	m.RLock()
	defer m.RUnlock()
	return m.handlers[tag]
}

func (m *Manager) AddHandlers(handlers ...i.Outbound) error {
	m.Lock()
	defer m.Unlock()
	for _, handler := range handlers {
		m.handlers[handler.Tag()] = handler
	}
	m.log()
	m.notifyHandlerObservers()
	return nil
}

func (m *Manager) log() {
	all := make([]string, 0, len(m.handlers))
	for _, handler := range m.handlers {
		all = append(all, handler.Tag())
	}
	log.Info().Strs("handlers", all).Msg("all outbound handlers")
}

// replace all proxy handlers with new ones
func (m *Manager) ReplaceHandlers(handlers ...i.Outbound) error {
	m.Lock()
	defer m.Unlock()

	handlersMap := make(map[string]i.Outbound)
	dnsHandler, ok := m.handlers[DnsHandlerTag]
	if ok {
		handlersMap[DnsHandlerTag] = dnsHandler
	}
	handlersMap[DirectHandlerTag] = m.handlers[DirectHandlerTag]
	for _, handler := range handlers {
		handlersMap[handler.Tag()] = handler
	}
	m.handlers = handlersMap
	m.log()
	m.notifyHandlerObservers()
	return nil
}

func (m *Manager) RemoveHandlers(tags []string) error {
	m.Lock()
	defer m.Unlock()
	for _, tag := range tags {
		delete(m.handlers, tag)
	}
	if len(m.handlers) == 0 {
		log.Warn().Msg("no handlers now")
	}
	m.log()
	m.notifyHandlerObservers()
	return nil
}

func (m *Manager) AddHandlerObserver(o HandlerObserver) {
	m.Lock()
	defer m.Unlock()
	m.handlerObservers = append(m.handlerObservers, o)
}

func (m *Manager) RemoveHandlerObserver(observer HandlerObserver) {
	m.Lock()
	defer m.Unlock()
	for i, o := range m.handlerObservers {
		if o == observer {
			m.handlerObservers = append(m.handlerObservers[:i], m.handlerObservers[i+1:]...)
			break
		}
	}
}

func (m *Manager) notifyHandlerObservers() {
	for _, o := range m.handlerObservers {
		go o.OnHandlerChanged()
	}
}

// return all handlers except direct and dns
func GetAllProxyhandlers(om i.OutboundManager) []i.Outbound {
	all := om.GetAllHandlers()
	proxyHandlers := make([]i.Outbound, 0, len(all))
	for _, handler := range all {
		if handler.Tag() == DirectHandlerTag || handler.Tag() == DnsHandlerTag {
			continue
		}
		proxyHandlers = append(proxyHandlers, handler)
	}
	return proxyHandlers
}
