package vision_test

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/vision"

	"github.com/google/go-cmp/cmp"
)

func TestVisionOthers(t *testing.T) {
	t.Run("run server", testVConnServerOthers)

	t.Run("run client", func(t *testing.T) {
		ctx := session.ContextWithInfo(context.Background(), &session.Info{ID: session.ID(0)})
		conn, err := net.Dial("tcp", "localhost:10000")
		common.Must(err)
		t.Log("client dialed server successfully")
		conn = tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
		t.Parallel()
		b := make([]byte, 5012)

		vConn := vision.NewVisionConn(ctx, conn, true, 0)
		defer vConn.Close()

		for i := 0; i < 5; i++ {
			_, err := vConn.Write([]byte{byte(i)})
			if err != nil {
				t.Fatal("cannotd write", err)
			}
			n, err := vConn.Read(b)
			if err != nil {
				t.Fatal(err)
			}
			if cmp.Diff(b[:n], []byte{byte(i)}) != "" {
				t.Fatal("not equal")
			}
		}
		time.Sleep(time.Second)
	})
}

func testVConnServerOthers(t *testing.T) {
	ctx := session.ContextWithInfo(context.Background(), &session.Info{ID: session.ID(0)})
	l, err := net.Listen("tcp", "localhost:10000")
	common.Must(err)
	t.Log("server started")
	t.Parallel()
	conn, err := l.Accept()
	common.Must(err)
	t.Log("client connected")
	certBytes, _ := os.ReadFile("rsa_cert.pem")
	keyBytes, _ := os.ReadFile("rsa_private.pem")
	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	common.Must(err)
	conn = tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	conn = vision.NewVisionConn(ctx, conn, false, 0)
	conn.SetDeadline(time.Time{})
	defer conn.Close()

	b := make([]byte, 5012)
	for i := 0; i < 5; i++ {
		n, err := conn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Write(b[:n])
		if err != nil {
			t.Fatal(err)
		}
	}
}
