package reject

import (
	"bytes"
	"net"
	"testing"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
)

func TestGenerateRstForTcpSynIPv6(t *testing.T) {
	// Create a mock IPv6 packet with TCP SYN
	mockIPv6Packet := createMockIPv6TcpSynPacket()
	ipv6Header := header.IPv6(mockIPv6Packet[:header.IPv6MinimumSize])
	tcpHeader := header.TCP(mockIPv6Packet[header.IPv6MinimumSize:])

	// Generate RST packet
	rstPacket := GenerateRstForTcpSynIPv6(ipv6Header, tcpHeader)
	if rstPacket == nil {
		t.Fatal("GenerateRstForTcpSynIPv6 returned nil")
	}

	// Verify the generated packet
	verifyRstPacket(t, rstPacket.Bytes(), ipv6Header, tcpHeader)
}

func TestGenerateRstForTcpSynIPv6_DifferentPorts(t *testing.T) {
	// Test with different source and destination ports
	mockIPv6Packet := createMockIPv6TcpSynPacket()
	ipv6Header := header.IPv6(mockIPv6Packet[:header.IPv6MinimumSize])
	tcpHeader := header.TCP(mockIPv6Packet[header.IPv6MinimumSize:])

	// Modify TCP header ports
	tcpHeader.SetSourcePort(12345)
	tcpHeader.SetDestinationPort(54321)

	// Generate RST packet
	rstPacket := GenerateRstForTcpSynIPv6(ipv6Header, tcpHeader)

	// Verify the generated packet
	verifyRstPacket(t, rstPacket.Bytes(), ipv6Header, tcpHeader)
}

func TestGenerateRstForTcpSynIPv6_DifferentAddresses(t *testing.T) {
	// Test with different IPv6 addresses
	mockIPv6Packet := createMockIPv6TcpSynPacket()
	ipv6Header := header.IPv6(mockIPv6Packet[:header.IPv6MinimumSize])
	tcpHeader := header.TCP(mockIPv6Packet[header.IPv6MinimumSize:])

	// Set custom IPv6 addresses
	srcIP := net.ParseIP("2001:db8::1").To16()
	dstIP := net.ParseIP("2001:db8::2").To16()

	srcAddr := tcpip.AddrFromSlice(srcIP)
	dstAddr := tcpip.AddrFromSlice(dstIP)

	ipv6Header.SetSourceAddress(srcAddr)
	ipv6Header.SetDestinationAddress(dstAddr)

	// Generate RST packet
	rstPacket := GenerateRstForTcpSynIPv6(ipv6Header, tcpHeader)

	// Verify the generated packet
	verifyRstPacket(t, rstPacket.Bytes(), ipv6Header, tcpHeader)
}

func TestGenerateRstForTcpSynIPv6_SequenceNumbers(t *testing.T) {
	// Test with specific sequence number
	mockIPv6Packet := createMockIPv6TcpSynPacket()
	ipv6Header := header.IPv6(mockIPv6Packet[:header.IPv6MinimumSize])
	tcpHeader := header.TCP(mockIPv6Packet[header.IPv6MinimumSize:])

	// Set specific sequence number
	tcpHeader.SetSequenceNumber(12345678)

	// Generate RST packet
	rstPacket := GenerateRstForTcpSynIPv6(ipv6Header, tcpHeader)

	// Verify the RST has correct acknowledgment number (sequence+1)
	rstTCP := header.TCP(rstPacket.Bytes()[header.IPv6MinimumSize:])

	if rstTCP.AckNumber() != tcpHeader.SequenceNumber()+1 {
		t.Errorf("RST packet has incorrect ACK number: got %d, want %d",
			rstTCP.AckNumber(), tcpHeader.SequenceNumber()+1)
	}
}

// Helper function to create a mock IPv6 packet with TCP SYN
func createMockIPv6TcpSynPacket() []byte {
	// Total size: IPv6 header + TCP header
	totalSize := header.IPv6MinimumSize + header.TCPMinimumSize
	packet := make([]byte, totalSize)

	// Initialize IPv6 header
	ipv6 := header.IPv6(packet[:header.IPv6MinimumSize])
	ipv6.SetPayloadLength(uint16(header.TCPMinimumSize))
	ipv6.SetNextHeader(uint8(header.TCPProtocolNumber))
	ipv6.SetHopLimit(64)

	// Set source and destination addresses
	srcIP := net.ParseIP("2001:db8::1").To16()
	dstIP := net.ParseIP("2001:db8::2").To16()

	srcAddr := tcpip.AddrFromSlice(srcIP)
	dstAddr := tcpip.AddrFromSlice(dstIP)

	ipv6.SetSourceAddress(srcAddr)
	ipv6.SetDestinationAddress(dstAddr)

	// Initialize TCP header with SYN flag
	tcp := header.TCP(packet[header.IPv6MinimumSize:])
	tcp.SetSourcePort(80)
	tcp.SetDestinationPort(45678)
	tcp.SetSequenceNumber(1000)
	tcp.SetDataOffset(header.TCPMinimumSize)
	tcp.SetFlags(uint8(header.TCPFlagSyn))

	return packet
}

// Helper function to verify the RST packet
func verifyRstPacket(t *testing.T, rstPacket []byte, originalIPv6 header.IPv6, originalTCP header.TCP) {
	// Check packet length
	expectedLength := header.IPv6MinimumSize + header.TCPMinimumSize
	if len(rstPacket) != expectedLength {
		t.Errorf("RST packet has incorrect length: got %d, want %d",
			len(rstPacket), expectedLength)
	}

	// Parse the RST packet
	rstIPv6 := header.IPv6(rstPacket[:header.IPv6MinimumSize])
	rstTCP := header.TCP(rstPacket[header.IPv6MinimumSize:])

	// Verify IPv6 header fields
	if rstIPv6.PayloadLength() != uint16(header.TCPMinimumSize) {
		t.Errorf("IPv6 payload length incorrect: got %d, want %d",
			rstIPv6.PayloadLength(), header.TCPMinimumSize)
	}

	if rstIPv6.NextHeader() != uint8(header.TCPProtocolNumber) {
		t.Errorf("IPv6 next header incorrect: got %d, want %d",
			rstIPv6.NextHeader(), header.TCPProtocolNumber)
	}

	if rstIPv6.HopLimit() != 64 {
		t.Errorf("IPv6 hop limit incorrect: got %d, want %d",
			rstIPv6.HopLimit(), 64)
	}

	// Verify source/destination addresses (should be swapped)
	srcAddrEqual := rstIPv6.SourceAddress() == originalIPv6.DestinationAddress()
	if !srcAddrEqual {
		t.Errorf("IPv6 source address incorrect: got %v, want %v",
			rstIPv6.SourceAddress(), originalIPv6.DestinationAddress())
	}

	dstAddrEqual := rstIPv6.DestinationAddress() == originalIPv6.SourceAddress()
	if !dstAddrEqual {
		t.Errorf("IPv6 destination address incorrect: got %v, want %v",
			rstIPv6.DestinationAddress(), originalIPv6.SourceAddress())
	}

	// Verify TCP header fields
	if rstTCP.SourcePort() != originalTCP.DestinationPort() {
		t.Errorf("TCP source port incorrect: got %d, want %d",
			rstTCP.SourcePort(), originalTCP.DestinationPort())
	}

	if rstTCP.DestinationPort() != originalTCP.SourcePort() {
		t.Errorf("TCP destination port incorrect: got %d, want %d",
			rstTCP.DestinationPort(), originalTCP.SourcePort())
	}

	// Verify sequence/ack numbers
	if rstTCP.AckNumber() != originalTCP.SequenceNumber()+1 {
		t.Errorf("TCP ack number incorrect: got %d, want %d",
			rstTCP.AckNumber(), originalTCP.SequenceNumber()+1)
	}

	// Verify RST flag is set and no other flags
	if rstTCP.Flags() != header.TCPFlagRst {
		t.Errorf("TCP flags incorrect: got %d, want %d",
			rstTCP.Flags(), header.TCPFlagRst)
	}

	// Verify TCP checksum is set (non-zero)
	if rstTCP.Checksum() == 0 {
		t.Error("TCP checksum is not set")
	}
}

func TestCreateICMPv4Unreachable(t *testing.T) {
	mockIPv4Packet := createMockIPv4UDPPacket(16)
	ipv4Header := header.IPv4(mockIPv4Packet)

	icmpPacket := CreateICMPv4Unreachable(ipv4Header)
	if icmpPacket == nil {
		t.Fatal("CreateICMPv4Unreachable returned nil")
	}

	verifyICMPv4UnreachablePacket(t, icmpPacket.Bytes(), ipv4Header)
}

func TestCreateICMPv4Unreachable_ShortPacketCopiesWholeOriginal(t *testing.T) {
	mockIPv4Packet := createMockIPv4UDPPacket(0)
	ipv4Header := header.IPv4(mockIPv4Packet)

	icmpPacket := CreateICMPv4Unreachable(ipv4Header)
	if icmpPacket == nil {
		t.Fatal("CreateICMPv4Unreachable returned nil")
	}

	packet := icmpPacket.Bytes()
	icmpv4 := header.ICMPv4(packet[header.IPv4MinimumSize:])
	copied := icmpv4[header.ICMPv4MinimumSize:]

	if !bytes.Equal(copied, mockIPv4Packet) {
		t.Errorf("ICMP payload copied incorrect original packet: got %v, want %v", copied, mockIPv4Packet)
	}
}

func createMockIPv4UDPPacket(payloadLen int) []byte {
	totalSize := header.IPv4MinimumSize + header.UDPMinimumSize + payloadLen
	packet := make([]byte, totalSize)

	ipv4 := header.IPv4(packet[:header.IPv4MinimumSize])
	srcAddr := tcpip.AddrFromSlice(net.ParseIP("192.0.2.1").To4())
	dstAddr := tcpip.AddrFromSlice(net.ParseIP("198.51.100.2").To4())
	ipv4.Encode(&header.IPv4Fields{
		TotalLength: uint16(totalSize),
		TTL:         64,
		Protocol:    uint8(header.UDPProtocolNumber),
		SrcAddr:     srcAddr,
		DstAddr:     dstAddr,
	})
	ipv4.SetChecksum(^ipv4.CalculateChecksum())

	udp := header.UDP(packet[header.IPv4MinimumSize:])
	udp.Encode(&header.UDPFields{
		SrcPort: 12345,
		DstPort: 443,
		Length:  uint16(header.UDPMinimumSize + payloadLen),
	})

	for i := 0; i < payloadLen; i++ {
		packet[header.IPv4MinimumSize+header.UDPMinimumSize+i] = byte(i + 1)
	}

	return packet
}

func verifyICMPv4UnreachablePacket(t *testing.T, icmpPacket []byte, originalIPv4 header.IPv4) {
	t.Helper()

	expectedCopiedLen := int(originalIPv4.HeaderLength()) + header.ICMPv4MinimumErrorPayloadSize
	if expectedCopiedLen > len(originalIPv4) {
		expectedCopiedLen = len(originalIPv4)
	}
	expectedLength := header.IPv4MinimumSize + header.ICMPv4MinimumSize + expectedCopiedLen
	if len(icmpPacket) != expectedLength {
		t.Errorf("ICMPv4 packet has incorrect length: got %d, want %d", len(icmpPacket), expectedLength)
	}

	ipv4 := header.IPv4(icmpPacket[:header.IPv4MinimumSize])
	icmpv4 := header.ICMPv4(icmpPacket[header.IPv4MinimumSize:])

	if ipv4.Protocol() != uint8(header.ICMPv4ProtocolNumber) {
		t.Errorf("IPv4 protocol incorrect: got %d, want %d", ipv4.Protocol(), header.ICMPv4ProtocolNumber)
	}

	if ipv4.TTL() != 64 {
		t.Errorf("IPv4 TTL incorrect: got %d, want %d", ipv4.TTL(), 64)
	}

	if ipv4.SourceAddress() != originalIPv4.DestinationAddress() {
		t.Errorf("IPv4 source address incorrect: got %v, want %v", ipv4.SourceAddress(), originalIPv4.DestinationAddress())
	}

	if ipv4.DestinationAddress() != originalIPv4.SourceAddress() {
		t.Errorf("IPv4 destination address incorrect: got %v, want %v", ipv4.DestinationAddress(), originalIPv4.SourceAddress())
	}

	if icmpv4.Type() != header.ICMPv4DstUnreachable {
		t.Errorf("ICMPv4 type incorrect: got %d, want %d", icmpv4.Type(), header.ICMPv4DstUnreachable)
	}

	if icmpv4.Code() != header.ICMPv4PortUnreachable {
		t.Errorf("ICMPv4 code incorrect: got %d, want %d", icmpv4.Code(), header.ICMPv4PortUnreachable)
	}

	if icmpv4.Checksum() == 0 {
		t.Error("ICMPv4 checksum is not set")
	}

	expectedCopied := originalIPv4[:expectedCopiedLen]
	gotCopied := icmpv4[header.ICMPv4MinimumSize:]
	if !bytes.Equal(gotCopied, expectedCopied) {
		t.Errorf("ICMPv4 payload incorrect: got %v, want %v", gotCopied, expectedCopied)
	}
}
