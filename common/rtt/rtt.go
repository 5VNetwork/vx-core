package rtt

import (
	"context"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type Protocol string

const (
	TCP  Protocol = "tcp"
	UDP  Protocol = "udp"
	ICMP Protocol = "icmp"
)

func TestRTT(host string, port string, protocol Protocol, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch protocol {
	case TCP:
		return testTCPRTT(ctx, host, port)
	case UDP:
		return testUDPRTT(ctx, host, port)
	case ICMP:
		return testICMPRTT(ctx, host, timeout)
	default:
		return 0, net.InvalidAddrError("unsupported protocol")
	}
}

// PingRTT is a convenience function for ICMP ping that doesn't require a port parameter
func PingRTT(host string, timeout time.Duration) (time.Duration, error) {
	return TestRTT(host, "", ICMP, timeout)
}

func testTCPRTT(ctx context.Context, host, port string) (time.Duration, error) {
	start := time.Now()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return time.Since(start), nil
}

func testUDPRTT(ctx context.Context, host, port string) (time.Duration, error) {
	start := time.Now()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", net.JoinHostPort(host, port))
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Send a small probe packet
	_, err = conn.Write([]byte{0})
	if err != nil {
		return 0, err
	}

	// Try to read response (with short timeout for UDP)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	buffer := make([]byte, 1)
	conn.Read(buffer)

	return time.Since(start), nil
}

func testICMPRTT(ctx context.Context, host string, timeout time.Duration) (time.Duration, error) {
	// Resolve the host to an IP address
	dst, err := net.ResolveIPAddr("ip4:icmp", host)
	if err != nil {
		return 0, err
	}

	// Try IPv6 if IPv4 resolution failed
	if dst == nil {
		dst, err = net.ResolveIPAddr("ip6:ipv6-icmp", host)
		if err != nil {
			return 0, err
		}
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		// Try IPv6 if IPv4 fails
		conn, err = icmp.ListenPacket("ip6:ipv6-icmp", "::")
		if err != nil {
			return 0, err
		}
	}
	defer conn.Close()

	// Create ICMP message
	var icmpType icmp.Type
	var icmpCode int
	if dst.IP.To4() != nil {
		icmpType = ipv4.ICMPTypeEcho
		icmpCode = 0
	} else {
		icmpType = ipv6.ICMPTypeEchoRequest
		icmpCode = 0
	}

	// Create echo request message
	message := &icmp.Message{
		Type: icmpType,
		Code: icmpCode,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("ping"),
		},
	}

	messageBytes, err := message.Marshal(nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()

	// Set write deadline
	conn.SetWriteDeadline(start.Add(timeout))

	// Send ICMP packet
	_, err = conn.WriteTo(messageBytes, dst)
	if err != nil {
		return 0, err
	}

	// Set read deadline
	conn.SetReadDeadline(start.Add(timeout))

	// Read response
	reply := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}

	_, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return 0, err
	}

	rtt := time.Since(start)

	// Verify the response is from the expected peer
	if peer.String() != dst.String() {
		return 0, net.InvalidAddrError("response from unexpected peer")
	}

	// For basic RTT measurement, we don't need to parse the full ICMP response
	// The fact that we received a response from the correct peer within timeout is sufficient
	return rtt, nil
}
