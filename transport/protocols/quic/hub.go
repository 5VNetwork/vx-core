package quic

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/security"
	mytls "github.com/5vnetwork/vx-core/transport/security/tls"

	"github.com/quic-go/quic-go"
	"github.com/rs/zerolog/log"
)

// Listener is an internet.Listener that listens for TCP connections.
type Listener struct {
	rawConn  *sysConn
	listener *quic.Listener
	done     *done.Instance
	h        func(net.Conn)
}

func (l *Listener) acceptStreams(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Err(err).Msg("failed to accept QUIC stream")
			select {
			case <-conn.Context().Done():
				return
			case <-l.done.Wait():
				if err := conn.CloseWithError(0, ""); err != nil {
					log.Err(err).Msg("failed to close connection")
				}
				return
			default:
				time.Sleep(time.Second)
				continue
			}
		}

		conn := &interConn{
			stream: stream,
			local:  conn.LocalAddr(),
			remote: conn.RemoteAddr(),
		}

		l.h(conn)
	}
}

func (l *Listener) keepAccepting() {
	for {
		conn, err := l.listener.Accept(context.Background())
		if err != nil {
			log.Err(err).Msg("failed to accept QUIC connections")
			if l.done.Done() {
				break
			}
			time.Sleep(time.Second)
			continue
		}
		go l.acceptStreams(conn)
	}
}

// Addr implements internet.Listener.Addr.
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close implements internet.Listener.Close.
func (l *Listener) Close() error {
	l.done.Close()
	l.listener.Close()
	l.rawConn.Close()
	return nil
}

// Listen creates a new Listener based on configurations.
func Listen(ctx context.Context, address net.Address, port net.Port, c *QuicConfig, e security.Engine,
	sockConfig *dlhelper.SocketSetting, h func(net.Conn)) (*Listener, error) {
	if address.Family().IsDomain() {
		return nil, errors.New("domain address is not allows for listening quic")
	}

	var tlsConfig *tls.Config
	var err error
	if e == nil {
		mt := &mytls.TlsConfig{
			Certificates: []*mytls.Certificate{mytls.ParseCertificate(cert.MustGenerate(nil, cert.DNSNames(internalDomain), cert.CommonName(internalDomain)))},
		}
		tlsConfig, err = mt.GetTLSConfig()
		if err != nil {
			return nil, err
		}
	} else {
		tlsConfig = e.GetTLSConfig()
	}

	rawConn, err := dlhelper.ListenSystemPacket(ctx, "udp", fmt.Sprintf("%s:%d", address, port), sockConfig)
	if err != nil {
		return nil, err
	}

	quicConfig := &quic.Config{
		HandshakeIdleTimeout:  time.Second * 8,
		MaxIdleTimeout:        time.Second * 45,
		MaxIncomingStreams:    32,
		MaxIncomingUniStreams: -1,
		KeepAlivePeriod:       time.Second * 15,
	}

	conn, err := wrapSysConn(rawConn.(*net.UDPConn), c)
	if err != nil {
		conn.Close()
		return nil, err
	}

	tr := quic.Transport{
		Conn:               conn,
		ConnectionIDLength: 12,
	}

	qListener, err := tr.Listen(tlsConfig, quicConfig)
	if err != nil {
		conn.Close()
		return nil, err
	}

	listener := &Listener{
		done:     done.New(),
		rawConn:  conn,
		listener: qListener,
		h:        h,
	}

	go listener.keepAccepting()

	return listener, nil
}
