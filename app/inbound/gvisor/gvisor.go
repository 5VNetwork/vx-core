package gvisor

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type TunGvisorInbound struct {
	option *GvisorInboundOption
	stack  *stack.Stack
	closed bool

	sync.RWMutex
	udpSession map[net.Destination]*udpSession0

	ctx    context.Context
	cancel context.CancelCauseFunc
}

type t2 struct {
	uuid uuid.UUID
	num  int
}

type StackOption func(*stack.Stack) error

func NewGvisorInbound(option *GvisorInboundOption) (*TunGvisorInbound, error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	gi := &TunGvisorInbound{
		option:     option,
		udpSession: make(map[net.Destination]*udpSession0),
		ctx:        ctx,
		cancel:     cancel,
	}

	return gi, nil
}

func (gi *TunGvisorInbound) Tag() string {
	return gi.option.Tag
}

func (gi *TunGvisorInbound) Start() error {
	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
	})
	nicID := tcpip.NICID(s.NextNICID())
	log.Info().Int("nicID", int(nicID)).Send()

	opts := []StackOption{}
	if gi.option.TcpOnly {
		opts = append(opts, SetTCPHandler(gi))
	} else if gi.option.UdpOnly {
		opts = append(opts, SetUDPHandler(gi))
	} else {
		opts = append(opts, SetTCPHandler(gi), SetUDPHandler(gi))
	}
	opts = append(opts,
		CreateNIC(nicID, gi.option.LinkEndpoint),
		SetRouteTable(nicID),
		SetPromiscuousMode(nicID),
		SetSpoofing(nicID),
	) // 256KB send buffer
	if runtime.GOOS == "ios" {
		opts = append(opts, // Set TCP buffer sizes (adjust as needed)
			SetTCPReceiveBufferSize(64*1024, 256*1024), // 256KB receive buffer
			SetTCPSendBufferSize(64*1024, 256*1024))
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}
	gi.stack = s
	return common.Start(gi.option.LinkEndpoint)
}

func (gi *TunGvisorInbound) Close() error {
	if gi.closed {
		return nil
	}
	gi.closed = true
	if gi.option.OnClose != nil {
		gi.option.OnClose()
	}
	gi.cancel(errors.ErrClosed)
	gi.option.LinkEndpoint.Attach(nil)
	gi.stack.Close()
	log.Debug().Msg("stack closed")
	for _, endpoint := range gi.stack.CleanupEndpoints() {
		endpoint.Abort()
	}
	log.Debug().Msg("endpoints closed")
	// gi.stack.Destroy()
	gi.option.LinkEndpoint.Close()
	log.Debug().Msg("link endpoint closed")
	return nil
}

func (gi *TunGvisorInbound) CloseBlocking() error {
	if gi.closed {
		return nil
	}
	gi.closed = true
	if gi.option.OnClose != nil {
		gi.option.OnClose()
	}
	gi.cancel(errors.ErrClosed)

	gi.option.LinkEndpoint.Attach(nil)
	gi.stack.Close()
	log.Debug().Msg("stack closed")
	for _, endpoint := range gi.stack.CleanupEndpoints() {
		endpoint.Abort()
	}
	log.Debug().Msg("endpoints closed")
	gi.stack.Wait()
	log.Debug().Msg("stack wait done")
	// gi.stack.Destroy()
	gi.option.LinkEndpoint.Close()
	log.Debug().Msg("link endpoint closed")
	return nil
}

func (gi *TunGvisorInbound) WithHandler(h i.Handler) {
	gi.option.Handler = h
}

type GvisorInboundOption struct {
	// Fd int
	// Name string
	Tag string
	i.Handler
	stack.LinkEndpoint
	TcpOnly bool
	UdpOnly bool
	OnClose func()
	//TODO: full cone
	// UdpFullCone bool
}

// func (gi *TunGvisorInbound) GetUuid(src net.Destination) *t2 {
// 	gi.Lock()
// 	defer gi.Unlock()
// 	u, found := gi.udpSrcToUuid[src]
// 	if !found {
// 		u = &t2{
// 			uuid: uuid.New(),
// 			num:  0,
// 		}
// 		gi.udpSrcToUuid[src] = u
// 	}
// 	u.num++
// 	return u
// }

func CreateNIC(id tcpip.NICID, linkEndpoint stack.LinkEndpoint) StackOption {
	return func(s *stack.Stack) error {
		if err := s.CreateNICWithOptions(id, linkEndpoint,
			stack.NICOptions{
				Disabled: false,
				QDisc:    nil,
			}); err != nil {
			return fmt.Errorf("failed to create NIC: %s", err.String())
		}
		return nil
	}
}

func SetPromiscuousMode(id tcpip.NICID) StackOption {
	return func(s *stack.Stack) error {
		if err := s.SetPromiscuousMode(id, true); err != nil {
			return fmt.Errorf("failed to set promiscuous mode: %s", err.String())
		}
		return nil
	}
}

func SetSpoofing(id tcpip.NICID) StackOption {
	return func(s *stack.Stack) error {
		if err := s.SetSpoofing(id, true); err != nil {
			return fmt.Errorf("failed to set spoofing: %s", err.String())
		}
		return nil
	}
}

func AddProtocolAddress(id tcpip.NICID, ips []*geo.CIDR) StackOption {
	return func(s *stack.Stack) error {
		for _, ip := range ips {
			tcpIPAddr := tcpip.AddrFrom4Slice(ip.Ip)
			protocolAddress := tcpip.ProtocolAddress{
				AddressWithPrefix: tcpip.AddressWithPrefix{
					Address:   tcpIPAddr,
					PrefixLen: int(ip.Prefix),
				},
			}

			switch tcpIPAddr.Len() {
			case 4:
				protocolAddress.Protocol = ipv4.ProtocolNumber
			case 16:
				protocolAddress.Protocol = ipv6.ProtocolNumber
			default:
				return fmt.Errorf("invalid ip address: %s", tcpIPAddr)
			}

			if err := s.AddProtocolAddress(id, protocolAddress, stack.AddressProperties{}); err != nil {
				return fmt.Errorf("failed to add protocol address: %s", err.String())
			}
		}

		return nil
	}
}

func SetRouteTable(id tcpip.NICID) StackOption {
	return func(s *stack.Stack) error {
		s.SetRouteTable([]tcpip.Route{
			{
				Destination: header.IPv4EmptySubnet,
				NIC:         id,
			},
			{
				Destination: header.IPv6EmptySubnet,
				NIC:         id,
			},
		})

		return nil
	}
}

func SetTCPSendBufferSize(defaultSize, maxSize int) StackOption {
	return func(s *stack.Stack) error {
		sendBufferSizeRangeOption := tcpip.TCPSendBufferSizeRangeOption{Min: tcp.MinBufferSize, Default: defaultSize, Max: maxSize}
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &sendBufferSizeRangeOption); err != nil {
			return fmt.Errorf("failed to set tcp send buffer size: %s", err.String())
		}
		return nil
	}
}

func SetTCPReceiveBufferSize(defaultSize, maxSize int) StackOption {
	return func(s *stack.Stack) error {
		receiveBufferSizeRangeOption := tcpip.TCPReceiveBufferSizeRangeOption{Min: tcp.MinBufferSize, Default: defaultSize, Max: maxSize}
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &receiveBufferSizeRangeOption); err != nil {
			return fmt.Errorf("failed to set tcp receive buffer size: %s", err.String())
		}
		return nil
	}
}
