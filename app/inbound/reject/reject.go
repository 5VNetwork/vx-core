package reject

import (
	"context"

	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

type TCPReject struct {
	FakeDnsPool fakeDnsPool
	NatIp6      net.IP
	Router      i.Router
	InboundTag  string
	UserLogger  *userlogger.UserLogger
}

type fakeDnsPool interface {
	IsIPInIPPool(ip net.Address) bool
}

func (r *TCPReject) Reject(p []byte) *buf.Buffer {
	if header.IPVersion(p) == header.IPv6Version {
		ipv6 := header.IPv6(p)
		if ipv6.TransportProtocol() == header.TCPProtocolNumber {
			tcp := header.TCP(p[header.IPv6MinimumSize:])
			if tcp.Flags().Contains(header.TCPFlagSyn) {
				dst := ipv6.DestinationAddress().As16()
				if r.FakeDnsPool != nil && r.FakeDnsPool.IsIPInIPPool(net.IPAddress(dst[:])) {
					return nil
				}
				if r.NatIp6 != nil && r.NatIp6.Equal(net.IP(dst[:])) {
					return nil
				}
				src := ipv6.SourceAddress().As16()
				srcDestination := net.Destination{
					Address: net.IPAddress(src[:]),
					Port:    net.Port(tcp.SourcePort()),
					Network: net.Network_TCP,
				}
				target := net.Destination{
					Address: net.IPAddress(dst[:]),
					Port:    net.Port(tcp.DestinationPort()),
					Network: net.Network_TCP,
				}
				info := &session.Info{
					Source:     srcDestination,
					Target:     target,
					InboundTag: r.InboundTag,
				}
				handler, _ := r.Router.PickHandler(context.Background(), info)
				if handler == nil {
					// log.Debug().Str("dst", target.String()).Msg("reject tcp")
					// r.UserLogger.LogReject(target)
					// return GenerateRstForTcpSynIPv60(ipv6, tcp)
					return nil
				}
				if outHandler, ok := handler.(i.HandlerWith6Info); ok && !outHandler.Support6() {
					log.Debug().Str("handler", handler.Tag()).Str("dst", target.String()).Msg("reject tcp because handler not support ipv6")
					r.UserLogger.LogReject(info, "handler not support ipv6")
					return GenerateRstForTcpSynIPv60(ipv6, tcp)
				}
			}

		}
	}
	return nil
}

func GenerateRstForTcpSynIPv6(ipv6Header header.IPv6, tcpHeader header.TCP) *buf.Buffer {
	// Calculate total packet size
	// IPv6 header (40 bytes) + TCP header with no options
	totalLen := header.IPv6MinimumSize + header.TCPMinimumSize
	b := buf.New()

	// Create the buffer for the new packet
	packet := b.Extend(int32(totalLen))
	// Create IPv6 header view
	ipv6 := header.IPv6(packet[:header.IPv6MinimumSize])

	// Set IPv6 header fields
	ipv6.SetPayloadLength(uint16(header.TCPMinimumSize))
	ipv6.SetNextHeader(uint8(header.TCPProtocolNumber))
	ipv6.SetHopLimit(64) // Standard hop limit

	// Swap source and destination addresses for the response
	srcPort := tcpHeader.DestinationPort()
	dstPort := tcpHeader.SourcePort()

	// Set source address to original destination
	srcIP := ipv6Header.DestinationAddress()
	ipv6.SetSourceAddress(srcIP)

	// Set destination address to original source
	dstIP := ipv6Header.SourceAddress()
	ipv6.SetDestinationAddress(dstIP)

	// Create TCP header view
	tcp := header.TCP(packet[header.IPv6MinimumSize:])

	// Set TCP header fields
	tcp.SetSourcePort(srcPort)
	tcp.SetDestinationPort(dstPort)

	// For RST in response to SYN, use the received ISN+1 as the sequence number
	// RFC 793 specifies this sequence number handling
	tcp.SetAckNumber(tcpHeader.SequenceNumber() + 1)
	tcp.SetSequenceNumber(0)

	tcp.SetDataOffset(header.TCPMinimumSize)

	// Set flags (RST flag only)
	tcp.SetFlags(uint8(header.TCPFlagRst))

	tcp.SetChecksum(^tcp.CalculateChecksum(header.PseudoHeaderChecksum(header.TCPProtocolNumber,
		ipv6.SourceAddress(), ipv6.DestinationAddress(), header.TCPMinimumSize)))

	return b
}

func GenerateRstForTcpSynIPv60(ipv6Header header.IPv6, tcpHeader header.TCP) *buf.Buffer {
	// Calculate total packet size
	// IPv6 header (40 bytes) + TCP header with no options
	totalLen := header.IPv6MinimumSize + header.TCPMinimumSize
	b := buf.New()

	// Create the buffer for the new packet
	packet := b.Extend(int32(totalLen))
	// Create IPv6 header view
	ipv6 := header.IPv6(packet[:header.IPv6MinimumSize])

	ipv6.Encode(&header.IPv6Fields{
		PayloadLength:     uint16(header.TCPMinimumSize),
		TransportProtocol: header.TCPProtocolNumber,
		SrcAddr:           ipv6Header.DestinationAddress(),
		DstAddr:           ipv6Header.SourceAddress(),
	})

	// Create TCP header view
	tcp := header.TCP(ipv6.Payload())

	fields := header.TCPFields{
		SrcPort:    tcpHeader.DestinationPort(),
		DstPort:    tcpHeader.SourcePort(),
		DataOffset: header.TCPMinimumSize,
		Flags:      header.TCPFlagRst | header.TCPFlagAck,
		AckNum:     tcpHeader.SequenceNumber() + 1,
	}

	tcp.Encode(&fields)

	tcp.SetChecksum(^tcp.CalculateChecksum(header.PseudoHeaderChecksum(header.TCPProtocolNumber,
		ipv6.SourceAddress(), ipv6.DestinationAddress(), header.TCPMinimumSize)))

	return b
}

type UdpReject struct {
	FakeDnsPool fakeDnsPool
	Router      i.Router
	InboundTag  string
	UserLogger  *userlogger.UserLogger
}

// p is a ip packet with udp payload
func (r *UdpReject) Reject(p []byte) *buf.Buffer {
	if header.IPVersion(p) == header.IPv6Version {
		ipv6 := header.IPv6(p)
		udp := header.UDP(p[header.IPv6MinimumSize:])
		dst := ipv6.DestinationAddress().As16()
		if r.FakeDnsPool != nil && r.FakeDnsPool.IsIPInIPPool(net.IPAddress(dst[:])) {
			return nil
		}
		src := ipv6.SourceAddress().As16()
		srcDestination := net.Destination{
			Address: net.IPAddress(src[:]),
			Port:    net.Port(udp.SourcePort()),
			Network: net.Network_UDP,
		}
		target := net.Destination{
			Address: net.IPAddress(dst[:]),
			Port:    net.Port(udp.DestinationPort()),
			Network: net.Network_UDP,
		}

		ctx := log.Logger.WithContext(context.Background())
		info := &session.Info{
			Source:     srcDestination,
			Target:     target,
			InboundTag: r.InboundTag,
		}
		handler, _ := r.Router.PickHandler(ctx, info)
		if handler == nil {
			return nil
		}
		if outHandler, ok := handler.(i.HandlerWith6Info); ok && !outHandler.Support6() {
			log.Debug().Str("handler", handler.Tag()).Str("dst", target.String()).Msg("reject udp because handler not support ipv6")
			r.UserLogger.LogReject(info, "handler not support ipv6")
			return CreateICMPv6Unreachable(ipv6)
		}

	}
	return nil
}

// CreateICMPv6Unreachable takes an IPv6 packet with a UDP payload and returns an IPv6 packet
// with an ICMPv6 Destination Unreachable (Type 1, Code 4) message.
// ipv6Hdr contains entire ipv6 packet
func CreateICMPv6Unreachable(ipv6Hdr header.IPv6) *buf.Buffer {
	icmpv6PayloadLen := header.IPv6MinimumMTU - header.IPv6MinimumSize - header.ICMPv6DstUnreachableMinimumSize
	if icmpv6PayloadLen > len(ipv6Hdr) {
		icmpv6PayloadLen = len(ipv6Hdr)
	}

	ipv6TotalLen := header.IPv6MinimumSize + header.ICMPv6DstUnreachableMinimumSize + icmpv6PayloadLen
	b := buf.New()
	packet := b.Extend(int32(ipv6TotalLen))

	ipv6 := header.IPv6(packet[:header.IPv6MinimumSize])
	ipv6.Encode(&header.IPv6Fields{
		PayloadLength:     uint16(ipv6TotalLen - header.IPv6MinimumSize),
		TransportProtocol: header.ICMPv6ProtocolNumber,
		SrcAddr:           ipv6Hdr.DestinationAddress(),
		DstAddr:           ipv6Hdr.SourceAddress(),
	})

	b.Zero(header.IPv6MinimumSize, header.IPv6MinimumSize+header.ICMPv6DstUnreachableMinimumSize)

	icmpv6 := header.ICMPv6(packet[header.IPv6MinimumSize:])
	icmpv6.SetType(header.ICMPv6DstUnreachable)
	icmpv6.SetCode(header.ICMPv6PortUnreachable)
	icmpv6.SetChecksum(header.ICMPv6Checksum(header.ICMPv6ChecksumParams{
		Header:      icmpv6[:header.ICMPv6DstUnreachableMinimumSize],
		Src:         ipv6Hdr.DestinationAddress(),
		Dst:         ipv6Hdr.SourceAddress(),
		PayloadCsum: checksum.Checksum(ipv6Hdr[:icmpv6PayloadLen], 0),
		PayloadLen:  icmpv6PayloadLen,
	}))
	copy(icmpv6[header.ICMPv6DstUnreachableMinimumSize:], ipv6Hdr[:icmpv6PayloadLen])

	return b
}

// func (r *TCPReject) Reject(src, dst net.Destination) bool {
// 	if !dst.Address.Family().IsIPv6() {
// 		return false
// 	}

// 	if r.FakeDnsPool.IsIPInIPPool(dst.Address) {
// 		return false
// 	}

// 	handler := r.Router.PickHandler(context.Background(), &session.Info{
// 		Source: src,
// 		Target: dst,
// 	})
// 	if handler == nil {
// 		log.Debug().Str("dst", dst.String()).Msg("reject")
// 		return true
// 	}
// 	if outHandler, ok := handler.(*outbound.Handler); ok && !outHandler.Support6() {
// 		log.Debug().Str("dst", dst.String()).Msg("reject")
// 		return true
// 	}

// 	return false
// }
