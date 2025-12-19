package gvisor

import (
	"github.com/5vnetwork/vx-core/app/inbound"
	tun_net "github.com/5vnetwork/vx-core/app/inbound/gvisor/net"
	"github.com/5vnetwork/vx-core/app/inbound/inboundcommon"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"

	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const (
	rcvWnd      = 0 // default settings
	maxInFlight = 2 << 10
)

type tcpConn struct {
	*gonet.TCPConn
	id stack.TransportEndpointID
}

func (c *tcpConn) ID() *stack.TransportEndpointID {
	return &c.id
}

func SetTCPHandler(in *TunGvisorInbound) StackOption {
	return func(s *stack.Stack) error {
		tcpForwarder := tcp.NewForwarder(s, rcvWnd, maxInFlight, func(r *tcp.ForwarderRequest) {
			wg := new(waiter.Queue)
			linkedEndpoint, err := r.CreateEndpoint(wg)
			if err != nil {
				log.Error().Str("err", err.String()).Send()
				r.Complete(true)
				return
			}
			defer r.Complete(false)

			// if config.SocketSettings != nil {
			// 	if err := applySocketOptions(s, linkedEndpoint, config.SocketSettings); err != nil {
			// 		errors.New("failed to apply socket options: ", err).WriteToLog(session.ExportIDToError(ctx))
			// 	}
			// }

			conn := &tcpConn{
				TCPConn: gonet.NewTCPConn(wg, linkedEndpoint),
				id:      r.ID(),
			}

			go in.Handle(conn)
		})

		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

		return nil
	}
}

func (h *TunGvisorInbound) Handle(conn tun_net.TCPConn) {
	defer conn.Close()
	id := conn.ID()
	dest := net.TCPDestination(tun_net.AddressFromTCPIPAddr(id.LocalAddress), net.Port(id.LocalPort))
	src := net.TCPDestination(tun_net.AddressFromTCPIPAddr(id.RemoteAddress), net.Port(id.RemotePort))

	ctx, cancel := inbound.GetCtx(src, dest, h.Tag())
	ctx = inbound.ContextWithRawConn(ctx, conn)
	err := h.option.Handler.HandleFlow(ctx, dest, buf.NewRWD(buf.NewReader(conn), buf.NewWriter(conn), conn))
	if err != nil {
		inboundcommon.HandleError(ctx, err)
	}
	cancel(err)
}

// func applySocketOptions(s *stack.Stack, endpoint tcpip.Endpoint, config *internet.SocketConfig) tcpip.Error {
// 	if config.TcpKeepAliveInterval > 0 {
// 		interval := tcpip.KeepaliveIntervalOption(time.Duration(config.TcpKeepAliveInterval) * time.Second)
// 		if err := endpoint.SetSockOpt(&interval); err != nil {
// 			return err
// 		}
// 	}

// 	if config.TcpKeepAliveIdle > 0 {
// 		idle := tcpip.KeepaliveIdleOption(time.Duration(config.TcpKeepAliveIdle) * time.Second)
// 		if err := endpoint.SetSockOpt(&idle); err != nil {
// 			return err
// 		}
// 	}

// 	if config.TcpKeepAliveInterval > 0 || config.TcpKeepAliveIdle > 0 {
// 		endpoint.SocketOptions().SetKeepAlive(true)
// 	}
// 	{
// 		var sendBufferSizeRangeOption tcpip.TCPSendBufferSizeRangeOption
// 		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &sendBufferSizeRangeOption); err == nil {
// 			endpoint.SocketOptions().SetReceiveBufferSize(int64(sendBufferSizeRangeOption.Default), false)
// 		}

// 		var receiveBufferSizeRangeOption tcpip.TCPReceiveBufferSizeRangeOption
// 		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &receiveBufferSizeRangeOption); err == nil {
// 			endpoint.SocketOptions().SetSendBufferSize(int64(receiveBufferSizeRangeOption.Default), false)
// 		}
// 	}

// 	return nil
// }
