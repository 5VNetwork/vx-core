package system

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type nat struct {
	sync.RWMutex
	srcToSession     map[SrcDst]*natSession
	natPortToSession map[uint16]*natSession
	natport          uint16
	// period           task.Periodic
}

func NewNat() *nat {
	n := &nat{
		srcToSession:     make(map[SrcDst]*natSession),
		natPortToSession: make(map[uint16]*natSession),
		natport:          10000,
	}
	return n
}

// every tcp connection has a session
type natSession struct {
	src     netip.AddrPort
	dst     netip.AddrPort
	natPort uint16
	t       time.Time
	ctx     context.Context
	cancel  context.CancelFunc
}

type SrcDst struct {
	src netip.AddrPort
	dst netip.AddrPort
}

// find a nat session, if not found, create a new one
func (n *nat) getNatSession(src netip.AddrPort, dst netip.AddrPort) (*natSession, bool) {
	n.RLock()
	defer n.RUnlock()
	s, found := n.srcToSession[SrcDst{src, dst}]
	return s, found
}

func (n *nat) createNatSession(src netip.AddrPort, dst netip.AddrPort) *natSession {
	n.Lock()
	defer n.Unlock()
	// reset nextPort
	natportCache := n.natport
	for {
		n.natport++
		if n.natport > 65534 {
			n.natport = 10000
		} else if n.natport == natportCache {
			log.Error().Msg("no available nat port")
			n.Unlock()
			time.Sleep(time.Second * 1)
			n.Lock()
		} else if n.natPortToSession[n.natport] == nil {
			break
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &natSession{src: src, dst: dst, natPort: n.natport, t: time.Now(), ctx: ctx, cancel: cancel}
	n.natPortToSession[s.natPort] = s
	n.srcToSession[SrcDst{src, dst}] = s
	return s

}

func (n *nat) findNatSessionByNatport(natPort uint16) *natSession {
	n.RLock()
	defer n.RUnlock()
	return n.natPortToSession[natPort]
}

func (n *nat) removeNatSession(s *natSession) {
	n.Lock()
	delete(n.srcToSession, SrcDst{s.src, s.dst})
	delete(n.natPortToSession, s.natPort)
	n.Unlock()
}
