package websocket_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/transport"

	"github.com/5vnetwork/vx-core/transport/dlhelper"
	. "github.com/5vnetwork/vx-core/transport/protocols/websocket"
	"github.com/5vnetwork/vx-core/transport/security"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

func Test_listenWSAndDial(t *testing.T) {
	listen, err := Listen(context.Background(),
		net.TCPDestination(net.DomainAddress("localhost"), 13146),
		&WebsocketConfig{
			Path: "ws",
		}, &dlhelper.SocketSetting{}, func(conn net.Conn) {
			go func(c net.Conn) {
				defer c.Close()

				var b [1024]byte
				_, err := c.Read(b[:])
				if err != nil {
					return
				}

				common.Must2(c.Write([]byte("Response")))
			}(conn)
		})
	common.Must(err)

	ctx := context.Background()
	conn, err := Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), 13146), &WebsocketConfig{Path: "ws"}, nil, &dlhelper.SocketSetting{})

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	n, err := conn.Read(b[:])
	common.Must(err)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}

	common.Must(conn.Close())
	<-time.After(time.Second * 5)
	conn, err = Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), 13146), &WebsocketConfig{Path: "ws"}, nil, &dlhelper.SocketSetting{})
	common.Must(err)
	_, err = conn.Write([]byte("Test connection 2"))
	common.Must(err)
	n, err = conn.Read(b[:])
	common.Must(err)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}
	common.Must(conn.Close())

	common.Must(listen.Close())
}

func TestDialWithRemoteAddr(t *testing.T) {
	listen, err := Listen(context.Background(),
		net.TCPDestination(net.DomainAddress("localhost"), 13148), &WebsocketConfig{
			Path: "ws",
		}, &dlhelper.SocketSetting{}, func(conn net.Conn) {
			go func(c net.Conn) {
				defer c.Close()

				var b [1024]byte
				_, err := c.Read(b[:])
				// common.Must(err)
				if err != nil {
					return
				}

				_, err = c.Write([]byte("Response"))
				common.Must(err)
			}(conn)
		})
	common.Must(err)

	conn, err := Dial(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), 13148),
		&WebsocketConfig{Path: "ws", Header: []*Header{{Key: "X-Forwarded-For", Value: "1.1.1.1"}}}, nil, &dlhelper.SocketSetting{})

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	n, err := conn.Read(b[:])
	common.Must(err)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}

	common.Must(listen.Close())
}

func Test_listenWSAndDial_TLS(t *testing.T) {
	if runtime.GOARCH == "arm64" {
		return
	}

	start := time.Now()

	listen, err := Listen(context.Background(),
		net.TCPDestination(net.DomainAddress("localhost"), 13143), &WebsocketConfig{
			Path: "wss",
		}, &transport.ListenerImpl{
			SecurityConfig: &tls.TlsConfig{
				AllowInsecure: true,
				Certificates:  []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil, cert.CommonName("localhost")))},
			},
		}, func(c net.Conn) {
			go func() {
				_ = c.Close()
			}()
		})
	common.Must(err)
	defer listen.Close()

	conn, err := Dial(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), 13143), &WebsocketConfig{Path: "wss"},
		common.Must2(tls.NewEngine(tls.EngineConfig{Config: &tls.TlsConfig{
			AllowInsecure: true,
			Certificates:  []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil, cert.CommonName("localhost")))},
		}, DnsServer: nil})).(security.Engine), &dlhelper.SocketSetting{})
	common.Must(err)
	_ = conn.Close()

	end := time.Now()
	if !end.Before(start.Add(time.Second * 5)) {
		t.Error("end: ", end, " start: ", start)
	}
}
