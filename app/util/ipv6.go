package util

import (
	"context"
	gonet "net"
	"slices"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

const (
	GoogleDNS6 = "2001:4860:4860::8888"
	AliDNS6    = "2400:3200::1"
)

func TestIpv6(ctx context.Context, h i.Outbound, address string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	hd := &FlowHandlerToDialer{FlowHandler: h}
	conn, err := hd.Dial(ctx, net.Destination{
		Network: net.Network_TCP,
		Address: net.ParseAddress(address),
		Port:    53,
	})
	if err != nil {
		return false, err
	}
	defer conn.Close()

	dnsConn := &dns.Conn{
		Conn: conn,
	}
	tcpClient := &dns.Client{
		Net: "tcp",
	}
	msg := &dns.Msg{}
	msg.SetQuestion("www.apple.com.", dns.TypeAAAA)
	_, _, err = tcpClient.ExchangeWithConnContext(ctx, msg, dnsConn)
	if err != nil {
		log.Debug().Err(err).Msg("test ipv6 ExchangeWithConnContext failed")
		return false, nil
	}
	return true, nil
}

func NICSupportIPv6Index(index uint32) bool {
	log.Debug().Uint32("index", index).Msg("NICSupportIPv6 start")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := dlhelper.DialSystemConn(ctx, net.TCPDestination(
		net.ParseAddress(AliDNS6), 53,
	), &dlhelper.SocketSetting{
		BindToDevice6: uint32(index),
	})
	if err != nil {
		log.Debug().Err(err).Uint32("index", index).Msg("NICSupportIPv6 dial error")
		return false
	}
	conn.Close()
	log.Debug().Uint32("index", index).Msg("NICSupportIPv6 yes")
	return true
}

func NICSupportIPv6Name(nicName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctx = log.Logger.With().Uint("id", uint(session.NewID())).Logger().WithContext(ctx)
	log.Ctx(ctx).Debug().Str("nicName", nicName).Msg("NICSupportIPv6 start")
	conn, err := dlhelper.DialSystemConn(ctx, net.TCPDestination(
		net.ParseAddress(AliDNS6), 53,
	), &dlhelper.SocketSetting{
		BindToDeviceName: nicName,
	})
	if err != nil {
		log.Ctx(ctx).Err(err).Str("nicName", nicName).Msg("NICSupportIPv6 dial error")
		return false
	}
	conn.Close()
	log.Ctx(ctx).Debug().Str("nicName", nicName).Msg("NICSupportIPv6 yes")
	return true
}

func NICHasGlobalIPv6Address(index uint32) (bool, error) {
	iff, err := gonet.InterfaceByIndex(int(index))
	if err != nil {
		return false, err
	}
	addrs, err := iff.Addrs()
	if err != nil {
		return false, err
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			log.Err(err).Msg("failed to parse cidr")
			continue
		}
		if ip.To4() == nil && ip.IsGlobalUnicast() {
			return true, nil
		}
	}
	return false, nil
}

type IPv6SupportChangeNotifier struct {
	lock      sync.RWMutex
	observers []i.IPv6SupportChangeObserver
}

type OnIPv6SupportChangedFunc func()

func (f OnIPv6SupportChangedFunc) OnIPv6SupportChanged() {
	f()
}

func (n *IPv6SupportChangeNotifier) Register(observer i.IPv6SupportChangeObserver) {
	n.lock.Lock()
	n.observers = append(n.observers, observer)
	n.lock.Unlock()
}

func (n *IPv6SupportChangeNotifier) Unregister(observer i.IPv6SupportChangeObserver) {
	n.lock.Lock()
	defer n.lock.Unlock()
	for i, o := range n.observers {
		if o == observer {
			n.observers = slices.Delete(n.observers, i, i+1)
			break
		}
	}
}

func (n *IPv6SupportChangeNotifier) Notify() {
	n.lock.RLock()
	defer n.lock.RUnlock()
	for _, o := range n.observers {
		go o.OnIPv6SupportChanged()
	}
}
