package vision_test

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/vision"
	"github.com/5vnetwork/vx-core/test/servers/tcp"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
)

var clientHello13 []byte
var serverHello13 []byte
var serverOthers13 []byte
var serverCert13 []byte
var serverCertVerify13 []byte
var serverFinished13 []byte
var clientOthers13 []byte
var clientApplicationData13 []byte

func init() {
	clientHello13, _ = hex.DecodeString("16030100f8010000f40303000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20e0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfeff000813021303130100ff010000a30000001800160000136578616d706c652e756c666865696d2e6e6574000b000403000102000a00160014001d0017001e0019001801000101010201030104002300000016000000170000000d001e001c040305030603080708080809080a080b080408050806040105010601002b0003020304002d00020101003300260024001d0020358072d6365880d1aeea329adf9121383851ed21a28e3b75e965d0d2cd166254")
	serverHello13, _ = hex.DecodeString("160303007a020000760303707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f20e0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfeff130200002e002b0002030400330024001d00209fd7ad6dcff4298dd3f96d5b1b2af910a0535b1488d7f8fabb349a982880b615")
	serverOthers13, _ = hex.DecodeString("14030300010117030300176be02f9da7c2dc9ddef56f2468b90adfa25101ab0344ae")
	serverCert13, _ = hex.DecodeString("1703030343baf00a9be50f3f2307e726edcbdacbe4b18616449d46c6207af6e9953ee5d2411ba65d31feaf4f78764f2d693987186cc01329c187a5e4608e8d27b318e98dd94769f7739ce6768392caca8dcc597d77ec0d1272233785f6e69d6f43effa8e7905edfdc4037eee5933e990a7972f206913a31e8d04931366d3d8bcd6a4a4d647dd4bd80b0ff863ce3554833d744cf0e0b9c07cae726dd23f9953df1f1ce3aceb3b7230871e92310cfb2b098486f43538f8e82d8404e5c6c25f66a62ebe3c5f26232640e20a769175ef83483cd81e6cb16e78dfad4c1b714b04b45f6ac8d1065ad18c13451c9055c47da300f93536ea56f531986d6492775393c4ccb095467092a0ec0b43ed7a0687cb470ce350917b0ac30c6e5c24725a78c45f9f5f29b6626867f6f79ce054273547b36df030bd24af10d632dba54fc4e890bd0586928c0206ca2e28e44e227a2d5063195935df38da8936092eef01e84cad2e49d62e470a6c7745f625ec39e4fc23329c79d1172876807c36d736ba42bb69b004ff55f93850dc33c1f98abb92858324c76ff1eb085db3c1fc50f74ec04442e622973ea70743418794c388140bb492d6294a0540e5a59cfae60ba0f14899fca71333315ea083a68e1d7c1e4cdc2f56bcd6119681a4adbc1bbf42afd806c3cbd42a076f545dee4e118d0b396754be2b042a685dd4727e89c0386a94d3cd6ecb9820e9d49afeed66c47e6fc243eabebbcb0b02453877f5ac5dbfbdf8db1052a3c994b224cd9aaaf56b026bb9efa2e01302b36401ab6494e7018d6e5b573bd38bcef023b1fc92946bbca0209ca5fa926b4970b1009103645cb1fcfe552311ff730558984370038fd2cce2a91fc74d6f3e3ea9f843eed356f6f82d35d03bc24b81b58ceb1a43ec9437e6f1e50eb6f555e321fd67c8332eb1b832aa8d795a27d479c6e27d5a61034683891903f66421d094e1b00a9a138d861e6f78a20ad3e1580054d2e305253c713a02fe1e28deee7336246f6ae34331806b46b47b833c39b9d31cd300c2a6ed831399776d07f570eaf0059a2c68a5f3ae16b617404af7b7231a4d942758fc020b3f23ee8c15e36044cfd67cd640993b16207597fbf385ea7a4d99e8d456ff83d41f7b8b4f069b028a2a63a919a70e3a10e3084158faa5bafa30186c6b2f238eb530c73e")
	serverCertVerify13, _ = hex.DecodeString("170303011973719fce07ec2f6d3bba0292a0d40b2770c06a271799a53314f6f77fc95c5fe7b9a4329fd9548c670ebeea2f2d5c351dd9356ef2dcd52eb137bd3a676522f8cd0fb7560789ad7b0e3caba2e37e6b4199c6793b3346ed46cf740a9fa1fec414dc715c415c60e575703ce6a34b70b5191aa6a61a18faff216c687ad8d17e12a7e99915a611bfc1a2befc15e6e94d784642e682fd17382a348c301056b940c9847200408bec56c81ea3d7217ab8e85a88715395899c90587f72e8ddd74b26d8edc1c7c837d9f2ebbc260962219038b05654a63a0b12999b4a8306a3ddcc0e17c53ba8f9c80363f7841354d291b4ace0c0f330c0fcd5aa9deef969ae8ab2d98da88ebb6ea80a3a11f00ea296a3232367ff075e1c66dd9cbedc4713")
	serverFinished13, _ = hex.DecodeString("17030300451061de27e51c2c9f342911806f282b710c10632ca5006755880dbf7006002d0e84fed9adf27a43b5192303e4df5c285d58e3c76224078440c0742374744aecf28cf3182fd0")
	clientOthers13, _ = hex.DecodeString("14030300010117030300459ff9b063175177322a46dd9896f3c3bb820ab51743ebc25fdadd53454b73deb54cc7248d411a18bccf657a960824e9a19364837c350a69a88d4bf635c85eb874aebc9dfde8")
	clientApplicationData13, _ = hex.DecodeString("1703030015828139cb7b73aaabf5b82fbf9a2961bcde10038a32")
}

var p uint16

func TestVConnClientTls13(t *testing.T) {
	// test.InitZeroLog()
	ctx := log.Logger.WithContext(context.Background())
	p = uint16(tcp.PickPort())
	t.Logf("port: %d", p)
	// run server
	t.Run("run server", testVConnServerTls13)
	t.Run("run client", func(t *testing.T) {
		ctx := session.ContextWithInfo(ctx, &session.Info{ID: session.ID(0)})

		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", p))
		common.Must(err)
		t.Log("client dialed server successfully")
		conn = tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
		t.Parallel()
		b := make([]byte, 5012)

		vConn := vision.NewVisionConn(ctx, conn, true, 0)
		defer vConn.Close()
		_, err = vConn.Write(clientHello13)
		if err != nil {
			t.Fatal("client failed to write clientHello", err)
		}
		t.Log("client write clientHello successfully")

		// read serverHello
		n, err := vConn.Read(b)
		if err != nil {
			t.Fatal("client failed to read serverHello", err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverHello13) {
			t.Fatal("serverHello not equal")
		}
		t.Log("client read serverHello successfully")

		// read serverOthers
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverOthers13) {
			t.Fatal("serverOthers not equal")
		}
		t.Log("client read serverOthers successfully")

		// read serverCert
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverCert13) {
			t.Fatal("serverOthers not equal")
		}
		t.Log("client read serverCert successfully")

		// read serverCertVerify
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:len(serverCertVerify13)]) != hex.EncodeToString(serverCertVerify13) {
			t.Fatal("serverCertVerify not equal")
		}
		t.Log("client read serverCertVerify successfully")

		// read serverFinished
		if n > len(serverCertVerify13) {
			if hex.EncodeToString(b[len(serverCertVerify13):n]) != hex.EncodeToString(serverFinished13) {
				t.Fatal("serverFinished not equal")
			}
		} else {
			n, err = vConn.Read(b)
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverFinished13) {
				t.Fatal("serverFinished not equal")
			}
		}
		t.Log("client read serverFinished successfully")
		common.Must2(vConn.Write(clientOthers13))
		t.Log("client write clientOthers successfully")
		common.Must2(vConn.Write(clientApplicationData13))
		t.Log("client write clientApplicationData successfully")
		// read applicationData from server
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if cmp.Diff(b[:n], []byte("hello")) != "" {
			t.Fatal("applicationData not equal")
		}
		t.Log("client read applicationData successfully")
	})
}

func testVConnServerTls13(t *testing.T) {
	ctx := session.ContextWithInfo(context.Background(), &session.Info{ID: session.ID(0)})
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", p))
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
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientHello13) {
		t.Fatal("clientHello not equal")
	}
	t.Log("server read clientHello succesfully")
	common.Must2(conn.Write(serverHello13))
	t.Log("server write serverHello succesfully")
	common.Must2(conn.Write(serverOthers13))
	t.Log("server write serverOthers succesfully")
	common.Must2(conn.Write(serverCert13))
	t.Log("server write serverCert succesfully")
	common.Must2(conn.Write(serverCertVerify13))
	t.Log("server write serverCertVerify succesfully")
	common.Must2(conn.Write(serverFinished13))
	t.Log("server write serverFinished succesfully")
	// read clientOthers
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read clientOthers failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientOthers13) {
		t.Fatal("clientOthers not equal")
	}
	// read applicationData
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read app data failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientApplicationData13) {
		t.Fatal("applicationData not equal")
	}

	_, err = conn.Write([]byte("hello"))
	if err != nil {
		t.Fatal("server failed to write app data ", err)
	}
	t.Log("server write app data succesfully")
	time.Sleep(1 * time.Second)
}

func TestVConnClientTls13WithProxyHeader(t *testing.T) {
	// test.InitZeroLog()
	ctx := log.Logger.WithContext(context.Background())
	p = uint16(tcp.PickPort())
	// run server
	t.Run("run  server", testVConnServerTls13withProxyHeader)
	t.Run("run client", func(t *testing.T) {
		ctx := session.ContextWithInfo(ctx, &session.Info{ID: session.ID(0)})

		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", p))
		common.Must(err)
		t.Log("client dialed server successfully")
		conn = tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
		})
		t.Parallel()
		b := make([]byte, 5012)

		vConn := vision.NewVisionConn(ctx, conn, true, 1)
		defer vConn.Close()

		// _, err = vConn.Write(nil)
		// if err != nil {
		// 	t.Fatal("client failed to write empty", err)
		// }
		// t.Log("client write empty successfully")
		var header [1]byte
		_, err = vConn.Write(append(header[:], clientHello13...))
		if err != nil {
			t.Fatal("client failed to write clientHello", err)
		}
		t.Log("client write clientHello successfully")

		// read serverHello
		n, err := vConn.Read(b)
		if err != nil {
			t.Fatal("client failed to read serverHello", err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverHello13) {
			t.Fatal("serverHello not equal")
		}
		t.Log("client read serverHello successfully")

		// read serverOthers
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverOthers13) {
			t.Fatal("serverOthers not equal")
		}
		t.Log("client read serverOthers successfully")

		// read serverCert
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverCert13) {
			t.Fatal("serverOthers not equal")
		}
		t.Log("client read serverCert successfully")

		// read serverCertVerify
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if hex.EncodeToString(b[:len(serverCertVerify13)]) != hex.EncodeToString(serverCertVerify13) {
			t.Fatal("serverCertVerify not equal")
		}
		t.Log("client read serverCertVerify successfully")

		// read serverFinished
		if n > len(serverCertVerify13) {
			if hex.EncodeToString(b[len(serverCertVerify13):n]) != hex.EncodeToString(serverFinished13) {
				t.Fatal("serverFinished not equal")
			}
		} else {
			n, err = vConn.Read(b)
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(b[:n]) != hex.EncodeToString(serverFinished13) {
				t.Fatal("serverFinished not equal")
			}
		}
		t.Log("client read serverFinished successfully")
		common.Must2(vConn.Write(clientOthers13))
		t.Log("client write clientOthers successfully")
		common.Must2(vConn.Write(clientApplicationData13))
		t.Log("client write clientApplicationData successfully")
		// read applicationData from server
		n, err = vConn.Read(b)
		if err != nil {
			t.Fatal(err)
		}
		if cmp.Diff(b[:n], []byte("hello")) != "" {
			t.Fatal("applicationData not equal")
		}
		t.Log("client read applicationData successfully")
	})
}

func testVConnServerTls13withProxyHeader(t *testing.T) {
	// test.InitZeroLog()
	ctx := log.Logger.WithContext(context.Background())
	ctx = session.ContextWithInfo(ctx, &session.Info{ID: session.ID(0)})

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", p))
	common.Must(err)
	t.Log("server started")
	t.Parallel()
	conn, err := l.Accept()
	common.Must(err)
	t.Log("client connected")
	certBytes, _ := os.ReadFile("../rsa_cert.pem")
	keyBytes, _ := os.ReadFile("../rsa_private.pem")
	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	common.Must(err)
	conn = tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})

	n, err := io.ReadFull(conn, make([]byte, 1))
	if err != nil || n != 1 {
		t.Fatal("failed to read proxy header", err)
	}

	conn = vision.NewVisionConn(ctx, conn, false, 0)
	conn.SetDeadline(time.Time{})
	defer conn.Close()
	// read clientHello
	b := make([]byte, 2048)
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("read ClientHello failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientHello13) {
		t.Fatal("clientHello not equal")
	}
	t.Log("server read clientHello succesfully")
	common.Must2(conn.Write(serverHello13))
	t.Log("server write serverHello succesfully")
	common.Must2(conn.Write(serverOthers13))
	t.Log("server write serverOthers succesfully")
	common.Must2(conn.Write(serverCert13))
	t.Log("server write serverCert succesfully")
	common.Must2(conn.Write(serverCertVerify13))
	t.Log("server write serverCertVerify succesfully")
	common.Must2(conn.Write(serverFinished13))
	t.Log("server write serverFinished succesfully")
	// read clientOthers
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read clientOthers failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientOthers13) {
		t.Fatal("clientOthers not equal")
	}
	// read applicationData
	n, err = conn.Read(b)
	if err != nil {
		t.Fatal("server read app data failed", err)
	}
	if hex.EncodeToString(b[:n]) != hex.EncodeToString(clientApplicationData13) {
		t.Fatal("applicationData not equal")
	}

	_, err = conn.Write([]byte("hello"))
	if err != nil {
		t.Fatal("server failed to write app data ", err)
	}
	t.Log("server write app data succesfully")
	time.Sleep(1 * time.Second)
}
