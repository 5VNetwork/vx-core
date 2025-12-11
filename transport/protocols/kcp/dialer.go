package kcp

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/dice"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"
)

var globalConv = uint32(dice.RollUint16())

func fetchInput(_ context.Context, input io.Reader, reader PacketReader, conn *Connection) {
	cache := make(chan *buf.Buffer, 1024)
	go func() {
		for {
			payload := buf.New()
			if _, err := payload.ReadOnce(input); err != nil {
				payload.Release()
				close(cache)
				return
			}
			select {
			case cache <- payload:
			default:
				payload.Release()
			}
		}
	}()

	for payload := range cache {
		segments := reader.Read(payload.Bytes())
		payload.Release()
		if len(segments) > 0 {
			conn.Input(segments)
		}
	}
}

type kcpDialer struct {
	config       *KcpConfig
	engine       security.Engine
	socketConfig i.Dialer
}

func NewKcpDialer(config *KcpConfig, engine security.Engine, socketConfig i.Dialer) *kcpDialer {
	return &kcpDialer{
		config:       config,
		engine:       engine,
		socketConfig: socketConfig,
	}
}

func (d *kcpDialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	return Dial(ctx, dest, d.config, d.engine, d.socketConfig)
}

// Dial dials a new KCP connections to the specific destination.
func Dial(ctx context.Context, dest net.Destination, kcpConfig *KcpConfig, securityConfig security.Engine, socketConfig i.Dialer) (net.Conn, error) {
	dest.Network = net.Network_UDP
	// errors.New("dialing mKCP to ", dest).WriteToLog()

	rawConn, err := socketConfig.Dial(ctx, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to dial system connection: %w", err)
	}

	header, err := kcpConfig.GetPackerHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to create packet header: %w", err)
	}
	s, err := kcpConfig.GetSecurity()
	if err != nil {
		return nil, fmt.Errorf("failed to get security: %w", err)
	}
	reader := &KCPPacketReader{
		Header:   header,
		Security: s,
	}
	writer := &KCPPacketWriter{
		Header:   header,
		Security: s,
		Writer:   rawConn,
	}

	conv := uint16(atomic.AddUint32(&globalConv, 1))
	session := NewConnection(ConnMetadata{
		LocalAddr:    rawConn.LocalAddr(),
		RemoteAddr:   rawConn.RemoteAddr(),
		Conversation: conv,
	}, writer, rawConn, kcpConfig)

	go fetchInput(ctx, rawConn, reader, session)

	var iConn net.Conn = session

	if securityConfig != nil {
		iConn, err = securityConfig.GetClientConn(iConn, security.OptionWithDestination{Dest: dest})
		if err != nil {
			return nil, err
		}
	}

	return iConn, nil
}
