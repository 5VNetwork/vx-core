package udp

import (
	"encoding/hex"
	"net/netip"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/gtcpip"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/checksum"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func TestUdpPacketToIpPacket_IPv4(t *testing.T) {
	packetRaw := []byte{0x45, 0x0, 0x0, 0x3b, 0xe4, 0xde, 0x0, 0x0, 0x40, 0x11, 0x14, 0x1d, 0xc0, 0xa8, 0x0, 0x65, 0xc0, 0xa8, 0x0, 0x1, 0x6f, 0xfd, 0x0, 0x35, 0x0, 0x27, 0x6b, 0xcf, 0xd3, 0xf4, 0x1, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x77, 0x77, 0x77, 0x5, 0x63, 0x61, 0x6e, 0x76, 0x61, 0x3, 0x63, 0x6f, 0x6d, 0x0, 0x0, 0x1, 0x0, 0x1}
	ipv4 := header.IPv4(packetRaw)
	udpPacket := header.UDP(packetRaw[header.IPv4MinimumSize : header.IPv4MinimumSize+header.UDPMinimumSize])
	udpPayload := packetRaw[header.IPv4MinimumSize+header.UDPMinimumSize:]

	assert.True(t, ipv4.IsChecksumValid())
	assert.True(t, udpPacket.IsChecksumValid(ipv4.SourceAddress(), ipv4.DestinationAddress(), checksum.Checksum(udpPayload, 0)))

	// Create source and destination addresses
	srcIP := ipv4.SourceAddress()
	srcPort := uint16(udpPacket.SourcePort())
	dstIP := ipv4.DestinationAddress()
	dstPort := uint16(udpPacket.DestinationPort())

	// Create a UDP packet with some payload
	buffer := buf.New()
	_, err := buffer.Write(udpPayload)
	require.NoError(t, err)

	packet := &Packet{
		Source: net.Destination{
			Address: net.IPAddress(srcIP.AsSlice()),
			Port:    net.Port(srcPort),
		},
		Target: net.Destination{
			Address: net.IPAddress(dstIP.AsSlice()),
			Port:    net.Port(dstPort),
		},
		Payload: buffer,
	}

	// Convert UDP packet to IP packet
	ipBuffer := UdpPacketToIpPacket(packet)
	require.NotNil(t, ipBuffer)

	// Validate the resulting IP packet
	bytes := ipBuffer.Bytes()
	require.GreaterOrEqual(t, len(bytes), header.IPv4MinimumSize+header.UDPMinimumSize+len(udpPayload))

	// Extract and verify IPv4 header
	ipv4Header := header.IPv4(bytes[:header.IPv4MinimumSize])
	assert.Equal(t, uint8(header.UDPProtocolNumber), ipv4Header.Protocol())
	assert.Equal(t, tcpip.AddrFromSlice(srcIP.AsSlice()), ipv4Header.SourceAddress())
	assert.Equal(t, tcpip.AddrFromSlice(dstIP.AsSlice()), ipv4Header.DestinationAddress())
	assert.Equal(t, uint16(ipBuffer.Len()), ipv4Header.TotalLength())
	assert.Equal(t, uint8(64), ipv4Header.TTL())
	assert.True(t, ipv4Header.IsChecksumValid())

	// Extract and verify UDP header
	udpHeader := header.UDP(bytes[header.IPv4MinimumSize : header.IPv4MinimumSize+header.UDPMinimumSize])
	assert.Equal(t, srcPort, udpHeader.SourcePort())
	assert.Equal(t, dstPort, udpHeader.DestinationPort())
	assert.Equal(t, uint16(header.UDPMinimumSize+len(udpPayload)), udpHeader.Length())

	// Extract payload
	actualPayload := bytes[header.IPv4MinimumSize+header.UDPMinimumSize:]
	assert.Equal(t, udpPayload, actualPayload)

	// Verify udp checksum
	assert.True(t, udpHeader.IsChecksumValid(ipv4Header.SourceAddress(), ipv4Header.DestinationAddress(), checksum.Checksum(actualPayload, 0)))
	assert.Equal(t, udpPacket.Checksum(), udpHeader.Checksum())
}

// [fc20::1]:58326 ipPacket= src=74.125.250.129:19302

func TestIPPacket(t *testing.T) {
	hexString := "60000000002811404a7dfa812d727072782d766973696f6efc2000000000000000000000000000014b66e3d60028c31a0101000c2112a4427932696f4e6b7a6a5a563441002000080001f5b717ed4e93"
	bytes, err := hex.DecodeString(hexString)
	common.Must(err)
	ipPacket := gtcpip.NewIPPacket(bytes)
	payload := ipPacket.Payload()
	udpPacket := gtcpip.UDP{
		UDP: header.UDP(payload),
	}
	valid := udpPacket.IsChecksumValid(ipPacket.SourceAddress(), ipPacket.DestinationAddress(), checksum.Checksum(udpPacket.Payload(), 0))
	assert.True(t, valid)
}

func TestUdpPacketToIpPacket_IPv6(t *testing.T) {
	// Create source and destination addresses
	srcIP := netip.MustParseAddr("2001:db8::1")
	srcPort := uint16(12345)
	dstIP := netip.MustParseAddr("2001:4860:4860::8888")
	dstPort := uint16(53)

	// Create a UDP packet with some payload
	payload := []byte("DNS query payload for IPv6")
	buffer := buf.NewWithSize(int32(len(payload) + header.IPv6MinimumSize + header.UDPMinimumSize))
	_, err := buffer.Write(payload)
	require.NoError(t, err)

	packet := &Packet{
		Source: net.Destination{
			Address: net.IPAddress(srcIP.AsSlice()),
			Port:    net.Port(srcPort),
		},
		Target: net.Destination{
			Address: net.IPAddress(dstIP.AsSlice()),
			Port:    net.Port(dstPort),
		},
		Payload: buffer,
	}

	// Convert UDP packet to IP packet
	ipBuffer := UdpPacketToIpPacket(packet)
	require.NotNil(t, ipBuffer)

	// Validate the resulting IP packet
	bytes := ipBuffer.Bytes()
	require.GreaterOrEqual(t, len(bytes), header.IPv6MinimumSize+header.UDPMinimumSize+len(payload))

	// Extract and verify IPv6 header
	ipv6Header := header.IPv6(bytes[:header.IPv6MinimumSize])
	assert.Equal(t, header.UDPProtocolNumber, ipv6Header.TransportProtocol())
	assert.Equal(t, tcpip.AddrFromSlice(srcIP.AsSlice()), ipv6Header.SourceAddress())
	assert.Equal(t, tcpip.AddrFromSlice(dstIP.AsSlice()), ipv6Header.DestinationAddress())
	assert.Equal(t, uint16(ipBuffer.Len()-header.IPv6MinimumSize), ipv6Header.PayloadLength())
	assert.Equal(t, uint8(64), ipv6Header.HopLimit())

	// Extract and verify UDP header
	udpHeader := header.UDP(bytes[header.IPv6MinimumSize : header.IPv6MinimumSize+header.UDPMinimumSize])
	assert.Equal(t, srcPort, udpHeader.SourcePort())
	assert.Equal(t, dstPort, udpHeader.DestinationPort())
	assert.Equal(t, uint16(header.UDPMinimumSize+len(payload)), udpHeader.Length())

	// Extract payload
	actualPayload := bytes[header.IPv6MinimumSize+header.UDPMinimumSize:]
	assert.Equal(t, payload, actualPayload)

	// Verify UDP checksum
	udpChecksum := checksum.Checksum(actualPayload, 0)
	assert.True(t, header.UDP(udpHeader).IsChecksumValid(
		tcpip.AddrFromSlice(srcIP.AsSlice()),
		tcpip.AddrFromSlice(dstIP.AsSlice()),
		udpChecksum,
	))
}
