package dlhelper_test

import (
	"context"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/test/servers/tcp"

	. "github.com/5vnetwork/vx-core/transport/dlhelper"

	"github.com/google/go-cmp/cmp"
)

func TestTCPFastOpen(t *testing.T) {
	tcpServer := tcp.Server{
		MsgProcessor: func(b []byte) []byte {
			return b
		},
	}
	dest, err := tcpServer.StartContext(context.Background(), &SocketSetting{Tfo: SocketConfig_Enable})
	common.Must(err)
	defer tcpServer.Close()

	ctx := context.Background()
	dialer := DefaultSystemDialer{}
	conn, err := dialer.DialConn(ctx, dest, &SocketSetting{
		Tfo: SocketConfig_Enable,
	})
	common.Must(err)
	defer conn.Close()

	_, err = conn.Write([]byte("abcd"))
	common.Must(err)

	b := buf.New()
	common.Must2(b.ReadOnce(conn))
	if r := cmp.Diff(b.Bytes(), []byte("abcd")); r != "" {
		t.Fatal(r)
	}
}
