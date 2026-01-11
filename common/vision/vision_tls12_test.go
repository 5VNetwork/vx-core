package vision_test

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/vision"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
)

var clientHello12 []byte
var serverHello12 []byte
var serverFinished12 []byte
var clientFinished12 []byte
var clientApplicationData12 []byte
var serverAppData []byte
var clientCloseNotify []byte

func TestVConnClientTls12(t *testing.T) {
	// test.InitZeroLog()
	ctx := log.Logger.WithContext(context.Background())
	// run server
	t.Run("run server", testVConnServerTls12)
	t.Run("run client", func(t *testing.T) {
		ctx := session.ContextWithInfo(ctx, &session.Info{ID: session.ID(0)})
		conn, err := net.Dial("tcp", "localhost:10001")
		common.Must(err)
		t.Log("client dialed server successfully")
		conn = tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
		t.Parallel()
		b := make([]byte, 5012)

		vConn := vision.NewVisionConn(ctx, conn, true, 0)
		defer vConn.Close()
		// write clientHello
		_, err = vConn.Write(clientHello12)
		if err != nil {
			t.Fatal("client failed to write clientHello", err)
		}
		t.Log("client write clientHello successfully")

		// read serverHello
		n, err := vConn.Read(b)
		if err != nil {
			t.Fatal("client failed to read serverHello", err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverHello12) {
			t.Fatal("serverHello not equal")
		}
		t.Log("client read serverHello successfully")

		// write clientFInished
		_, err = vConn.Write(clientFinished12)
		if err != nil {
			t.Fatal("client failed to write clientFinished", err)
		}
		t.Log("client write clientFinished successfully")

		// read serverFinished
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverFinished12) {
			t.Fatal("serverFinished12 not equal")
		}
		t.Log("client read serverFinished12 successfully")

		common.Must2(vConn.Write(clientApplicationData12))
		t.Log("client write clientApplicationData successfully")
		// read applicationData from server
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if cmp.Diff(b[:n], serverAppData) != "" {
			t.Fatal("applicationData not equal")
		}
		t.Log("client read applicationData successfully")

		// write closeNotify
		_, err = vConn.Write(clientCloseNotify)
		if err != nil {
			t.Fatal("client failed to write closeNotify", err)
		}
		t.Log("client write closeNotify successfully")

		vConn.Close()
		time.Sleep(1 * time.Second)
	})
}

func testVConnServerTls12(t *testing.T) {
	ctx := session.ContextWithInfo(context.Background(), &session.Info{ID: session.ID(0)})

	l, err := net.Listen("tcp", "localhost:10001")
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
	// read clientHello
	b := make([]byte, 2048)
	n, err := conn.Read(b)
	if err != nil {
		t.Fatal("read ClientHello failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientHello12) {
		t.Fatal("clientHello not equal")
	}
	t.Log("server read clientHello succesfully")
	t.Log("serverHello len", len(serverHello12))
	common.Must2(conn.Write(serverHello12))
	t.Log("server write serverHello succesfully")

	// read clientFinished
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read clientFinished failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientFinished12) {
		t.Fatal("clientFinished not equal")
	}
	t.Log("server read clientFinished succesfully")

	common.Must2(conn.Write(serverFinished12))
	t.Log("server write serverFinished succesfully")

	// read applicationData
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read app data failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientApplicationData12) {
		t.Fatal("applicationData not equal")
	}

	_, err = conn.Write(serverAppData)
	if err != nil {
		t.Fatal("server failed to write app data ", err)
	}
	t.Log("server write app data succesfully")

	// read closeNotify
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read closeNotify failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientCloseNotify) {
		t.Fatal("closeNotify not equal")
	}
	t.Log("server read closeNotify succesfully")

	time.Sleep(1 * time.Second)
}

func init() {
	chRaw := "16 03 01 00 a5 01 00 00 a1 03 03 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 18 19 1a 1b 1c 1d 1e 1f 00 00 20 cc a8 cc a9 c0 2f c0 30 c0 2b c0 2c c0 13 c0 09 c0 14 c0 0a 00 9c 00 9d 00 2f 00 35 c0 12 00 0a 01 00 00 58 00 00 00 18 00 16 00 00 13 65 78 61 6d 70 6c 65 2e 75 6c 66 68 65 69 6d 2e 6e 65 74 00 05 00 05 01 00 00 00 00 00 0a 00 0a 00 08 00 1d 00 17 00 18 00 19 00 0b 00 02 01 00 00 0d 00 12 00 10 04 01 04 03 05 01 05 03 06 01 06 03 02 01 02 03 ff 01 00 01 00 00 12 00 00"
	clientHello12, _ = hex.DecodeString(strings.ReplaceAll(chRaw, " ", ""))
	shRaw := "16 03 03 00 31 02 00 00 2d 03 03 70 71 72 73 74 75 76 77 78 79 7a 7b 7c 7d 7e 7f 80 81 82 83 84 85 86 87 88 89 8a 8b 8c 8d 8e 8f 00 c0 13 00 00 05 ff 01 00 01 0016 03 03 03 2f 0b 00 03 2b 00 03 28 00 03 25 30 82 03 21 30 82 02 09 a0 03 02 01 02 02 08 15 5a 92 ad c2 04 8f 90 30 0d 06 09 2a 86 48 86 f7 0d 01 01 0b 05 00 30 22 31 0b 30 09 06 03 55 04 06 13 02  55 53 31 13 30 11 06 03 55 04 0a 13 0a 45 78 61 6d 70 6c 65 20 43 41 30 1e 17 0d 31 38 31 30 30 35 30	   31 33 38 31 37 5a 17 0d 31 39 31 30 30 35 30 31 33 38 31 37 5a 30 2b 31 0b 30 09 06 03 55 04 06 13 02	    55 53 31 1c 30 1a 06 03 55 04 03 13 13 65 78 61 6d 70 6c 65 2e 75 6c 66 68 65 69 6d 2e 6e 65 74 30 		82 01 22 30 0d 06 09 2a 86 48 86 f7 0d 01 01 01 05 00 03 82 01 0f 00 30 82 01 0a 02 82 01 01 00 c4 		80 36 06 ba e7 47 6b 08 94 04 ec a7 b6 91 04 3f f7 92 bc 19 ee fb 7d 74 d7 a8 0d 00 1e 7b 4b 3a 4a 		e6 0f e8 c0 71 fc 73 e7 02 4c 0d bc f4 bd d1 1d 39 6b ba 70 46 4a 13 e9 4a f8 3d f3 e1 09 59 54 7b		 c9 55 fb 41 2d a3 76 52 11 e1 f3 dc 77 6c aa 53 37 6e ca 3a ec be c3 aa b7 3b 31 d5 6c b6 52 9c 80		  98 bc c9 e0 28 18 e2 0b f7 f8 a0 3a fd 17 04 50 9e ce 79 bd 9f 39 f1 ea 69 ec 47 97 2e 83 0f b5 ca		   95 de 95 a1 e6 04 22 d5 ee be 52 79 54 a1 e7 bf 8a 86 f6 46 6d 0d 9f 16 95 1a 4c f7 a0 46 92 59 		   5c 13 52 f2 54 9e 5a fb 4e bf d7 7a 37 95 01 44 e4 c0 26 87 4c 65 3e 40 7d 7d 23 07 44 01 f4 84 		   ff d0 8f 7a 1f a0 52 10 d1 f4 f0 d5 ce 79 70 29 32 e2 ca be 70 1f df ad 6b 4b b7 11 01 f4 4b ad 		   66 6a 11 13 0f e2 ee 82 9e 4d 02 9d c9 1c dd 67 16 db b9 06 18 86 ed c1 ba 94 21 02 03 01 00 01 		   a3 52 30 50 30 0e 06 03 55 1d 0f 01 01 ff 04 04 03 02 05 a0 30 1d 06 03 55 1d 25 04 16 30 14 06 		   08 2b 06 01 05 05 07 03 02 06 08 2b 06 01 05 05 07 03 01 30 1f 06 03 55 1d 23 04 18 30 16 80 14 		   89 4f de 5b cc 69 e2 52 cf 3e a3 00 df b1 97 b8 1d e1 c1 46 30 0d 06 09 2a 86 48 86 f7 0d 01 01 		   0b 05 00 03 82 01 01 00 59 16 45 a6 9a 2e 37 79 e4 f6 dd 27 1a ba 1c 0b fd 6c d7 55 99 b5 e7 c3		    6e 53 3e ff 36 59 08 43 24 c9 e7 a5 04 07 9d 39 e0 d4 29 87 ff e3 eb dd 09 c1 cf 1d 91 44 55 87			 0b 57 1d d1 9b df 1d 24 f8 bb 9a 11 fe 80 fd 59 2b a0 39 8c de 11 e2 65 1e 61 8c e5 98 fa 96 			 e5 37 2e ef 3d 24 8a fd e1 74 63 eb bf ab b8 e4 d1 ab 50 2a 54 ec 00 64 e9 2f 78 19 66 0d 3f 			 27 cf 20 9e 66 7f ce 5a e2 e4 ac 99 c7 c9 38 18 f8 b2 51 07 22 df ed 97 f3 2e 3e 93 49 d4 c6 			 6c 9e a6 39 6d 74 44 62 a0 6b 42 c6 d5 ba 68 8e ac 3a 01 7b dd fc 8e 2c fc ad 27 cb 69 d3 cc			  dc a2 80 41 44 65 d3 ae 34 8c e0 f3 4a b2 fb 9c 61 83 71 31 2b 19 10 41 64 1c 23 7f 11 a5 d6			   5c 84 4f 04 04 84 99 38 71 2b 95 9e d6 85 bc 5c 5d d6 45 ed 19 90 94 73 40 29 26 dc b4 0e			    34 69 a1 59 41 e8 e2 cc a8 4b b6 08 46 36 a0 16 03 03 01 2c 0c 00 01 28 03 00 1d 20 9f d7 ad 				6d cf f4 29 8d d3 f9 6d 5b 1b 2a f9 10 a0 53 5b 14 88 d7 f8 fa bb 34 9a 98 28 80 b6 15 04 01				 01 00 04 02 b6 61 f7 c1 91 ee 59 be 45 37 66 39 bd c3 d4 bb 81 e1 15 ca 73 c8 34 8b 52 5b 0d 				 23 38 aa 14 46 67 ed 94 31 02 14 12 cd 9b 84 4c ba 29 93 4a aa cc e8 73 41 4e c1 1c b0 2e 27 				 2d 0a d8 1f 76 7d 33 07 67 21 f1 3b f3 60 20 cf 0b 1f d0 ec b0 78 de 11 28 be ba 09 49 eb ec				  e1 a1 f9 6e 20 9d c3 6e 4f ff d3 6b 67 3a 7d dc 15 97 ad 44 08 e4 85 c4 ad b2 c8 73 84 12 49				   37 25 23 80 9e 43 12 d0 c7 b3 52 2e f9 83 ca c1 e0 39 35 ff 13 a8 e9 6b a6 81 a6 2e 40 d3 				   e7 0a 7f f3 58 66 d3 d9 99 3f 9e 26 a6 34 c8 1b 4e 71 38 0f cd d6 f4 e8 35 f7 5a 64 09 c7 				   dc 2c 07 41 0e 6f 87 85 8c 7b 94 c0 1c 2e 32 f2 91 76 9e ac ca 71 64 3b 8b 98 a9 63 df 0a 32				    9b ea 4e d6 39 7e 8c d0 1a 11 0a b3 61 ac 5b ad 1c cd 84 0a 6c 8a 6e aa 00 1a 9d 7d 87 dc 33					 18 64 35 71 22 6c 4d d2 c2 ac 41 fb 16 03 03 00 04 0e 00 00 00"
	serverHello12, _ = hex.DecodeString(strings.ReplaceAll(shRaw, " ", ""))
	cfRaw := "16 03 03 00 25 10 00 00 21 20 35 80 72 d6 36 58 80 d1 ae ea 32 9a df 91 21 38 38 51 ed 21 a2 8e 3b 75 e9 65 d0 d2 cd 16 62 54 14 03 03 00 01 01 16 03 03 00 40 40 41 42 43 44 45 46 47 48 49 4a 4b 4c 4d 4e 4f 22 7b c9 ba 81 ef 30 f2 a8 a7 8f f1 df 50 84 4d 58 04 b7 ee b2 e2 14 c3 2b 68 92 ac a3 db 7b 78 07 7f dd 90 06 7c  51 6b ac b3 ba 90 de df 72 0f"
	clientFinished12, _ = hex.DecodeString(strings.ReplaceAll(cfRaw, " ", ""))
	sfRaw := "14 03 03 00 01 0116 03 03 00 40 51 52 53 54 55 56 57 58 59 5a 5b 5c 5d 5e 5f 60 18 e0 75 31 7b 10 03 15 f6 08 1f cb f3 13 78 1a ac 73 ef e1 9f e25b a1 af 59 c2 0b e9 4f c0 1b da 2d 68 00 29 8b 73 a7 e8 49 d7 4b d4 94 cf 7d"
	serverFinished12, _ = hex.DecodeString(strings.ReplaceAll(sfRaw, " ", ""))
	caRaw := "17 03 03 00 30 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 6c 42 1c 71 c4 2b 18 3b fa 06 19 5d 13 3d 0a 09 d0 0f c7 cb 4e 0f 5d 1c da 59 d1 47 ec 79 0c 99"
	clientApplicationData12, _ = hex.DecodeString(strings.ReplaceAll(caRaw, " ", ""))
	saRaw := "17 03 03 00 30 61 62 63 64 65 66 67 68 69 6a 6b 6c 6d 6e 6f 70 97 83 48 8a f5 fa 20 bf 7a 2e f6 9d eb b5 34 db 9f b0 7a 8c 27 21 de e5 40 9f 77 af 0c 3d de 56"
	serverAppData, _ = hex.DecodeString(strings.ReplaceAll(saRaw, " ", ""))
	ccnRaw := "15 03 03 00 30 10 11 12 13 14 15 16 17 18 19 1a 1b 1c 1d 1e 1f 0d 83 f9 79 04 75 0d d8 fd 8a a1 30 21 86 32 63 4f d0 65 e4 62 83 79 b8 8b bf 9e fd 12 87 a6 2d"
	clientCloseNotify, _ = hex.DecodeString(strings.ReplaceAll(ccnRaw, " ", ""))
}
