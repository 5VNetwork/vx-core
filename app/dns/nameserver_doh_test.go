package dns

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	xnet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/test/nameserver"
	"github.com/5vnetwork/vx-core/transport"
	d "github.com/miekg/dns"
)

var (
	localDoHURL string
)

func init() {
	// Pick a random port for the local DoH server
	port := xnet.PickTCPPort()
	localDoHURL = fmt.Sprintf("http://127.0.0.1:%d/dns-query", port)

	// Set up HTTP server that implements DoH protocol
	mux := http.NewServeMux()
	mux.HandleFunc("/dns-query", handleDoHRequest)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("DoH server failed: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	log.Printf("Local DoH server started at: %s", localDoHURL)
}

// handleDoHRequest implements a simple DoH server using our StaticHandler
func handleDoHRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/dns-message" {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Read the DNS message from the request body
	dnsBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the DNS message
	msg := new(d.Msg)
	if err := msg.Unpack(dnsBytes); err != nil {
		http.Error(w, "Failed to parse DNS message", http.StatusBadRequest)
		return
	}

	// Use our StaticHandler to handle the DNS query
	handler := &nameserver.StaticHandler{}

	// Create a mock ResponseWriter to capture the response
	mockWriter := &mockDNSResponseWriter{}
	handler.ServeDNS(mockWriter, msg)

	// Pack the response
	responseBytes, err := mockWriter.response.Pack()
	if err != nil {
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/dns-message")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

// mockDNSResponseWriter implements dns.ResponseWriter for our DoH server
type mockDNSResponseWriter struct {
	response *d.Msg
}

func (m *mockDNSResponseWriter) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53}
}

func (m *mockDNSResponseWriter) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

func (m *mockDNSResponseWriter) WriteMsg(msg *d.Msg) error {
	m.response = msg
	return nil
}

func (m *mockDNSResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockDNSResponseWriter) Close() error {
	return nil
}

func (m *mockDNSResponseWriter) TsigStatus() error {
	return nil
}

func (m *mockDNSResponseWriter) TsigTimersOnly(bool) {
}

func (m *mockDNSResponseWriter) Hijack() {
}

func TestDohNameServer(t *testing.T) {
	doh, _ := NewDoHNameServer(DoHNameServerOption{
		Name: "test",
		Url:  localDoHURL,
		Handler: freedom.New(
			transport.DefaultDialer,
			transport.DefaultPacketListener,
			"", nil,
		),
	})

	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := doh.HandleQuery(context.Background(), m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) == 0 {
		t.Fatal("expected at least one answer")
	}
	log.Print("Reply:", reply)
}

func TestDohNameServer1111(t *testing.T) {
	doh, _ := NewDoHNameServer(DoHNameServerOption{
		Name: "test",
		Url:  localDoHURL,
		Handler: freedom.New(
			transport.DefaultDialer,
			transport.DefaultPacketListener,
			"", nil,
		),
	})

	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := doh.HandleQuery(context.Background(), m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) == 0 {
		t.Fatal("expected at least one answer")
	}
	log.Print("Reply:", reply)
}

func TestDohNameServerClientIp(t *testing.T) {
	doh, _ := NewDoHNameServer(DoHNameServerOption{
		Name: "test",
		Url:  localDoHURL,
		Handler: freedom.New(
			transport.DefaultDialer,
			transport.DefaultPacketListener,
			"", nil,
		),
		ClientIP: net.ParseIP("21.54.95.47"),
	})

	m := new(d.Msg)
	m.SetQuestion("www.apple.com.", d.TypeA)
	reply, err := doh.HandleQuery(context.Background(), m, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Answer) == 0 {
		t.Fatal("expected at least one answer")
	}
	log.Print("Reply:", reply)
}
