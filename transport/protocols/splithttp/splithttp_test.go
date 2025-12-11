package splithttp_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/transport/dlhelper"
	. "github.com/5vnetwork/vx-core/transport/protocols/splithttp"
	"github.com/5vnetwork/vx-core/transport/security/tls"
)

func Test_ListenXHAndDial(t *testing.T) {
	listenPort := tcp.PickPort()
	listen, err := ListenXH(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort), &SplitHttpConfig{
		Path: "/sh",
	}, &dlhelper.SocketSetting{}, func(conn net.Conn) {
		go func(c net.Conn) {
			defer c.Close()

			var b [1024]byte
			c.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, err := c.Read(b[:])
			if err != nil {
				return
			}

			common.Must2(c.Write([]byte("Response")))
		}(conn)
	})
	common.Must(err)
	ctx := context.Background()

	dialer, err := NewXhttpDialer(&SplitHttpConfig{Path: "sh"}, nil, &dlhelper.SocketSetting{}, nil)
	common.Must(err)

	conn, err := dialer.Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), listenPort))

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	fmt.Println("test2")
	n, _ := io.ReadFull(conn, b[:])
	fmt.Println("string is", n)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}

	common.Must(conn.Close())
	conn, err = dialer.Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), listenPort))

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 2"))
	common.Must(err)
	n, _ = io.ReadFull(conn, b[:])
	common.Must(err)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}
	common.Must(conn.Close())

	common.Must(listen.Close())
}

func TestDialWithRemoteAddr(t *testing.T) {
	listenPort := tcp.PickPort()
	listen, err := ListenXH(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort), &SplitHttpConfig{
		Path: "sh",
	}, &dlhelper.SocketSetting{}, func(conn net.Conn) {
		go func(c net.Conn) {
			defer c.Close()

			var b [1024]byte
			_, err := c.Read(b[:])
			// common.Must(err)
			if err != nil {
				return
			}

			_, err = c.Write([]byte(c.RemoteAddr().String()))
			common.Must(err)
		}(conn)
	})
	common.Must(err)

	dialer, err := NewXhttpDialer(&SplitHttpConfig{Path: "sh",
		Headers: map[string]string{"X-Forwarded-For": "1.1.1.1"}},
		nil, &dlhelper.SocketSetting{}, nil)
	common.Must(err)

	conn, err := dialer.Dial(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort))

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	n, _ := io.ReadFull(conn, b[:])
	if string(b[:n]) != "1.1.1.1:0" {
		t.Error("response: ", string(b[:n]))
	}

	common.Must(listen.Close())
}

func Test_ListenXHAndDial_TLS(t *testing.T) {
	if runtime.GOARCH == "arm64" {
		return
	}

	listenPort := tcp.PickPort()
	start := time.Now()

	config := &SplitHttpConfig{
		Path: "shs",
	}
	tlsConfig := &tls.TlsConfig{
		AllowInsecure: true,
		Certificates:  []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil, cert.CommonName("localhost")))},
	}
	listenerImpl := &transport.ListenerImpl{
		Socket:         &dlhelper.SocketSetting{},
		SecurityConfig: tlsConfig,
	}
	listen, err := ListenXH(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort), config,
		listenerImpl, func(conn net.Conn) {
			go func() {
				defer conn.Close()

				var b [1024]byte
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, err := conn.Read(b[:])
				if err != nil {
					return
				}

				common.Must2(conn.Write([]byte("Response")))
			}()
		})
	common.Must(err)
	defer listen.Close()

	e, err := tls.NewEngine(tls.EngineConfig{Config: tlsConfig, DnsServer: nil})
	common.Must(err)
	dialer, err := NewXhttpDialer(config, e, &dlhelper.SocketSetting{}, nil)
	common.Must(err)

	conn, err := dialer.Dial(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort))
	common.Must(err)

	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	n, _ := io.ReadFull(conn, b[:])
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}

	end := time.Now()
	if !end.Before(start.Add(time.Second * 5)) {
		t.Error("end: ", end, " start: ", start)
	}
}

func Test_ListenXHAndDial_H2C(t *testing.T) {
	if runtime.GOARCH == "arm64" {
		return
	}

	listenPort := tcp.PickPort()

	config := &SplitHttpConfig{
		Path: "shs",
	}
	listen, err := ListenXH(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort), config,
		&dlhelper.SocketSetting{}, func(conn net.Conn) {
			go func() {
				_ = conn.Close()
			}()
		})
	common.Must(err)
	defer listen.Close()

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	client := http.Client{
		Transport: &http.Transport{
			Protocols: protocols,
		},
	}

	resp, err := client.Get("http://" + net.LocalHostIP.String() + ":" + listenPort.String())
	common.Must(err)

	if resp.StatusCode != 404 {
		t.Error("Expected 404 but got:", resp.StatusCode)
	}

	if resp.ProtoMajor != 2 {
		t.Error("Expected h2 but got:", resp.ProtoMajor)
	}
}

// func Test_ListenXHAndDial_QUIC(t *testing.T) {
// 	if runtime.GOARCH == "arm64" {
// 		return
// 	}

// 	listenPort := udp.PickPort()
// 	fmt.Println("listenPort: ", listenPort)
// 	start := time.Now()

// 	config := &SplitHttpConfig{
// 		Path: "shs",
// 	}
// 	tlsConfig := &tls.TlsConfig{
// 		AllowInsecure: true,
// 		Certificates:  []*tls.Certificate{tls.ParseCertificate(cert.MustGenerate(nil, cert.CommonName("localhost")))},
// 		NextProtocol:  []string{"h3"},
// 	}
// 	e, err := tls.NewEngine(tls.EngineConfig{Config: tlsConfig, DnsServer: nil})
// 	common.Must(err)
// 	serverClosed := false
// 	listen, err := ListenXH(context.Background(), net.UDPDestination(net.DomainAddress("localhost"), listenPort), config,
// 		&dlhelper.SocketSetting{}, func(conn net.Conn) {
// 			go func() {
// 				defer conn.Close()

// 				b := buf.New()
// 				defer b.Release()

// 				for {
// 					b.Clear()
// 					if _, err := b.ReadOnce(conn); err != nil {
// 						break
// 					}
// 					common.Must2(conn.Write(b.Bytes()))
// 				}

// 				serverClosed = true
// 			}()
// 		})
// 	common.Must(err)
// 	defer listen.Close()

// 	time.Sleep(time.Second)

// 	dialer, err := NewXhttpDialer(config, e, &dlhelper.SocketSetting{}, nil)
// 	common.Must(err)

// 	conn, err := dialer.Dial(context.Background(), net.UDPDestination(net.DomainAddress("localhost"), listenPort))
// 	common.Must(err)

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

// 	conn.Close()
// 	time.Sleep(100 * time.Millisecond)
// 	if !serverClosed {
// 		t.Error("server did not get closed")
// 	}

// 	end := time.Now()
// 	if !end.Before(start.Add(time.Second * 5)) {
// 		t.Error("end: ", end, " start: ", start)
// 	}
// }

func Test_ListenXHAndDial_Unix(t *testing.T) {
	tempDir := t.TempDir()
	tempSocket := tempDir + "/server.sock"

	listen, err := ListenXH(context.Background(), net.UnixDestination(net.DomainAddress(tempSocket)), &SplitHttpConfig{
		Path: "/sh",
	}, &dlhelper.SocketSetting{}, func(conn net.Conn) {
		go func(c net.Conn) {
			defer c.Close()

			var b [1024]byte
			c.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, err := c.Read(b[:])
			if err != nil {
				return
			}

			common.Must2(c.Write([]byte("Response")))
		}(conn)
	})
	common.Must(err)
	ctx := context.Background()
	config := &SplitHttpConfig{
		Host: "example.com",
		Path: "sh",
	}
	dialer, err := NewXhttpDialer(config, nil, &dlhelper.SocketSetting{}, nil)
	common.Must(err)
	conn, err := dialer.Dial(ctx, net.UnixDestination(net.DomainAddress(tempSocket)))

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	fmt.Println("test2")
	n, _ := io.ReadFull(conn, b[:])
	fmt.Println("string is", n)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}

	common.Must(conn.Close())
	conn, err = dialer.Dial(ctx, net.UnixDestination(net.DomainAddress(tempSocket)))

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 2"))
	common.Must(err)
	n, _ = io.ReadFull(conn, b[:])
	common.Must(err)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}
	common.Must(conn.Close())

	common.Must(listen.Close())
}

func Test_queryString(t *testing.T) {
	listenPort := tcp.PickPort()
	listen, err := ListenXH(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort), &SplitHttpConfig{
		// this querystring does not have any effect, but sometimes people blindly copy it from websocket config. make sure the outbound doesn't break
		Path: "/sh?ed=2048",
	}, &dlhelper.SocketSetting{}, func(conn net.Conn) {
		go func(c net.Conn) {
			defer c.Close()

			var b [1024]byte
			c.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, err := c.Read(b[:])
			if err != nil {
				return
			}

			common.Must2(c.Write([]byte("Response")))
		}(conn)
	})
	common.Must(err)
	ctx := context.Background()

	dialer, err := NewXhttpDialer(&SplitHttpConfig{Path: "sh?ed=2048"}, nil, &dlhelper.SocketSetting{}, nil)
	conn, err := dialer.Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), listenPort))

	common.Must(err)
	_, err = conn.Write([]byte("Test connection 1"))
	common.Must(err)

	var b [1024]byte
	fmt.Println("test2")
	n, _ := io.ReadFull(conn, b[:])
	fmt.Println("string is", n)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}

	common.Must(conn.Close())
	common.Must(listen.Close())
}

func Test_maxUpload(t *testing.T) {
	listenPort := tcp.PickPort()
	config := &SplitHttpConfig{
		Path: "/sh",
		ScMaxEachPostBytes: &RangeConfig{
			From: 10000,
			To:   10000,
		},
	}

	var uploadSize int
	listen, err := ListenXH(context.Background(), net.TCPDestination(net.DomainAddress("localhost"), listenPort), config,
		&dlhelper.SocketSetting{}, func(conn net.Conn) {
			go func(c net.Conn) {
				defer c.Close()
				var b [10240]byte
				c.SetReadDeadline(time.Now().Add(2 * time.Second))
				n, err := c.Read(b[:])
				if err != nil {
					return
				}

				uploadSize = n

				common.Must2(c.Write([]byte("Response")))
			}(conn)
		})
	common.Must(err)
	ctx := context.Background()

	dialer, err := NewXhttpDialer(&SplitHttpConfig{Path: "sh"}, nil, &dlhelper.SocketSetting{}, nil)
	conn, err := dialer.Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), listenPort))

	// send a slightly too large upload
	var upload [10001]byte
	_, err = conn.Write(upload[:])
	common.Must(err)

	var b [10240]byte
	n, _ := io.ReadFull(conn, b[:])
	fmt.Println("string is", n)
	if string(b[:n]) != "Response" {
		t.Error("response: ", string(b[:n]))
	}
	common.Must(conn.Close())

	if uploadSize > 10000 || uploadSize == 0 {
		t.Error("incorrect upload size: ", uploadSize)
	}

	common.Must(listen.Close())
}
