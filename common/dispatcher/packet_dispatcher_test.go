package dispatcher_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	. "github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/test/mocks"
)

func TestPacketDispatcherSameDestination(t *testing.T) {
	var value int64
	dest := net.Destination{
		Address: net.DomainAddress("a.com"),
		Port:    89,
		Network: net.Network_UDP,
	}
	ctx := session.ContextWithInfo(context.Background(), &session.Info{})
	pd := NewPacketDispatcher(ctx, &mocks.LoopbackHandler{})
	pd.SetResponseCallback(func(p *udp.Packet) {
		if p.Payload.String() != "a" {
			t.Error("unexpected payload ", p.Payload.String())
		}
		atomic.AddInt64(&value, -1)
		if p.Source != dest {
			t.Error("unexpected source")
		}
	})

	b := buf.New()
	b.WriteString("a")

	for i := 0; i < 5; i++ {
		pd.DispatchPacket(dest, b)
		atomic.AddInt64(&value, 1)
	}
	time.Sleep(time.Second)
	pd.Close()
	if value != 0 {
		t.Error("value is ", value)
	}
}

func TestPacketDispatcherDifferentDestination(t *testing.T) {
	var value1 int64
	var value2 int64
	dest1 := net.Destination{
		Address: net.DomainAddress("a.com"),
		Port:    89,
		Network: net.Network_UDP,
	}
	dest2 := net.Destination{
		Address: net.DomainAddress("b.com"),
		Port:    89,
		Network: net.Network_UDP,
	}
	ctx := session.ContextWithInfo(context.Background(), &session.Info{})
	pd := NewPacketDispatcher(ctx, &mocks.LoopbackHandler{})
	pd.SetResponseCallback(func(p *udp.Packet) {
		if p.Source.Address == dest1.Address {
			atomic.AddInt64(&value1, -1)
		} else if p.Source.Address == dest2.Address {
			atomic.AddInt64(&value2, -1)
		} else {
			t.Error("unexpected source ", p.Source)
		}
		if p.Payload.String() != "a" {
			t.Error("unexpected payload ", p.Payload.String())
		}
	})

	b := buf.New()
	b.WriteString("a")

	for i := 0; i < 5; i++ {
		pd.DispatchPacket(dest1, b)
		atomic.AddInt64(&value1, 1)
		pd.DispatchPacket(dest2, b)
		atomic.AddInt64(&value2, 1)
	}
	time.Sleep(time.Second)
	if value1 != 0 {
		t.Error("value1 is ", value1)
	}
	if value2 != 0 {
		t.Error("value2 is ", value2)
	}
}
