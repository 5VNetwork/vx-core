/*

Some of codes are copied from https://github.com/octeep/wireproxy, license below.

Copyright (c) 2022 Wind T.F. Wong <octeep@pm.me>

Permission to use, copy, modify, and distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

*/

package wireguard

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"

	"buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/proxy/wireguard"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/dice"
	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/domain"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/helper"
	"github.com/rs/zerolog/log"
)

// Handler is an outbound connection that silently swallow the entire payload.
type Handler struct {
	HandlerSettings
	net  Tunnel
	bind *netBindClient
	// cached configuration
	endpoints        []netip.Addr
	hasIPv4, hasIPv6 bool
	wgLock           sync.Mutex
	ctx              context.Context
	cancel           context.CancelFunc
}

type HandlerSettings struct {
	Name string
	Conf *wireguard.DeviceConfig
	// used to resolve request addresses
	DnsForRequestAddress i.IPResolver
	DnsForEndpoint       i.IPResolver
	// control how to resolve peer's endpoint
	Strategy domain.DomainStrategy
	Dialer   i.Dialer
}

// New creates a new wireguard handler.
func New(settings HandlerSettings) (*Handler, error) {
	endpoints, hasIPv4, hasIPv6, err := parseEndpoints(settings.Conf)
	if err != nil {
		return nil, err
	}

	str := settings.Conf.SecretKey
	if len(str) != 64 {
		var dat []byte
		str = strings.TrimSuffix(str, "=")
		if strings.ContainsRune(str, '+') || strings.ContainsRune(str, '/') {
			dat, err = base64.RawStdEncoding.DecodeString(str)
		} else {
			dat, err = base64.RawURLEncoding.DecodeString(str)
		}
		if err == nil {
			str = hex.EncodeToString(dat)
		}
	}
	settings.Conf.SecretKey = str

	for _, peer := range settings.Conf.Peers {
		str := peer.PublicKey
		if len(str) != 64 {
			var dat []byte
			str = strings.TrimSuffix(str, "=")
			if strings.ContainsRune(str, '+') || strings.ContainsRune(str, '/') {
				dat, err = base64.RawStdEncoding.DecodeString(str)
			} else {
				dat, err = base64.RawURLEncoding.DecodeString(str)
			}
			if err == nil {
				str = hex.EncodeToString(dat)
			}
		}
		peer.PublicKey = str
	}

	if settings.Conf.Mtu == 0 {
		settings.Conf.Mtu = 1420
	}

	if settings.Conf.Reserved == nil {
		settings.Conf.Reserved = []byte{0x00, 0x00, 0x00}
	}
	settings.Conf.IsClient = true

	ctx, cancel := context.WithCancel(context.Background())
	return &Handler{
		HandlerSettings: settings,
		endpoints:       endpoints,
		hasIPv4:         hasIPv4,
		hasIPv6:         hasIPv6,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

func (h *Handler) Close() (err error) {
	h.cancel()
	go func() {
		h.wgLock.Lock()
		defer h.wgLock.Unlock()

		if h.net != nil {
			_ = h.net.Close()
			h.net = nil
		}
	}()

	return nil
}

func (h *Handler) Tag() string {
	return h.Name
}

func (h *Handler) processWireGuard(dialer i.Dialer) (err error) {
	h.wgLock.Lock()
	defer h.wgLock.Unlock()

	if h.bind != nil && h.bind.dialer == dialer && h.net != nil {
		return nil
	}

	if h.net != nil {
		_ = h.net.Close()
		h.net = nil
	}
	if h.bind != nil {
		_ = h.bind.Close()
		h.bind = nil
	}

	// bind := conn.NewStdNetBind() // TODO: conn.Bind wrapper for dialer
	h.bind = &netBindClient{
		netBind: netBind{
			dns:      h.DnsForEndpoint,
			strategy: h.Strategy,
			workers:  int(h.Conf.NumWorkers),
		},
		ctx:      h.ctx,
		dialer:   dialer,
		reserved: h.Conf.Reserved,
	}
	defer func() {
		if err != nil {
			h.bind.Close()
			h.bind = nil
		}
	}()

	h.net, err = h.makeVirtualTun()
	if err != nil {
		return fmt.Errorf("failed to create virtual tun interface: %w", err)
	}
	return nil
}

// Process implements OutboundHandler.Dispatch().
func (h *Handler) HandleFlow(ctx context.Context, destination net.Destination, rw buf.ReaderWriter) error {
	log.Ctx(ctx).Debug().Str("destination", destination.String()).Msg("wireguard handle flow")
	if err := h.processWireGuard(h.Dialer); err != nil {
		return err
	}

	// resolve dns
	addr := destination.Address
	if addr.Family().IsDomain() {
		var ips []net.IP
		if h.hasIPv4 && h.hasIPv6 {
			ips, _ = h.DnsForRequestAddress.LookupIP(ctx, addr.Domain())
		} else if h.hasIPv4 {
			ips, _ = h.DnsForRequestAddress.LookupIPv4(ctx, addr.Domain())
		} else if h.hasIPv6 {
			ips, _ = h.DnsForRequestAddress.LookupIPv6(ctx, addr.Domain())
		}
		if len(ips) == 0 {
			return errors.New("empty response")
		}
		addr = net.IPAddress(ips[dice.Roll(len(ips))])
	}

	addrPort := netip.AddrPortFrom(toNetIpAddr(addr), destination.Port.Value())

	if destination.Network == net.Network_TCP {
		conn, err := h.net.DialContextTCPAddrPort(ctx, addrPort)
		if err != nil {
			return fmt.Errorf("failed to create TCP connection: %w", err)
		}
		defer conn.Close()

		return helper.Relay(ctx, rw, rw, buf.NewReader(conn), buf.NewWriter(conn))
	} else if destination.Network == net.Network_UDP {
		conn, err := h.net.DialUDPAddrPort(netip.AddrPort{}, addrPort)
		if err != nil {
			return fmt.Errorf("failed to create UDP connection: %w", err)
		}
		defer conn.Close()
		log.Ctx(ctx).Debug().Str("destination", destination.String()).Msg("wireguard handle udp dial ok")

		return helper.Relay(ctx, rw, rw, buf.NewReader(conn), buf.NewWriter(conn))

	} else {
		return errors.New("unsupported network")
	}
}

func (h *Handler) HandlePacketConn(ctx context.Context, dst net.Destination,
	pc udp.PacketReaderWriter) error {
	d := dispatcher.NewPacketDispatcher(ctx, h, dispatcher.WithResponseCallback(func(packet *udp.Packet) {
		pc.WritePacket(packet)
	}))
	defer d.Close()

	for {
		packet, err := pc.ReadPacket()
		if err != nil {
			return err
		}
		err = d.DispatchPacket(packet.Target, packet.Payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to dispatch packet")
		}
	}
}

// creates a tun interface on netstack given a configuration
func (h *Handler) makeVirtualTun() (Tunnel, error) {
	t, err := createTun(h.Conf)(h.endpoints, int(h.Conf.Mtu), nil)
	if err != nil {
		return nil, err
	}

	if err = t.BuildDevice(h.createIPCRequest(), h.bind); err != nil {
		_ = t.Close()
		return nil, err
	}
	return t, nil
}

// serialize the config into an IPC request
func (h *Handler) createIPCRequest() string {
	var request strings.Builder

	request.WriteString(fmt.Sprintf("private_key=%s\n", h.Conf.SecretKey))

	if !h.Conf.IsClient {
		// placeholder, we'll handle actual port listening on Xray
		request.WriteString("listen_port=1337\n")
	}

	for _, peer := range h.Conf.Peers {
		if peer.PublicKey != "" {
			request.WriteString(fmt.Sprintf("public_key=%s\n", peer.PublicKey))
		}

		if peer.PreSharedKey != "" {
			request.WriteString(fmt.Sprintf("preshared_key=%s\n", peer.PreSharedKey))
		}

		address, port, err := net.SplitHostPort(peer.Endpoint)
		if err != nil {
			log.Error().Msgf("failed to split endpoint %s into address and port: %v", peer.Endpoint, err)
		}
		addr := net.ParseAddress(address)
		if addr.Family().IsDomain() {
			ips := domain.GetIPs(context.Background(), addr.Domain(), h.Strategy, h.DnsForEndpoint)
			if len(ips) == 0 {
				log.Error().Msgf("createIPCRequest empty lookup DNS for %s", addr.Domain())
			} else {
				addr = net.IPAddress(ips[0])
			}
		}

		if peer.Endpoint != "" {
			request.WriteString(fmt.Sprintf("endpoint=%s:%s\n", addr, port))
		}

		for _, ip := range peer.AllowedIps {
			request.WriteString(fmt.Sprintf("allowed_ip=%s\n", ip))
		}

		if peer.KeepAlive != 0 {
			request.WriteString(fmt.Sprintf("persistent_keepalive_interval=%d\n", peer.KeepAlive))
		}
	}

	return request.String()[:request.Len()]
}
