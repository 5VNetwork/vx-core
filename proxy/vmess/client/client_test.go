package client

import (
	"context"
	"io"
	stdnet "net"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/app/user"
	"github.com/5vnetwork/vx-core/common/mux"
	vxnet "github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy/vmess"
	"github.com/5vnetwork/vx-core/proxy/vmess/encoding"
)

type singleServerPicker struct {
	server *protocol.ServerSpec
}

func (p *singleServerPicker) AddServer(server *protocol.ServerSpec) {
	p.server = server
}

func (p *singleServerPicker) PickServer() *protocol.ServerSpec {
	return p.server
}

type pipeDialer struct {
	clientConn vxnet.Conn
	dialedDst  vxnet.Destination
}

func (d *pipeDialer) Dial(_ context.Context, dst vxnet.Destination) (vxnet.Conn, error) {
	d.dialedDst = dst
	return d.clientConn, nil
}

type decodedRequestResult struct {
	request *protocol.RequestHeader
	body    []byte
	err     error
}

func newTestClient(t *testing.T, clientConn vxnet.Conn, account *vmess.MemoryAccount, serverDest vxnet.Destination) (*Client, *pipeDialer) {
	t.Helper()

	dialer := &pipeDialer{clientConn: clientConn}
	picker := &singleServerPicker{
		server: protocol.NewServerSpec(serverDest, protocol.AlwaysValid(), account),
	}

	return NewClient(ClientSettings{
		ServerPicker: picker,
		Dialer:       dialer,
	}), dialer
}

func startServerDecode(t *testing.T, serverConn stdnet.Conn, account *vmess.MemoryAccount, bodyLen int) <-chan decodedRequestResult {
	t.Helper()

	resultCh := make(chan decodedRequestResult, 1)

	go func() {
		defer close(resultCh)
		defer serverConn.Close()

		validator := vmess.NewTimedUserValidator(protocol.DefaultIDHash)
		defer validator.Close()
		if err := validator.Add(account); err != nil {
			resultCh <- decodedRequestResult{err: err}
			return
		}

		sessionHistory := encoding.NewSessionHistory()
		defer sessionHistory.Close()

		session := encoding.NewServerSession(validator, sessionHistory)
		request, err := session.DecodeRequestHeader(serverConn, true)
		if err != nil {
			resultCh <- decodedRequestResult{err: err}
			return
		}

		result := decodedRequestResult{request: request}
		if bodyLen > 0 {
			bodyReader, err := session.DecodeRequestBody1(context.Background(), request, serverConn)
			if err != nil {
				resultCh <- decodedRequestResult{err: err}
				return
			}

			result.body = make([]byte, bodyLen)
			if _, err := io.ReadFull(bodyReader, result.body); err != nil {
				resultCh <- decodedRequestResult{err: err}
				return
			}
		}

		resultCh <- result
	}()

	return resultCh
}

func newTestMemoryAccount() *vmess.MemoryAccount {
	userID := uuid.New()
	secret := uuid.New()
	return vmess.NewMemoryAccount(
		user.NewUser(userID.String(), secret.String()),
		0,
		protocol.SecurityType_NONE,
		false,
		false,
	)
}

func TestDialWithInitialData_WritesInitialData(t *testing.T) {
	t.Parallel()

	clientSide, serverSide := stdnet.Pipe()
	deadline := time.Now().Add(5 * time.Second)
	_ = clientSide.SetDeadline(deadline)
	_ = serverSide.SetDeadline(deadline)

	account := newTestMemoryAccount()
	serverDest := vxnet.TCPDestination(vxnet.LocalHostIP, 443)
	targetDest := vxnet.TCPDestination(vxnet.DomainAddress("example.com"), 80)
	initialData := []byte("hello")
	followUp := []byte(" world")

	resultCh := startServerDecode(t, serverSide, account, len(initialData)+len(followUp))

	client, dialer := newTestClient(t, clientSide, account, serverDest)
	conn, err := client.DialWithInitialData(context.Background(), targetDest, initialData)
	if err != nil {
		t.Fatalf("DialWithInitialData failed: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write(followUp); err != nil {
		t.Fatalf("follow-up write failed: %v", err)
	}

	result := <-resultCh
	if result.err != nil {
		t.Fatalf("server decode failed: %v", result.err)
	}

	if got := dialer.dialedDst; got != serverDest {
		t.Fatalf("dialed destination = %v, want %v", got, serverDest)
	}
	if got := result.request.Command; got != protocol.RequestCommandTCP {
		t.Fatalf("request command = %v, want %v", got, protocol.RequestCommandTCP)
	}
	if got := result.request.Destination(); got != targetDest {
		t.Fatalf("request destination = %v, want %v", got, targetDest)
	}
	if got := string(result.body); got != string(append(append([]byte(nil), initialData...), followUp...)) {
		t.Fatalf("decoded body = %q, want %q", got, string(append(initialData, followUp...)))
	}
}

func TestDialWithInitialData_SelectsExpectedCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dst     vxnet.Destination
		command protocol.RequestCommand
		wantDst vxnet.Destination
	}{
		{
			name:    "tcp",
			dst:     vxnet.TCPDestination(vxnet.DomainAddress("example.com"), 80),
			command: protocol.RequestCommandTCP,
			wantDst: vxnet.TCPDestination(vxnet.DomainAddress("example.com"), 80),
		},
		{
			name:    "udp",
			dst:     vxnet.UDPDestination(vxnet.DomainAddress("example.com"), 53),
			command: protocol.RequestCommandUDP,
			wantDst: vxnet.UDPDestination(vxnet.DomainAddress("example.com"), 53),
		},
		{
			name:    "mux",
			dst:     vxnet.TCPDestination(mux.MuxCoolAddressDst, mux.MuxCoolPortDst),
			command: protocol.RequestCommandMux,
			wantDst: vxnet.TCPDestination(mux.MuxCoolAddressDst, 0),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clientSide, serverSide := stdnet.Pipe()
			deadline := time.Now().Add(5 * time.Second)
			_ = clientSide.SetDeadline(deadline)
			_ = serverSide.SetDeadline(deadline)

			account := newTestMemoryAccount()
			serverDest := vxnet.TCPDestination(vxnet.LocalHostIP, 443)
			resultCh := startServerDecode(t, serverSide, account, 0)

			client, _ := newTestClient(t, clientSide, account, serverDest)
			conn, err := client.DialWithInitialData(context.Background(), tt.dst, nil)
			if err != nil {
				t.Fatalf("DialWithInitialData failed: %v", err)
			}
			defer conn.Close()

			result := <-resultCh
			if result.err != nil {
				t.Fatalf("server decode failed: %v", result.err)
			}
			if got := result.request.Command; got != tt.command {
				t.Fatalf("request command = %v, want %v", got, tt.command)
			}
			if got := result.request.Destination(); got != tt.wantDst {
				t.Fatalf("request destination = %v, want %v", got, tt.wantDst)
			}
		})
	}
}

var _ i.Dialer = (*pipeDialer)(nil)
