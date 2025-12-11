//go:build !windows && !android && server
// +build !windows,!android,server

package domainsocket_test

import (
	"context"
	"log"
	"runtime"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	. "github.com/5vnetwork/vx-core/transport/protocols/domainsocket"
)

func TestListen(t *testing.T) {
	ctx := context.Background()
	streamSettings := &DomainSocketConfig{
		Path: "/tmp/ts3",
	}
	listener, err := Listen(ctx, nil, net.Port(0), streamSettings, nil)
	common.Must(err)
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("failed to accept connection: ", err)
				return
			}
			func(conn net.Conn) {
				defer conn.Close()

				b := buf.New()
				defer b.Release()
				common.Must2(b.ReadOnce(conn))
				b.WriteString("Response")

				common.Must2(conn.Write(b.Bytes()))
			}(conn)
		}
	}()

	conn, err := Dial(ctx, net.Destination{}, streamSettings, nil)
	common.Must(err)
	defer conn.Close()

	common.Must2(conn.Write([]byte("Request")))

	b := buf.New()
	defer b.Release()
	common.Must2(b.ReadOnce(conn))

	if b.String() != "RequestResponse" {
		t.Error("expected response as 'RequestResponse' but got ", b.String())
	}
}

func TestListenAbstract(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	ctx := context.Background()
	streamSettings := &DomainSocketConfig{
		Path:     "/tmp/ts3",
		Abstract: true,
	}
	listener, err := Listen(ctx, nil, net.Port(0), streamSettings, nil)
	common.Must(err)
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("failed to accept connection: ", err)
				return
			}
			func(conn net.Conn) {
				defer conn.Close()

				b := buf.New()
				defer b.Release()
				common.Must2(b.ReadOnce(conn))
				b.WriteString("Response")

				common.Must2(conn.Write(b.Bytes()))
			}(conn)
		}
	}()

	conn, err := Dial(ctx, net.Destination{}, streamSettings, nil)
	common.Must(err)
	defer conn.Close()

	common.Must2(conn.Write([]byte("Request")))

	b := buf.New()
	defer b.Release()
	common.Must2(b.ReadOnce(conn))

	if b.String() != "RequestResponse" {
		t.Error("expected response as 'RequestResponse' but got ", b.String())
	}
}
