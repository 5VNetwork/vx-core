package udp

import (
	"crypto/rand"
	"fmt"
	"log"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"

	"github.com/google/go-cmp/cmp"
)

func RunPeerAUdp(t *testing.T, dst net.IP, port uint16) {
	c, err := net.ListenUDP("udp", &net.UDPAddr{})
	common.Must(err)
	defer c.Close()
	buf := make([]byte, 2048)
	received := make([]byte, 2048)
	for i := 0; i < 100; i++ {
		rand.Read(buf[:200])
		_, err = c.WriteToUDP(buf[:200], &net.UDPAddr{IP: dst, Port: int(port)})
		if err != nil {
			t.Fatal("PeerA cannot write", err)
		}
		n, _, err := c.ReadFromUDP(received)
		if err != nil {
			t.Fatal("PeerA cannot read", err)
		}
		if cmp.Diff(buf[:200], received[:n]) != "" {
			t.Fatalf("PeerA expected %v but got %v", buf[:200], received[:n])
		}
	}
}

func RunPeerAUdps(t *testing.T, dsts ...*net.UDPAddr) {
	c, err := net.ListenUDP("udp", &net.UDPAddr{})
	common.Must(err)
	defer c.Close()

	m := make(map[string][]byte)

	for _, dst := range dsts {
		b := make([]byte, 200)
		rand.Read(b[:200])
		m[dst.String()] = b
	}

	for _, dst := range dsts {
		go func(dst *net.UDPAddr) error {
			for i := 0; i < 10; i++ {
				_, err = c.WriteToUDP(m[dst.String()], dst)
				if err != nil {
					return err
				}
			}
			return nil
		}(dst)
	}

	received := make([]byte, 2048)
	n, src, err := c.ReadFromUDP(received)
	if err != nil {
		t.Fatal(err)
	}
	expectedBytes, found := m[src.String()]
	if !found {
		t.Fatalf("PeerA received from %v, but not expected", src)
	}
	if cmp.Diff(expectedBytes, received[:n]) != "" {
		t.Fatalf("PeerA expected %v but got %v", expectedBytes, received[:n])
	}
}

type Server struct {
	Port         net.Port
	MsgProcessor func(msg []byte) []byte
	accepting    bool
	conn         *net.UDPConn
	i            int
}

func (server *Server) Start() (net.Destination, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: int(server.Port),
		Zone: "",
	})
	if err != nil {
		return net.Destination{}, err
	}
	server.Port = net.Port(conn.LocalAddr().(*net.UDPAddr).Port)

	server.conn = conn
	go server.handleConnection(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return net.UDPDestination(net.IPAddress(localAddr.IP), net.Port(localAddr.Port)), nil
}

func (server *Server) handleConnection(conn *net.UDPConn) {
	server.accepting = true
	for server.accepting {
		buffer := make([]byte, 2*1024)
		nBytes, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Failed to read from UDP: %v\n", err)
			return
		}
		response := server.MsgProcessor(buffer[:nBytes])
		if _, err := conn.WriteToUDP(response, addr); err != nil {
			fmt.Println("Failed to write to UDP: ", err.Error())
		}
		server.i++
	}
}

func (server *Server) Close() error {
	server.accepting = false
	log.Println("received", server.i, "packets")
	return server.conn.Close()
}
