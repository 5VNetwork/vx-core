package scenarios

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/errors"
	net1 "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/common/units"
	"github.com/5vnetwork/vx-core/proxy/socks"
	"github.com/5vnetwork/vx-core/test/nameserver"
	"github.com/5vnetwork/vx-core/test/servers/tcp"
	"github.com/5vnetwork/vx-core/test/servers/udp"
	"github.com/miekg/dns"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var Timeout = time.Second * 2

func StartDnsServer() (*dns.Server, uint16) {
	dnsPort := net1.PickUDPPort()
	dnsServer := dns.Server{
		Addr:    "127.0.0.1:" + common.Uint16ToString(dnsPort),
		Net:     "udp",
		Handler: &nameserver.StaticHandler{},
		UDPSize: 1200,
	}
	go dnsServer.ListenAndServe()
	return &dnsServer, dnsPort
}

func StartTcpServer() (tcp.Server, net1.Destination) {
	tcpServer := tcp.Server{
		MsgProcessor: Xor,
	}
	tcpDest, err := tcpServer.Start()
	common.Must(err)
	return tcpServer, tcpDest
}

func StartUdpServer() (*udp.Server, net1.Destination) {
	udpServer := udp.Server{
		MsgProcessor: Xor,
	}
	udpDest, err := udpServer.Start()
	common.Must(err)
	return &udpServer, udpDest
}

func Xor(b []byte) []byte {
	r := make([]byte, len(b))
	for i, v := range b {
		r[i] = v ^ 'c'
	}
	return r
}

func readFrom(conn net.Conn, timeout time.Duration, length int) []byte {
	b := make([]byte, length)
	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)
	n, err := io.ReadFull(conn, b[:length])
	if err != nil {
		fmt.Println("Unexpected error from readFrom:", err)
	}
	return b[:n]
}

func readFrom2(conn net.Conn, timeout time.Duration, length int) ([]byte, error) {
	b := make([]byte, length)
	deadline := time.Now().Add(timeout)
	conn.SetReadDeadline(deadline)
	n, err := io.ReadFull(conn, b[:length])
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

var (
	testBinaryPath    string
	testBinaryPathGen sync.Once
)

func genTestBinaryPath() {
	testBinaryPathGen.Do(func() {
		var tempDir string
		common.Must(retry.Timed(5, 100).On(func() error {
			dir, err := os.MkdirTemp("", "vray")
			if err != nil {
				return err
			}
			tempDir = dir
			return nil
		}))
		file := filepath.Join(tempDir, "vray.test")
		if runtime.GOOS == "windows" {
			file += ".exe"
		}
		testBinaryPath = file
		fmt.Printf("Generated binary path: %s\n", file)
	})
}

func GetSourcePath() string {
	return filepath.Join("../../main.go")
	// return filepath.Join("github.com/5vnetwork/vx-core/github.com/5vnetwork/vx-core", "main")
}

func CloseAllServers(servers []*exec.Cmd) {
	log.Print("Closing all servers.")
	for _, server := range servers {
		if runtime.GOOS == "windows" {
			server.Process.Kill()
		} else {
			server.Process.Signal(syscall.SIGTERM)
		}
	}
	for _, server := range servers {
		server.Process.Wait()
	}
	log.Print("All servers closed.")

}

func CloseServer(server *exec.Cmd) {
	log.Print("Closing server.")

	if runtime.GOOS == "windows" {
		server.Process.Kill()
	} else {
		server.Process.Signal(syscall.SIGTERM)
	}
	server.Process.Wait()
	log.Print("server closed.")

}
func TestTCPConn(port net1.Port, payloadSize int, timeout time.Duration) func() error {
	return func() error {
		conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(port),
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		return WriteToConn(conn, payloadSize, timeout)()
	}
}

func TestTCPConnTls(port net1.Port, payloadSize int, timeout time.Duration) func() error {
	return func() error {
		c, err := net.DialTCP("tcp", nil, &net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(port),
		})
		if err != nil {
			return err
		}
		// log.Print("TCP connects to port ", conn.LocalAddr().(*net.TCPAddr).Port)

		conn := tls.Client(c, &tls.Config{
			InsecureSkipVerify: true,
		})
		defer func() {
			time.Sleep(timeout)
			err := conn.Close()
			if err != nil {
				log.Print("Error closing connection", err)
			}
		}()

		return TestTCPConnTls2(conn, payloadSize, timeout)()
	}
}

func TestTCPConnTls2(conn net.Conn, payloadSize int, timeout time.Duration) func() error {
	return func() (err1 error) {
		start := time.Now()
		defer func() {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			// For info on each, see: https://golang.org/pkg/runtime/#MemStats
			fmt.Println("testConn finishes:", time.Since(start).Milliseconds(), "ms\t",
				err1, "\tAlloc =", units.ByteSize(m.Alloc).String(),
				"\tTotalAlloc =", units.ByteSize(m.TotalAlloc).String(),
				"\tSys =", units.ByteSize(m.Sys).String(),
				"\tNumGC =", m.NumGC)
		}()
		payload := make([]byte, 1200)
		common.Must2(rand.Read(payload))

		totalNumWritten := 0
		for totalNumWritten < payloadSize {
			nBytes, err := conn.Write(payload)
			if err != nil {
				return err
			}
			if nBytes != len(payload) {
				return errors.New("expect ", len(payload), " written, but actually ", nBytes)
			}
			totalNumWritten += nBytes

			response, err := readFrom2(conn, timeout, nBytes)
			if err != nil {
				return err
			}
			_ = response
			if r := bytes.Compare(response, Xor(payload)); r != 0 {
				return errors.New(r)
			}
		}

		return nil
	}
}

// Write [payloadSize] bytes to [conn] and read [payloadSize] bytes from [conn]
// and using xor to check the correctness of the data
func WriteToConn(conn net.Conn, payloadSize int, timeout time.Duration) func() error {
	return func() (err1 error) {
		start := time.Now()
		defer func() {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			// For info on each, see: https://golang.org/pkg/runtime/#MemStats
			fmt.Println("testConn finishes:", time.Since(start).Milliseconds(), "ms\t",
				err1, "\tAlloc =", units.ByteSize(m.Alloc).String(),
				"\tTotalAlloc =", units.ByteSize(m.TotalAlloc).String(),
				"\tSys =", units.ByteSize(m.Sys).String(),
				"\tNumGC =", m.NumGC)
		}()
		payload := make([]byte, payloadSize)
		common.Must2(rand.Read(payload))

		nBytes, err := conn.Write(payload)
		if err != nil {
			return err
		}
		if nBytes != len(payload) {
			return errors.New("expect ", len(payload), " written, but actually ", nBytes)
		}

		response, err := readFrom2(conn, timeout, payloadSize)
		if err != nil {
			return err
		}
		_ = response

		if r := bytes.Compare(response, Xor(payload)); r != 0 {
			return errors.New(r)
		}

		return nil
	}
}

func TestUDPConn(port net1.Port, payloadSize int, timeout time.Duration) func() error { // nolint: unparam
	return func() error {
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(port),
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		return WriteToConn(conn, payloadSize, timeout)()
	}
}

func TestUDPConnN(port net1.Port, payloadSize int, timeout time.Duration, num int) func() error { // nolint: unparam
	return func() error {
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: int(port),
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		for i := 0; i < num; i++ {
			payload := make([]byte, payloadSize)
			common.Must2(rand.Read(payload))
			nBytes, err := conn.Write(payload)
			if err != nil {
				return err
			}
			if nBytes != len(payload) {
				return errors.New("expect ", len(payload), " written, but actually ", nBytes)
			}
			response, err := readFrom2(conn, timeout, payloadSize)
			if err != nil {
				return err
			}
			_ = response
			if r := bytes.Compare(response, Xor(payload)); r != 0 {
				return errors.New(r)
			}
		}
		return nil
	}
}

func testUDPConnConeClient(payloadSize int, ports ...net1.Port) func() error { // nolint: unparam
	return func() error {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{
			Port: int(udp.PickPort()),
		})
		if err != nil {
			return err
		}
		defer conn.Close()
		log.Print("UDP listens on port ", conn.LocalAddr().(*net.UDPAddr).Port)

		m := make(map[int][]byte)

		for _, port := range ports {
			buf := make([]byte, payloadSize)
			common.Must2(rand.Read(buf))

			p := port
			_, err := conn.WriteToUDP(buf, &net.UDPAddr{IP: net.IP{127, 0, 0, 1}, Port: int(p)})
			if err != nil {
				log.Print("cannot write", err)
				return err
			}
			m[int(p)] = buf
		}

		buf1 := make([]byte, payloadSize)
		for {
			_, from, err := conn.ReadFromUDP(buf1)
			if err != nil {
				log.Print("cannot read", err)
				return err
			}
			if m[from.Port] != nil {
				if r := bytes.Compare(m[from.Port], buf1); r != 0 {
					log.Print("expected", m[from.Port], "but got", buf1)
					return errors.New(r)
				}
				delete(m, from.Port)
			}
			if len(m) == 0 {
				break
			}
		}

		return nil
	}
}

func testUDPConnConeServer(ports ...net1.Port) error {
	l := sync.Mutex{}
	var srcs []*net.UDPAddr
	var errg errgroup.Group

	for _, port := range ports {
		c, err := net.ListenUDP("udp", &net.UDPAddr{
			Port: int(port),
		})
		common.Must(err)
		errg.Go(func() error {
			_, src, err := c.ReadFromUDP(make([]byte, 2048))
			if err != nil {
				log.Print("cannot read", err)
				return err
			}
			l.Lock()
			srcs = append(srcs, src)
			l.Unlock()
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return err
	}
	if len(srcs) != len(ports) {
		return errors.New("expected", len(ports), "but got", len(srcs))
	}
	port := srcs[0].Port
	for i := 1; i < len(srcs); i++ {
		if srcs[i].Port != port {
			return errors.New("expected", port, "but got", srcs[i].Port)
		}
	}
	return nil
}

func socksUdpWriter(dstPort uint16) socks.UDPWriter {
	c, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net1.LocalHostIP.IP(),
		Port: int(dstPort),
	})
	common.Must(err)
	return *socks.NewUDPWriter(&protocol.RequestHeader{}, c)
}

func StartUdpServers() map[*udp.Server]net1.Destination {
	m := make(map[*udp.Server]net1.Destination)
	for i := 0; i < 10; i++ {
		server, dst := StartUdpServer()
		m[server] = dst
	}
	return m
}
