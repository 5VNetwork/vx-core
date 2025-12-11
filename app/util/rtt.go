package util

import (
	context "context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/apernet/quic-go"
	"github.com/rs/zerolog/log"
)

func RttTest(ctx context.Context, ap net.AddressPort,
	dl i.Dialer, pl i.PacketListener, ipResolver i.IPResolver) (int, error) {
	if ap.Address.Family().IsDomain() {
		ips, err := ipResolver.LookupIP(ctx, ap.Address.Domain())
		if err != nil {
			return 0, err
		}
		if len(ips) == 0 {
			return 0, errors.New("no ip found")
		}
		ap.Address = net.IPAddress(ips[0])
	}
	start := time.Now()

	conn, err := dl.Dial(ctx, net.TCPDestination(ap.Address, ap.Port))
	if err == nil {
		defer conn.Close()
		return int(time.Since(start).Milliseconds()), nil
	}
	log.Ctx(ctx).Debug().Err(err).Msg("tcp dial")

	start = time.Now()
	packetConn, err := pl.ListenPacket(ctx, "udp", "")
	if err != nil {
		return 0, fmt.Errorf("failed to listen packet: %v", err)
	}
	defer packetConn.Close()
	addr := &net.UDPAddr{
		IP:   net.ParseIP(ap.Address.IP().String()),
		Port: int(ap.Port),
	}
	c := make(chan struct{})
	packetConnWrapper := &packetConnWrapper{
		PacketConn: packetConn,
		addr:       addr,
		onReceivedResponse: func() {
			close(c)
		},
	}
	packetConnWrapper.SetReadDeadline(time.Now().Add(5 * time.Second))
	go quic.Dial(ctx, packetConnWrapper, addr, &tls.Config{}, &quic.Config{})
	select {
	case <-c:
		return int(time.Since(start).Milliseconds()), nil
	case <-time.After(5 * time.Second):
		return 0, errors.New("no response received")
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

type packetConnWrapper struct {
	net.PacketConn
	addr               net.Addr
	onReceivedResponse func()
	once               sync.Once
}

func (p *packetConnWrapper) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, addr, err = p.PacketConn.ReadFrom(b)
	if err != nil {
		return
	}
	if addr.String() == p.addr.String() {
		p.once.Do(p.onReceivedResponse)
	}
	return
}

// generateFakeQUICInitialPacket creates a fake QUIC initial packet for RTT testing
func generateFakeQUICInitialPacket() []byte {
	// QUIC Initial packet structure (RFC 9000):
	// Header Form (1 bit) + Fixed Bit (1 bit) + Long Packet Type (2 bits) + Reserved Bits (2 bits) + Packet Number Length (2 bits)
	// Version (32 bits)
	// Destination Connection ID Length (8 bits) + Destination Connection ID (0-160 bits)
	// Source Connection ID Length (8 bits) + Source Connection ID (0-160 bits)
	// Token Length (Variable) + Token (Variable)
	// Length (Variable) + Packet Number (8-32 bits)
	// Payload (Variable)

	packet := make([]byte, 0, 1200) // QUIC minimum packet size

	// Header byte: 0xC0 = 11000000 (Long header, Initial packet type)
	// Bits: 1 (Long header) + 1 (Fixed bit) + 00 (Initial) + 00 (Reserved) + 00 (PN length = 1)
	packet = append(packet, 0xC0)

	// Version: QUIC version 1 (0x00000001)
	version := make([]byte, 4)
	binary.BigEndian.PutUint32(version, 0x00000001)
	packet = append(packet, version...)

	// Destination Connection ID Length (8 bytes)
	packet = append(packet, 8)

	// Destination Connection ID (8 random bytes)
	destConnID := make([]byte, 8)
	rand.Read(destConnID)
	packet = append(packet, destConnID...)

	// Source Connection ID Length (8 bytes)
	packet = append(packet, 8)

	// Source Connection ID (8 random bytes)
	srcConnID := make([]byte, 8)
	rand.Read(srcConnID)
	packet = append(packet, srcConnID...)

	// Token Length (0 for initial packet from client)
	packet = append(packet, 0)

	// Calculate payload length
	payloadLength := 100
	// Length field includes: packet number (1 byte) + payload
	lengthField := 1 + payloadLength
	packet = append(packet, byte(lengthField))

	// Packet Number (1 byte, value 0)
	packet = append(packet, 0)

	// Payload: CRYPTO frame with fake handshake data
	// CRYPTO frame format: Type (1 byte) + Offset (variable) + Length (variable) + Data
	cryptoFrame := make([]byte, payloadLength)
	cryptoFrame[0] = 0x06 // CRYPTO frame type

	// Offset = 0 (encoded as 1-byte variable-length integer)
	cryptoFrame[1] = 0

	// Length = remaining data length (encoded as 1-byte variable-length integer)
	dataLength := payloadLength - 3 // 3 bytes for type + offset + length
	cryptoFrame[2] = byte(dataLength)

	// Fill with fake TLS handshake data (ClientHello-like)
	// This makes it look more like a real QUIC initial packet
	fakeHandshake := []byte{
		0x01, 0x00, 0x00, 0x5C, // Handshake type (ClientHello) + length
		0x03, 0x03, // TLS version (TLS 1.2)
	}

	// Copy fake handshake data
	copy(cryptoFrame[3:], fakeHandshake)

	// Fill remaining with random data
	if len(fakeHandshake) < dataLength {
		rand.Read(cryptoFrame[3+len(fakeHandshake):])
	}

	packet = append(packet, cryptoFrame...)

	return packet
}
