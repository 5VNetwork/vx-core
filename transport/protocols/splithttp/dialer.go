package splithttp

import (
	"context"
	gotls "crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/pipe"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/common/uuid"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"
	"github.com/5vnetwork/vx-core/transport/security/reality"
	"github.com/5vnetwork/vx-core/transport/security/tls"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
)

type XhttpDialer struct {
	config      *SplitHttpConfig
	sc          security.Engine
	dialer      i.DialerListener
	downAddress string
	downPort    int
	downConfig  *SplitHttpConfig
	downSc      security.Engine
	dnsServer   i.ECHResolver
}

func NewXhttpDialer(config *SplitHttpConfig, sc security.Engine,
	dialer i.DialerListener, dnsServer i.ECHResolver) (*XhttpDialer, error) {
	d := &XhttpDialer{
		config: config,
		sc:     sc,
		dialer: dialer,
	}
	d.downConfig = config.GetDownloadSettings().GetXhttpConfig()
	d.downPort = int(config.GetDownloadSettings().GetPort())
	d.downAddress = config.GetDownloadSettings().GetAddress()
	if d.config.GetXmux() == nil {
		d.config.Xmux = &XmuxConfig{}
	}
	d.config.Xmux.Init()
	if d.downConfig != nil {
		if d.downConfig.Xmux == nil {
			d.downConfig.Xmux = &XmuxConfig{}
		}
		d.downConfig.Xmux.Init()
	}

	var securityEngine security.Engine
	var err error
	switch sc := config.GetDownloadSettings().GetSecurity().(type) {
	case *DownConfig_Tls:
		securityEngine, err = tls.NewEngine(
			tls.EngineConfig{Config: sc.Tls, DnsServer: d.dnsServer})
		if err != nil {
			return nil, fmt.Errorf("failed to create tls engine: %w", err)
		}
	case *DownConfig_Reality:
		securityEngine, err = reality.NewEngine(sc.Reality)
		if err != nil {
			return nil, fmt.Errorf("failed to create reality engine: %w", err)
		}
	}
	d.downSc = securityEngine
	return d, nil
}

func getSecurityConfig(sc security.Engine, dst *net.Destination) (tlsConfig *gotls.Config, realityConfig *reality.RealityConfig) {
	switch sc := sc.(type) {
	case *tls.Engine:
		if dst == nil {
			tlsConfig = sc.GetTLSConfig()
		} else {
			tlsConfig = sc.GetTLSConfig(security.OptionWithDestination{Dest: *dst})
		}
	case *reality.Engine:
		realityConfig = sc.Config
	}
	return
}

func (x *XhttpDialer) Dial(ctx context.Context, dest net.Destination) (net.Conn, error) {
	tlsConfig, realityConfig := getSecurityConfig(x.sc, &dest)

	httpVersion := decideHTTPVersion(tlsConfig, realityConfig)
	if httpVersion == "3" {
		dest.Network = net.Network_UDP
	}

	config := x.config
	var requestURL url.URL

	if tlsConfig != nil || realityConfig != nil {
		requestURL.Scheme = "https"
	} else {
		requestURL.Scheme = "http"
	}
	log.Ctx(ctx).Debug().Msgf("requestURL.Scheme: %s", requestURL.Scheme)
	requestURL.Host = config.Host
	if requestURL.Host == "" && tlsConfig != nil {
		requestURL.Host = tlsConfig.ServerName
	}
	if requestURL.Host == "" && realityConfig != nil {
		requestURL.Host = realityConfig.ServerName
	}
	if requestURL.Host == "" {
		requestURL.Host = dest.Address.String()
	}
	log.Ctx(ctx).Debug().Msgf("requestURL.Host: %s", requestURL.Host)

	sessionIdUuid := uuid.New()
	requestURL.Path = config.GetNormalizedPath() + sessionIdUuid.String()
	requestURL.RawQuery = config.GetNormalizedQuery()

	httpClient, xmuxClient := getHTTPClient(ctx, dest, x.config, x.sc, x.dialer)

	mode := config.Mode
	if mode == "" || mode == "auto" {
		mode = "packet-up"
		if realityConfig != nil {
			mode = "stream-one"
			if config.DownloadSettings != nil {
				mode = "stream-up"
			}
		}
	}

	log.Ctx(ctx).Debug().Msgf("XHTTP is dialing to %s, mode %s, HTTP version %s, host %s", dest, mode, httpVersion, requestURL.Host)

	requestURL2 := requestURL
	httpClient2 := httpClient
	xmuxClient2 := xmuxClient
	if config.DownloadSettings != nil {
		globalDialerAccess.Lock()
		globalDialerAccess.Unlock()
		dest2 := net.Destination{
			Address: net.ParseAddress(x.downAddress),
			Port:    net.Port(x.downPort),
			Network: net.Network_TCP,
		} // just panic
		tlsConfig2, realityConfig2 := getSecurityConfig(x.downSc, &dest2)
		httpVersion2 := decideHTTPVersion(tlsConfig2, realityConfig2)
		if httpVersion2 == "3" {
			dest2.Network = net.Network_UDP
		}
		if tlsConfig2 != nil || realityConfig2 != nil {
			requestURL2.Scheme = "https"
		} else {
			requestURL2.Scheme = "http"
		}
		config2 := x.downConfig
		requestURL2.Host = config2.Host
		if requestURL2.Host == "" && tlsConfig2 != nil {
			requestURL2.Host = tlsConfig2.ServerName
		}
		if requestURL2.Host == "" && realityConfig2 != nil {
			requestURL2.Host = realityConfig2.ServerName
		}
		if requestURL2.Host == "" {
			requestURL2.Host = dest2.Address.String()
		}
		requestURL2.Path = config2.GetNormalizedPath() + sessionIdUuid.String()
		requestURL2.RawQuery = config2.GetNormalizedQuery()
		httpClient2, xmuxClient2 = getHTTPClient(ctx, dest2, x.downConfig, x.downSc, x.dialer)
	}

	if xmuxClient != nil {
		xmuxClient.OpenUsage.Add(1)
	}
	if xmuxClient2 != nil && xmuxClient2 != xmuxClient {
		xmuxClient2.OpenUsage.Add(1)
	}
	var closed atomic.Int32

	reader, writer := io.Pipe()
	conn := splitConn{
		writer: writer,
		onClose: func() {
			if closed.Add(1) > 1 {
				return
			}
			if xmuxClient != nil {
				xmuxClient.OpenUsage.Add(-1)
			}
			if xmuxClient2 != nil && xmuxClient2 != xmuxClient {
				xmuxClient2.OpenUsage.Add(-1)
			}
		},
		localAddr: &net.TCPAddr{
			IP:   net.AnyIP.IP(),
			Port: 0,
		},
		remoteAddr: &net.TCPAddr{
			IP:   net.AnyIP.IP(),
			Port: 0,
		},
	}

	var err error
	if mode == "stream-one" {
		requestURL.Path = config.GetNormalizedPath()
		if xmuxClient != nil {
			xmuxClient.LeftRequests.Add(-1)
		}
		conn.reader, conn.remoteAddr, conn.localAddr, err = httpClient.OpenStream(ctx, requestURL.String(), reader, false)
		if err != nil { // browser dialer only
			return nil, err
		}
		return &conn, nil
	} else { // stream-down
		if xmuxClient2 != nil {
			xmuxClient2.LeftRequests.Add(-1)
		}
		conn.reader, conn.remoteAddr, conn.localAddr, err = httpClient2.OpenStream(ctx, requestURL2.String(), nil, false)
		if err != nil { // browser dialer only
			return nil, err
		}
	}
	if mode == "stream-up" {
		if xmuxClient != nil {
			xmuxClient.LeftRequests.Add(-1)
		}
		_, _, _, err = httpClient.OpenStream(ctx, requestURL.String(), reader, true)
		if err != nil { // browser dialer only
			return nil, err
		}
		return &conn, nil
	}

	scMaxEachPostBytes := config.GetNormalizedScMaxEachPostBytes()
	scMinPostsIntervalMs := config.GetNormalizedScMinPostsIntervalMs()

	if scMaxEachPostBytes.From <= buf.Size {
		panic("`scMaxEachPostBytes` should be bigger than " + strconv.Itoa(buf.Size))
	}

	maxUploadSize := scMaxEachPostBytes.rand()
	// WithSizeLimit(0) will still allow single bytes to pass, and a lot of
	// code relies on this behavior. Subtract 1 so that together with
	// uploadWriter wrapper, exact size limits can be enforced
	// uploadPipeReader, uploadPipeWriter := pipe.New(pipe.WithSizeLimit(maxUploadSize - 1))
	pp := pipe.NewPipe(maxUploadSize-buf.Size, false)

	conn.writer = uploadWriter{
		pp,
		maxUploadSize,
	}

	go func() {
		var seq int64
		var lastWrite time.Time

		for {
			wroteRequest := done.New()

			ctx := httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
				WroteRequest: func(httptrace.WroteRequestInfo) {
					wroteRequest.Close()
				},
			})

			// this intentionally makes a shallow-copy of the struct so we
			// can reassign Path (potentially concurrently)
			url := requestURL
			url.Path += "/" + strconv.FormatInt(seq, 10)

			seq += 1

			if scMinPostsIntervalMs.From > 0 {
				time.Sleep(time.Duration(scMinPostsIntervalMs.rand())*time.Millisecond - time.Since(lastWrite))
			}

			// by offloading the uploads into a buffered pipe, multiple conn.Write
			// calls get automatically batched together into larger POST requests.
			// without batching, bandwidth is extremely limited.
			chunk, err := pp.ReadMultiBuffer()
			if err != nil {
				break
			}

			lastWrite = time.Now()

			if xmuxClient != nil && (xmuxClient.LeftRequests.Add(-1) <= 0 ||
				(xmuxClient.UnreusableAt != time.Time{} && lastWrite.After(xmuxClient.UnreusableAt))) {
				httpClient, xmuxClient = getHTTPClient(ctx, dest, x.config, x.sc, x.dialer)
			}

			go func() {
				err := httpClient.PostPacket(
					ctx,
					url.String(),
					&buf.MultiBufferContainer{MultiBuffer: chunk},
					int64(chunk.Len()),
				)
				wroteRequest.Close()
				if err != nil {
					// errors.LogInfoInner(ctx, err, "failed to send upload")
					pp.Interrupt(err)
				}
			}()

			if _, ok := httpClient.(*DefaultDialerClient); ok {
				<-wroteRequest.Wait()
			}
		}
	}()

	return &conn, nil
}

// A wrapper around pipe that ensures the size limit is exactly honored.
//
// The MultiBuffer pipe accepts any single WriteMultiBuffer call even if that
// single MultiBuffer exceeds the size limit, and then starts blocking on the
// next WriteMultiBuffer call. This means that ReadMultiBuffer can return more
// bytes than the size limit. We work around this by splitting a potentially
// too large write up into multiple.
type uploadWriter struct {
	*pipe.Pipe
	maxLen int32
}

func (w uploadWriter) Write(b []byte) (int, error) {
	/*
		capacity := int(w.maxLen - w.Len())
		if capacity > 0 && capacity < len(b) {
			b = b[:capacity]
		}
	*/

	buffer := buf.New()
	n, err := buffer.Write(b)
	if err != nil {
		return 0, err
	}

	err = w.WriteMultiBuffer(buf.MultiBuffer{buffer})
	if err != nil {
		return 0, err
	}
	return n, nil
}

type dialerConf struct {
	net.Destination
	config       *SplitHttpConfig
	sc           security.Engine
	socketConfig i.DialerListener
}

var (
	globalDialerMap    map[dialerConf]*XmuxManager
	globalDialerAccess sync.Mutex
)

func getHTTPClient(ctx context.Context, dest net.Destination, config *SplitHttpConfig, sc security.Engine, socketConfig i.DialerListener) (DialerClient, *XmuxClient) {
	globalDialerAccess.Lock()
	defer globalDialerAccess.Unlock()

	if globalDialerMap == nil {
		globalDialerMap = make(map[dialerConf]*XmuxManager)
	}

	key := dialerConf{dest, config, sc, socketConfig}

	xmuxManager, found := globalDialerMap[key]

	if !found {
		transportConfig := config
		var xmuxConfig XmuxConfig
		if transportConfig.Xmux != nil {
			xmuxConfig = *transportConfig.Xmux
		}

		xmuxManager = NewXmuxManager(xmuxConfig, func() XmuxConn {
			return createHTTPClient(dest, config, sc, socketConfig)
		})
		globalDialerMap[key] = xmuxManager
	}

	xmuxClient := xmuxManager.GetXmuxClient(ctx)
	return xmuxClient.XmuxConn.(DialerClient), xmuxClient
}

func decideHTTPVersion(tlsConfig *gotls.Config, realityConfig *reality.RealityConfig) string {
	if realityConfig != nil {
		return "2"
	}
	if tlsConfig == nil {
		return "1.1"
	}
	if len(tlsConfig.NextProtos) != 1 {
		return "2"
	}
	if tlsConfig.NextProtos[0] == "http/1.1" {
		return "1.1"
	}
	if tlsConfig.NextProtos[0] == "h3" {
		return "3"
	}
	return "2"
}

// consistent with quic-go
const QuicgoH3KeepAlivePeriod = 10 * time.Second

// consistent with chrome
const ChromeH2KeepAlivePeriod = 45 * time.Second
const ConnIdleTimeout = 300 * time.Second

type FakePacketConn struct {
	net.Conn
}

func (c *FakePacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = c.Read(p)
	return n, c.RemoteAddr(), err
}

func (c *FakePacketConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.Write(p)
}

func (c *FakePacketConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IP{byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256))},
		Port: rand.Intn(65536),
	}
}

func (c *FakePacketConn) SetReadBuffer(bytes int) error {
	// do nothing, this function is only there to suppress quic-go printing
	// random warnings about UDP buffers to stdout
	return nil
}

func createHTTPClient(dest net.Destination, config *SplitHttpConfig, sc security.Engine, socketConfig i.DialerListener) DialerClient {
	tlsConfig, realityConfig := getSecurityConfig(sc, &dest)

	httpVersion := decideHTTPVersion(tlsConfig, realityConfig)
	if httpVersion == "3" {
		dest.Network = net.Network_UDP // better to keep this line
	}

	var gotlsConfig *gotls.Config

	if tlsConfig != nil {
		gotlsConfig = tlsConfig
	}

	transportConfig := config

	dialContext := func(ctxInner context.Context) (net.Conn, error) {
		conn, err := socketConfig.Dial(ctxInner, dest)
		if err != nil {
			return nil, err
		}
		if sc != nil {
			return sc.GetClientConn(conn, security.OptionWithDestination{Dest: dest})
		}
		return conn, nil
	}

	var keepAlivePeriod time.Duration
	if config.Xmux != nil {
		keepAlivePeriod = time.Duration(config.Xmux.HKeepAlivePeriod) * time.Second
	}

	var transport http.RoundTripper

	if httpVersion == "3" {
		if keepAlivePeriod == 0 {
			keepAlivePeriod = QuicgoH3KeepAlivePeriod
		}
		if keepAlivePeriod < 0 {
			keepAlivePeriod = 0
		}
		quicConfig := &quic.Config{
			MaxIdleTimeout: ConnIdleTimeout,

			// these two are defaults of quic-go/http3. the default of quic-go (no
			// http3) is different, so it is hardcoded here for clarity.
			// https://github.com/quic-go/quic-go/blob/b8ea5c798155950fb5bbfdd06cad1939c9355878/http3/client.go#L36-L39
			MaxIncomingStreams: -1,
			KeepAlivePeriod:    keepAlivePeriod,
		}
		transport = &http3.Transport{
			QUICConfig:      quicConfig,
			TLSClientConfig: gotlsConfig,
			Dial: func(ctx context.Context, addr string, tlsCfg *gotls.Config, cfg *quic.Config) (*quic.Conn, error) {
				udpDest := net.Destination{
					Address: dest.Address,
					Port:    dest.Port,
					Network: net.Network_UDP,
				}
				conn, err := socketConfig.Dial(ctx, udpDest)
				if err != nil {
					return nil, err
				}
				ret, err := quic.DialEarly(ctx, &net.NetConnToPacketConn{
					Conn: conn,
				}, conn.RemoteAddr(), tlsCfg, cfg)
				if err != nil {
					return nil, err
				}
				return ret, nil
			},
		}
	} else if httpVersion == "2" {
		if keepAlivePeriod == 0 {
			keepAlivePeriod = ChromeH2KeepAlivePeriod
		}
		if keepAlivePeriod < 0 {
			keepAlivePeriod = 0
		}
		transport = &http2.Transport{
			DialTLSContext: func(ctxInner context.Context, network string, addr string, cfg *gotls.Config) (net.Conn, error) {
				return dialContext(ctxInner)
			},
			IdleConnTimeout: ConnIdleTimeout,
			ReadIdleTimeout: keepAlivePeriod,
		}
	} else {
		httpDialContext := func(ctxInner context.Context, network string, addr string) (net.Conn, error) {
			return dialContext(ctxInner)
		}

		transport = &http.Transport{
			DialTLSContext:  httpDialContext,
			DialContext:     httpDialContext,
			IdleConnTimeout: ConnIdleTimeout,
			// chunked transfer download with KeepAlives is buggy with
			// http.Client and our custom dial context.
			DisableKeepAlives: true,
		}
	}

	client := &DefaultDialerClient{
		transportConfig: transportConfig,
		client: &http.Client{
			Transport: transport,
		},
		httpVersion:    httpVersion,
		uploadRawPool:  &sync.Pool{},
		dialUploadConn: dialContext,
	}

	return client
}
