package shadowsocks_2022

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/singbridge"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/helper"
	shadowsocks "github.com/sagernet/sing-shadowsocks"
	"github.com/sagernet/sing-shadowsocks/shadowaead_2022"
	C "github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	"lukechampine.com/blake3"
)

type Outbound struct {
	*ClientSettings
	method shadowsocks.Method
}

type ClientSettings struct {
	Address    net.Address
	PortPicker i.PortSelector
	Key        string
	Dialer     i.Dialer
	Method     string
}

// keySizeForMethod returns the PSK length in bytes for the given Shadowsocks 2022 method.
func keySizeForMethod(method string) int {
	switch method {
	case "2022-blake3-aes-128-gcm":
		return 16
	case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
		return 32
	default:
		return 32
	}
}

// toBase64PSK converts the key to a base64-encoded PSK of the length required by the method.
// If the key is valid base64 and decodes to >= keySize bytes, it is truncated and re-encoded.
// Otherwise (plain password, UUID, etc.) it is derived with BLAKE3 to the required length.
func toBase64PSK(key string, keySize int) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("missing psk")
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err == nil && len(decoded) >= keySize {
		return base64.StdEncoding.EncodeToString(decoded[:keySize]), nil
	}
	// Not valid base64 or too short: derive from password (UUID, plain text, etc.)
	digest := blake3.Sum256([]byte(key))
	return base64.StdEncoding.EncodeToString(digest[:keySize]), nil
}

func NewClient(config *ClientSettings) (*Outbound, error) {
	o := &Outbound{
		ClientSettings: config,
	}
	if C.Contains(shadowaead_2022.List, config.Method) {
		if config.Key == "" {
			return nil, errors.New("missing psk")
		}
		keySize := keySizeForMethod(config.Method)
		key, err := toBase64PSK(config.Key, keySize)
		if err != nil {
			return nil, err
		}
		method, err := shadowaead_2022.NewWithPassword(config.Method, key, nil)
		if err != nil {
			return nil, fmt.Errorf("create method: %w", err)
		}
		o.method = method
	} else {
		return nil, fmt.Errorf("unknown method %s", config.Method)
	}
	return o, nil
}

func (o *Outbound) HandleFlow(ctx context.Context, dst net.Destination,
	rw buf.ReaderWriter) error {

	port := o.PortPicker.SelectPort()

	if dst.Network == net.Network_TCP {
		destination := net.TCPDestination(o.Address, net.Port(port))
		connection, err := o.Dialer.Dial(ctx, destination)
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		defer connection.Close()

		serverConn := o.method.DialEarlyConn(connection,
			singbridge.ToSocksaddr(dst))
		wr := buf.NewWriter(serverConn)
		if err = buf.CopyOnceTimeout(rw, wr, proxy.FirstPayloadTimeout); err != nil &&
			err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout {
			return fmt.Errorf("failed to write A request payload: %w", err)
		}

		return helper.Relay(ctx, rw, rw, buf.NewReader(serverConn), wr)
	} else {
		packetReaderWriter := &udp.BufReaderWriterToPacketReaderWriter{
			ReaderWriter: rw,
			Dest:         dst,
		}
		return o.HandlePacketConn(ctx, dst, packetReaderWriter)
	}
}

func (o *Outbound) HandlePacketConn(ctx context.Context, dst net.Destination,
	p udp.PacketReaderWriter) error {
	port := o.PortPicker.SelectPort()
	destination := net.UDPDestination(o.Address, net.Port(port))
	connection, err := o.Dialer.Dial(ctx, destination)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer connection.Close()

	serverConn := o.method.DialPacketConn(connection)
	packetConn := &singbridge.PacketConnWrapper{
		PacketReaderWriter: p,
	}
	return bufio.CopyPacketConn(ctx, packetConn, serverConn)
}
