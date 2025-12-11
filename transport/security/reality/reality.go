package reality

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	gotls "crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	sync "sync"
	"time"
	"unsafe"

	"github.com/5vnetwork/vx-core/common/crypto"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/transport/security"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/xtls/reality"

	utls "github.com/refraction-networking/utls"
	"github.com/rs/zerolog/log"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/net/http2"
)

type Engine struct {
	Config *RealityConfig
}

func NewEngine(config *RealityConfig) (*Engine, error) {
	if len(config.PublicKey) == 0 && config.Pbk != "" {
		key, err := base64.URLEncoding.DecodeString(config.Pbk)
		if err != nil {
			key, err = base64.RawURLEncoding.DecodeString(config.Pbk)
			if err != nil {
				return nil, err
			}
		}
		if len(key) != 32 {
			return nil, errors.New("REALITY: key length is not 32")
		}
		config.PublicKey = key
	}
	if len(config.ShortId) == 0 && config.Sid != "" {
		config.ShortId = make([]byte, 8)
		if _, err := hex.Decode(config.ShortId, []byte(config.Sid)); err != nil {
			return nil, err
		}
	}

	if config.SpiderX == "" {
		config.SpiderX = "/"
	}
	if config.SpiderX[0] != '/' {
		return nil, errors.New(`invalid "spiderX": `, config.SpiderX)
	}
	config.SpiderY = make([]int64, 10)
	u, _ := url.Parse(config.SpiderX)
	q := u.Query()
	parse := func(param string, index int) {
		if q.Get(param) != "" {
			s := strings.Split(q.Get(param), "-")
			if len(s) == 1 {
				config.SpiderY[index], _ = strconv.ParseInt(s[0], 10, 64)
				config.SpiderY[index+1], _ = strconv.ParseInt(s[0], 10, 64)
			} else {
				config.SpiderY[index], _ = strconv.ParseInt(s[0], 10, 64)
				config.SpiderY[index+1], _ = strconv.ParseInt(s[1], 10, 64)
			}
		}
		q.Del(param)
	}
	parse("p", 0) // padding
	parse("c", 2) // concurrency
	parse("t", 4) // times
	parse("i", 6) // interval
	parse("r", 8) // return
	u.RawQuery = q.Encode()
	config.SpiderX = u.String()

	return &Engine{
		Config: config,
	}, nil
}

func (c *Engine) GetTLSConfig(_ ...security.Option) *gotls.Config {
	return &gotls.Config{
		ServerName: c.Config.ServerName,
	}
}

func (c *Engine) GetClientConn(conn net.Conn, opts ...security.Option) (net.Conn, error) {
	var dst *net.Destination
	for _, o := range opts {
		switch v := o.(type) {
		case security.OptionWithDestination:
			dst = &v.Dest
		}
	}
	if dst == nil {
		return nil, errors.New("REALITY: no destination provided")
	}
	return UClient(conn, c.Config, context.Background(), *dst)
}

func (c *Engine) Listener(l net.Listener) net.Listener {
	panic("implement me")
}

var (
	Version_x byte = 25
	Version_y byte = 5
	Version_z byte = 16
)

type Conn struct {
	*reality.Conn
}

func (c *Conn) HandshakeAddress() net.Address {
	if err := c.Handshake(); err != nil {
		return nil
	}
	state := c.ConnectionState()
	if state.ServerName == "" {
		return nil
	}
	return net.ParseAddress(state.ServerName)
}

func (c *Conn) GetConnectionApplicationProtocol() (string, error) {
	if err := c.Handshake(); err != nil {
		return "", err
	}
	return c.ConnectionState().NegotiatedProtocol, nil
}

func Server(c net.Conn, config *reality.Config) (*Conn, error) {
	realityConn, err := reality.Server(context.Background(), c, config)
	return &Conn{Conn: realityConn}, err
}

type UConn struct {
	*utls.UConn
	ServerName string
	// a shared secret that is used to encrypt
	AuthKey  []byte
	Verified bool
}

func (c *UConn) HandshakeAddress() net.Address {
	if err := c.Handshake(); err != nil {
		return nil
	}
	state := c.ConnectionState()
	if state.ServerName == "" {
		return nil
	}
	return net.ParseAddress(state.ServerName)
}

func (c *UConn) VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	p, _ := reflect.TypeOf(c.Conn).Elem().FieldByName("peerCertificates")
	certs := *(*([]*x509.Certificate))(unsafe.Pointer(uintptr(unsafe.Pointer(c.Conn)) + p.Offset))
	if pub, ok := certs[0].PublicKey.(ed25519.PublicKey); ok {
		h := hmac.New(sha512.New, c.AuthKey)
		h.Write(pub)
		if bytes.Equal(h.Sum(nil), certs[0].Signature) {
			c.Verified = true
			return nil
		}
	}
	// at this point, it is confirmed that certs are not issued by a reality server.
	// try to verify original certificate (e.g. from a website)
	log.Warn().Msg("get a certificate from a non-reality server")
	opts := x509.VerifyOptions{
		DNSName:       c.ServerName,
		Intermediates: x509.NewCertPool(),
	}
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}
	if _, err := certs[0].Verify(opts); err != nil {
		return err
	}
	return nil
}

func UClient(c net.Conn, config *RealityConfig, ctx context.Context, dest net.Destination) (net.Conn, error) {
	localAddr := c.LocalAddr().String()
	uConn := &UConn{}
	utlsConfig := &utls.Config{
		VerifyPeerCertificate:  uConn.VerifyPeerCertificate,
		ServerName:             config.ServerName,
		InsecureSkipVerify:     true,
		SessionTicketsDisabled: true,
		KeyLogWriter:           KeyLogWriterFromConfig(config),
	}
	if utlsConfig.ServerName == "" {
		utlsConfig.ServerName = dest.Address.String()
	}
	uConn.ServerName = utlsConfig.ServerName
	fingerprint, found := tls.GetFingerprint(config.Fingerprint)
	if !found {
		return nil, errors.New("REALITY: failed to get fingerprint: ", config.Fingerprint)
	}
	uConn.UConn = utls.UClient(c, utlsConfig, *fingerprint)
	{
		uConn.BuildHandshakeState()
		hello := uConn.HandshakeState.Hello
		hello.SessionId = make([]byte, 32)

		// why is this needed? hello.Raw is used as additional data might be the reason
		copy(hello.Raw[39:], hello.SessionId) // the fixed location of `Session ID`

		// copy auth info. 16 byte
		hello.SessionId[0] = Version_x
		hello.SessionId[1] = Version_y
		hello.SessionId[2] = Version_z
		hello.SessionId[3] = 0 // reserved
		binary.BigEndian.PutUint32(hello.SessionId[4:], uint32(time.Now().Unix()))
		copy(hello.SessionId[8:], config.ShortId)

		publicKey, err := ecdh.X25519().NewPublicKey(config.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("ecdh.X25519().NewPublicKey(config.PublicKey) err: %w", err)
		}
		if uConn.HandshakeState.State13.KeyShareKeys.Ecdhe == nil {
			return nil, errors.New("Current fingerprint ", uConn.ClientHelloID.Client, uConn.ClientHelloID.Version, " does not support TLS 1.3, REALITY handshake cannot establish.")
		}
		// EcdheKey is a private key. Server has this shared secret(uConn.AuthKey) too
		uConn.AuthKey, _ = uConn.HandshakeState.State13.KeyShareKeys.Ecdhe.ECDH(publicKey)
		if uConn.AuthKey == nil {
			return nil, errors.New("REALITY: SharedKey == nil")
		}

		// get aead
		if _, err := hkdf.New(sha256.New, uConn.AuthKey, hello.Random[:20], []byte("REALITY")).Read(uConn.AuthKey); err != nil {
			return nil, err
		}
		block, _ := aes.NewCipher(uConn.AuthKey)
		aead, _ := cipher.NewGCM(block)
		aead.Seal(hello.SessionId[:0], hello.Random[20:], hello.SessionId[:16], hello.Raw)
		copy(hello.Raw[39:], hello.SessionId)
	}
	if err := uConn.HandshakeContext(ctx); err != nil {
		return nil, err
	}
	log.Ctx(ctx).Debug().Str("localAddr", localAddr).Bool("verified", uConn.Verified).Msg("REALITY")

	if !uConn.Verified {
		go func() {
			client := &http.Client{
				Transport: &http2.Transport{
					DialTLSContext: func(ctx context.Context, network, addr string, cfg *gotls.Config) (net.Conn, error) {
						return uConn, nil
					},
				},
			}
			prefix := []byte("https://" + uConn.ServerName)
			maps.Lock()
			if maps.maps == nil {
				maps.maps = make(map[string]map[string]struct{})
			}
			paths := maps.maps[uConn.ServerName]
			if paths == nil {
				paths = make(map[string]struct{})
				paths[config.SpiderX] = struct{}{}
				maps.maps[uConn.ServerName] = paths
			}
			firstURL := string(prefix) + getPathLocked(paths)
			maps.Unlock()
			get := func(first bool) {
				var (
					req  *http.Request
					resp *http.Response
					err  error
					body []byte
				)
				if first {
					req, _ = http.NewRequest("GET", firstURL, nil)
				} else {
					maps.Lock()
					req, _ = http.NewRequest("GET", string(prefix)+getPathLocked(paths), nil)
					maps.Unlock()
				}
				if req == nil {
					return
				}
				req.Header.Set("User-Agent", fingerprint.Client) // TODO: User-Agent map
				times := 1
				if !first {
					times = int(crypto.RandBetween(config.SpiderY[4], config.SpiderY[5]))
				}
				for j := 0; j < times; j++ {
					if !first && j == 0 {
						req.Header.Set("Referer", firstURL)
					}
					req.AddCookie(&http.Cookie{Name: "padding", Value: strings.Repeat("0", int(crypto.RandBetween(config.SpiderY[0], config.SpiderY[1])))})
					if resp, err = client.Do(req); err != nil {
						break
					}
					defer resp.Body.Close()
					req.Header.Set("Referer", req.URL.String())
					if body, err = io.ReadAll(resp.Body); err != nil {
						break
					}
					maps.Lock()
					for _, m := range href.FindAllSubmatch(body, -1) {
						m[1] = bytes.TrimPrefix(m[1], prefix)
						if !bytes.Contains(m[1], dot) {
							paths[string(m[1])] = struct{}{}
						}
					}
					req.URL.Path = getPathLocked(paths)
					maps.Unlock()
					if !first {
						time.Sleep(time.Duration(crypto.RandBetween(config.SpiderY[6], config.SpiderY[7])) * time.Millisecond) // interval
					}
				}
			}
			get(true)
			concurrency := int(crypto.RandBetween(config.SpiderY[2], config.SpiderY[3]))
			for i := 0; i < concurrency; i++ {
				go get(false)
			}
			// Do not close the connection
		}()
		time.Sleep(time.Duration(crypto.RandBetween(config.SpiderY[8], config.SpiderY[9])) * time.Millisecond) // return
		return nil, errors.New("REALITY: processed invalid connection")
	}
	return uConn, nil
}

var (
	href = regexp.MustCompile(`href="([/h].*?)"`)
	dot  = []byte(".")
)

var maps struct {
	sync.Mutex
	maps map[string]map[string]struct{}
}

func randBetween(left int64, right int64) int64 {
	if left == right {
		return left
	}
	bigInt, _ := rand.Int(rand.Reader, big.NewInt(right-left))
	return left + bigInt.Int64()
}

func getPathLocked(paths map[string]struct{}) string {
	stopAt := int(crypto.RandBetween(0, int64(len(paths)-1)))
	i := 0
	for s := range paths {
		if i == stopAt {
			return s
		}
		i++
	}
	return "/"
}
