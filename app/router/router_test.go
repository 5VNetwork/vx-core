package router_test

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/outbound"
	. "github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/common"
	net1 "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
)

func TestSimpleRouter(t *testing.T) {
	om := outbound.NewManager()
	h := outbound.NewProxyHandler(outbound.ProxyHandlerSettings{
		Tag: "test",
	})
	om.AddHandlers(h)

	router, _ := NewRouter(&RouterConfig{
		OutboundManager: om,
	})
	router.AddRule(NewRule(
		"test",
		"test", "",
		NewNetworkMatcher([]net1.Network{net1.Network_TCP}),
	))

	si := &session.Info{
		Target: net1.TCPDestination(net1.DomainAddress("a.b"), 80),
	}
	handler, _ := router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
}

func TestAppIdRouter(t *testing.T) {
	om := outbound.NewManager()
	h2 := outbound.NewProxyHandler(outbound.ProxyHandlerSettings{
		Tag: "test2",
	})
	om.AddHandlers(h2)
	h := outbound.NewProxyHandler(outbound.ProxyHandlerSettings{
		Tag: "test",
	})
	om.AddHandlers(h)

	router, err := NewRouter(&RouterConfig{
		RouterConfig: &configs.RouterConfig{
			Rules: []*configs.RuleConfig{
				{
					OutboundTag: "test",
					AppIds: []*configs.AppId{
						{
							Type:  configs.AppId_Prefix,
							Value: "C:\\abc",
						},
						{
							Type:  configs.AppId_Exact,
							Value: "C:\\aasdfc\\b.exe",
						},
						{
							Type:  configs.AppId_Keyword,
							Value: "bbbb",
						},
					},
				},
			},
		},
		OutboundManager: om,
	})
	common.Must(err)
	// prefix
	si := &session.Info{
		AppId: "C:\\abc\\a.exe",
	}
	handler, _ := router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
	// case insensitive
	si = &session.Info{
		AppId: "C:\\ABC\\b.exe",
	}
	handler, _ = router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
	// exact
	si = &session.Info{
		AppId: "C:\\aasdfc\\b.exe",
	}
	handler, _ = router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
	si = &session.Info{
		AppId: "c:\\AASDFC\\b.exe",
	}
	handler, _ = router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
	// keyword
	si = &session.Info{
		AppId: "C:\\bbbb\\b.exe",
	}
	handler, _ = router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
	si = &session.Info{
		AppId: "C:\\BBBB\\b.exe",
	}
	handler, _ = router.PickHandler(context.Background(), si)
	if handler == nil || handler.Tag() != "test" {
		t.Error("expect tag 'test', bug actually ", handler)
	}
}
