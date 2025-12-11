package quic_test

// import (
// 	"context"
// 	"crypto/rand"
// 	"testing"
// 	"time"

// 	"github.com/5vnetwork/vx-core/common"
// 	"github.com/5vnetwork/vx-core/common/buf"
// 	"github.com/5vnetwork/vx-core/common/net"
// 	protocol "github.com/5vnetwork/vx-core/common/protocol"
// 	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
// 	"github.com/5vnetwork/vx-core/common/serial"
// 	"github.com/5vnetwork/vx-core/test/servers/udp"
// 	"github.com/5vnetwork/vx-core/transport/dlhelper"
// 	"github.com/5vnetwork/vx-core/transport/headers/wireguard"
// 	"github.com/5vnetwork/vx-core/transport/protocols/quic"
// 	"github.com/5vnetwork/vx-core/transport/security/tls"
// 	"github.com/google/go-cmp/cmp"
// )

// func TestQuicConnection(t *testing.T) {
// 	port := udp.PickPort()

// 	engine, err := tls.NewEngine(&tls.TlsConfig{
// 		Certificates: []*tls.Certificate{
// 			tls.ParseCertificate(
// 				cert.MustGenerate(nil,
// 					cert.DNSNames("www.v2fly.org"),
// 				),
// 			),
// 		},
// 	})
// 	common.Must(err)
// 	listener, err := quic.Listen(context.Background(), net.LocalHostIP, port, &quic.QuicConfig{}, engine, &dlhelper.SocketSetting{}, func(conn net.Conn) {
// 		go func() {
// 			defer conn.Close()

// 			b := buf.New()
// 			defer b.Release()

// 			for {
// 				b.Clear()
// 				if _, err := b.ReadOnce(conn); err != nil {
// 					return
// 				}
// 				common.Must2(conn.Write(b.Bytes()))
// 			}
// 		}()
// 	})
// 	common.Must(err)

// 	defer listener.Close()

// 	time.Sleep(time.Second)

// 	dctx := context.Background()
// 	engine1, err1 := tls.NewEngine(&tls.TlsConfig{
// 		ServerName:    "www.v2fly.org",
// 		AllowInsecure: true,
// 	})
// 	common.Must(err1)
// 	conn, err := quic.Dial(dctx, net.TCPDestination(net.LocalHostIP, port), &quic.QuicConfig{}, engine1, &dlhelper.SocketSetting{})
// 	common.Must(err)
// 	defer conn.Close()

// 	const N = 1024
// 	b1 := make([]byte, N)
// 	common.Must2(rand.Read(b1))
// 	b2 := buf.New()

// 	common.Must2(conn.Write(b1))

// 	b2.Clear()
// 	common.Must2(b2.ReadFullFrom(conn, N))
// 	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
// 		t.Error(r)
// 	}

// 	common.Must2(conn.Write(b1))

// 	b2.Clear()
// 	common.Must2(b2.ReadFullFrom(conn, N))
// 	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
// 		t.Error(r)
// 	}
// }

// func TestQuicConnectionWithoutTLS(t *testing.T) {
// 	port := udp.PickPort()

// 	listener, err := quic.Listen(context.Background(), net.LocalHostIP, port, &quic.QuicConfig{}, nil, nil, func(conn net.Conn) {
// 		go func() {
// 			defer conn.Close()

// 			b := buf.New()
// 			defer b.Release()

// 			for {
// 				b.Clear()
// 				if _, err := b.ReadOnce(conn); err != nil {
// 					return
// 				}
// 				common.Must2(conn.Write(b.Bytes()))
// 			}
// 		}()
// 	})
// 	common.Must(err)

// 	defer listener.Close()

// 	time.Sleep(time.Second)

// 	dctx := context.Background()
// 	conn, err := quic.Dial(dctx, net.TCPDestination(net.LocalHostIP, port), &quic.QuicConfig{}, nil, &dlhelper.SocketSetting{})
// 	common.Must(err)
// 	defer conn.Close()

// 	const N = 1024
// 	b1 := make([]byte, N)
// 	common.Must2(rand.Read(b1))
// 	b2 := buf.New()

// 	common.Must2(conn.Write(b1))

// 	b2.Clear()
// 	common.Must2(b2.ReadFullFrom(conn, N))
// 	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
// 		t.Error(r)
// 	}

// 	common.Must2(conn.Write(b1))

// 	b2.Clear()
// 	common.Must2(b2.ReadFullFrom(conn, N))
// 	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
// 		t.Error(r)
// 	}
// }

// func TestQuicConnectionAuthHeader(t *testing.T) {
// 	port := udp.PickPort()

// 	listener, err := quic.Listen(context.Background(), net.LocalHostIP, port, &quic.QuicConfig{
// 		Header:   serial.ToTypedMessage(&wireguard.WireguardConfig{}),
// 		Key:      "abcd",
// 		Security: protocol.SecurityType_AES128_GCM,
// 	},
// 		nil, nil, func(conn net.Conn) {
// 			go func() {
// 				defer conn.Close()

// 				b := buf.New()
// 				defer b.Release()

// 				for {
// 					b.Clear()
// 					if _, err := b.ReadOnce(conn); err != nil {
// 						return
// 					}
// 					common.Must2(conn.Write(b.Bytes()))
// 				}
// 			}()
// 		})
// 	common.Must(err)

// 	defer listener.Close()

// 	time.Sleep(time.Second)

// 	dctx := context.Background()
// 	conn, err := quic.Dial(dctx, net.TCPDestination(net.LocalHostIP, port), &quic.QuicConfig{
// 		Header:   serial.ToTypedMessage(&wireguard.WireguardConfig{}),
// 		Key:      "abcd",
// 		Security: protocol.SecurityType_AES128_GCM,
// 	},
// 		nil, &dlhelper.SocketSetting{})
// 	common.Must(err)
// 	defer conn.Close()

// 	const N = 1024
// 	b1 := make([]byte, N)
// 	common.Must2(rand.Read(b1))
// 	b2 := buf.New()

// 	common.Must2(conn.Write(b1))

// 	b2.Clear()
// 	common.Must2(b2.ReadFullFrom(conn, N))
// 	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
// 		t.Error(r)
// 	}

// 	common.Must2(conn.Write(b1))

// 	b2.Clear()
// 	common.Must2(b2.ReadFullFrom(conn, N))
// 	if r := cmp.Diff(b2.Bytes(), b1); r != "" {
// 		t.Error(r)
// 	}
// }
