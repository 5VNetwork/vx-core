package http_test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	"github.com/5vnetwork/vx-core/transport/protocols/http"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/google/go-cmp/cmp"
)

func TestHTTPConnection(t *testing.T) {
	port := tcp.PickPort()

	listener, err := http.Listen(context.Background(), net.TCPDestination(net.LocalHostIP, port),
		&http.HttpConfig{}, &transport.ListenerImpl{
			Socket: &dlhelper.SocketSetting{},
			SecurityConfig: &tls.TlsConfig{
				NextProtocol: []string{"h2"},
				Certificates: []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil,
					cert.CommonName("www.v2fly.org")))},
			},
		}, func(conn net.Conn) {
			go func() {
				defer conn.Close()

				b := buf.New()
				defer b.Release()

				for {
					if _, err := b.ReadOnce(conn); err != nil {
						return
					}
					_, err := conn.Write(b.Bytes())
					if err != nil {
						// Connection closed or error occurred, exit gracefully
						return
					}
				}
			}()
		})
	common.Must(err)

	defer listener.Close()

	time.Sleep(time.Second)

	dctx := context.Background()
	security1, err1 := tls.NewEngine(tls.EngineConfig{Config: &tls.TlsConfig{
		ServerName:    "www.v2fly.org",
		AllowInsecure: true,
	}, DnsServer: nil})
	common.Must(err1)
	conn, err := http.Dial(dctx, net.TCPDestination(net.LocalHostIP, port), &http.HttpConfig{}, security1, &dlhelper.SocketSetting{})
	common.Must(err)
	defer conn.Close()

	const N = 1024
	b1 := make([]byte, N)
	common.Must2(rand.Read(b1))
	b2 := buf.New()

	nBytes, err := conn.Write(b1)
	common.Must(err)
	if nBytes != N {
		t.Error("write: ", nBytes)
	}

	b2.Clear()
	common.Must2(b2.ReadFullFrom(conn, N))
	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
		t.Error(r)
	}

	nBytes, err = conn.Write(b1)
	common.Must(err)
	if nBytes != N {
		t.Error("write: ", nBytes)
	}

	b2.Clear()
	common.Must2(b2.ReadFullFrom(conn, N))
	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
		t.Error(r)
	}
}
