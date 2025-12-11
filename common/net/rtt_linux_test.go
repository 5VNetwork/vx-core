//go:build linux && !android

package net_test

import (
	"net"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	. "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
)

func TestGetTCPConnectionRTT(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: func(msg []byte) []byte {
			return msg
		},
	}
	dest, err := tcpServer.Start()
	common.Must(err)
	defer tcpServer.Close()

	c, err := net.Dial("tcp", dest.Address.String()+":"+dest.Port.String())
	if err != nil {
		t.Error(err)
		return
	}

	b := make([]byte, 1024)
	for i := 0; i < 10000; i++ {
		c.Write(b)
	}
	rtt, err := GetTCPConnectionRTT(c.(*net.TCPConn))
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("RTT: %v", rtt)
}

// func TestGetBBRInfo(t *testing.T) {
// 	tcpServer := tcp.Server{
// 		MsgProcessor: func(msg []byte) []byte {
// 			return msg
// 		},
// 	}
// 	dest, err := tcpServer.Start()
// 	common.Must(err)
// 	defer tcpServer.Close()

// 	c, err := net.Dial("tcp", dest.Address.String()+":"+dest.Port.String())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	b := make([]byte, 1024)
// 	go func() {
// 		for i := 0; i < 100; i++ {
// 			c.Write(b)
// 			time.Sleep(time.Millisecond * 100)
// 		}
// 	}()
// 	bbr, err := GetBBRInfo(c.(*net.TCPConn))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Logf("bbr %v", bbr)
// }

// func TestGetBBRInfo1(t *testing.T) {
// 	tcpServer := tcp.Server{
// 		MsgProcessor: func(msg []byte) []byte {
// 			return msg
// 		},
// 	}
// 	dest, err := tcpServer.Start()
// 	common.Must(err)
// 	defer tcpServer.Close()

// 	c, err := net.Dial("tcp", dest.Address.String()+":"+dest.Port.String())
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	b := make([]byte, 1024)
// 	go func() {
// 		for i := 0; i < 100; i++ {
// 			c.Write(b)
// 			time.Sleep(time.Millisecond * 100)
// 			if i%10 == 0 {
// 				rtt, err := GetTCPConnectionRTT(c.(*net.TCPConn))
// 				if err != nil {
// 					t.Error(err)
// 					return
// 				}
// 				t.Logf("RTT: %v", rtt)
// 			}
// 		}
// 	}()
// 	bbr, err := GetBBRInfo1(c.(*net.TCPConn))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Logf("bbr %v", bbr)
// }
