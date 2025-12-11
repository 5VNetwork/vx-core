package mux

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/task"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

type ClientManager struct {
	clientsAccessLock sync.Mutex
	clients           []*client
	cleanupTask       task.Periodic
	Strategy          ClientStrategy
	handler           i.FlowHandler
}

func (m *ClientManager) Start() error {
	m.cleanupTask = task.Periodic{
		Interval: time.Second * 30,
		Execute:  m.cleanupFunc,
	}
	return m.cleanupTask.Start()
}

func (m *ClientManager) Close() error {
	m.cleanupTask.Close()
	return nil
}

func NewClientManager(strategy ClientStrategy, oh i.FlowHandler) *ClientManager {
	return &ClientManager{
		Strategy: strategy,
		handler:  oh,
	}
}

// a worker is chosen to handle a link
// ctx is used to get outboundInfo and sessionID
func (m *ClientManager) HandleReaderWriter(ctx context.Context, dst net.Destination,
	rw buf.ReaderWriter) error {
	m.clientsAccessLock.Lock()

	var client *client
	var err error
	// find any client that is not full
	for idx, w := range m.clients {
		if !w.IsFull() {
			n := len(m.clients)
			if n > 1 && idx != n-1 {
				m.clients[n-1], m.clients[idx] = m.clients[idx], m.clients[n-1] //exchange position
			}
			client = m.clients[idx]
			break
		}
	}

	// When no client is found, create a new one
	if client == nil {
		client, err = m.Create()
		if err != nil {
			m.clientsAccessLock.Unlock()
			return err
		}
		m.clients = append(m.clients, client)
	}

	sm := &clientSession{
		ID:              uint16(client.count.Add(1)),
		errChan:         make(chan error, 1),
		leftToRightDone: done.New(),
		rightToLeftDone: done.New(),
		ctx:             ctx,
		rw:              rw,
	}
	client.AddSession(sm)
	defer client.RemoveSession(sm)
	log.Ctx(ctx).Debug().Uint16("mux_sid", sm.ID).Msg("new mux session")

	m.clientsAccessLock.Unlock()
	defer m.tryRetire(client)

	go client.merge(ctx, dst, sm)

	var leftToRight, rightToLeft bool
	for {
		if leftToRight && rightToLeft {
			return nil
		}
		select {
		case err := <-sm.errChan:
			return err
		case <-sm.leftToRightDone.Wait():
			leftToRight = true
		case <-sm.rightToLeftDone.Wait():
			rightToLeft = true
			return nil
		}
	}
}

func (p *ClientManager) tryRetire(client *client) {
	p.clientsAccessLock.Lock()
	defer p.clientsAccessLock.Unlock()

	if client.IsEmpty() && client.IsClosing() {
		for i, w := range p.clients {
			if w == client {
				p.clients = append(p.clients[:i], p.clients[i+1:]...)
				break
			}
		}
		client.Close()
	}
}

func (p *ClientManager) Create() (*client, error) {
	iLink, oLink := pipe.NewLinks(64*1024, false)

	logger := log.With().Uint32("sid", rand.Uint32()).Logger()
	ctx, cancelCause := context.WithCancelCause(logger.WithContext(context.Background()))
	log.Ctx(ctx).Debug().Msg("new mux client")

	c, _ := NewClient(ctx, iLink, p.Strategy)

	go func() {
		err := p.handler.HandleFlow(ctx, net.TCPDestination(MuxCoolAddressDst, MuxCoolPortDst), oLink)
		if err != nil {
			log.Ctx(ctx).Err(err).Msg("mux client end with error")
		}
		c.interrupt()
		cancelCause(err)
		p.deleteClient(c)
	}()

	return c, nil
}

func (p *ClientManager) deleteClient(c *client) {
	p.clientsAccessLock.Lock()
	defer p.clientsAccessLock.Unlock()

	for i, w := range p.clients {
		if w == c {
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
			break
		}
	}
	c.Close()
}

func (p *ClientManager) cleanupFunc() error {
	p.clientsAccessLock.Lock()
	defer p.clientsAccessLock.Unlock()

	if len(p.clients) == 0 {
		return nil
	}

	var activeWorkers []*client
	for _, client := range p.clients {
		if client.IsIdle() {
			client.Close()
		} else {
			activeWorkers = append(activeWorkers, client)
		}
	}
	p.clients = activeWorkers
	return nil
}

type ClientStrategy struct {
	MaxConcurrency uint32
	MaxConnection  uint32
}

var DefaultClientStrategy = ClientStrategy{
	MaxConnection:  16,
	MaxConcurrency: 2,
}
