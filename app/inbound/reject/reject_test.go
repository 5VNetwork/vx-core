package reject

import (
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
