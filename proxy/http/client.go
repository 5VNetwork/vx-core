package http

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/http2"

	proxyconfig "github.com/5vnetwork/vx-core/app/configs/proxy"
	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/5vnetwork/vx-core/common/bytespool"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/net/udp"
	"github.com/5vnetwork/vx-core/common/retry"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/proxy"
	"github.com/5vnetwork/vx-core/proxy/helper"
	"github.com/5vnetwork/vx-core/transport/security"
	"github.com/rs/zerolog/log"
)

type Client struct {
	ClientSettings
}

type h2Conn struct {
	rawConn net.Conn
	h2Conn  *http2.ClientConn
}

var (
	cachedH2Mutex sync.Mutex
	cachedH2Conns map[net.Destination]h2Conn
)

type ClientSettings struct {
	Address            net.Address
	PortPicker         i.PortSelector
	Account            *proxyconfig.Account
	H1SkipWaitForReply bool
	Dialer             i.Dialer
}

// NewClient create a new http client based on the given config.
func NewClient(settings ClientSettings) *Client {
	return &Client{
		ClientSettings: settings,
	}
}

func (c *Client) HandlePacketConn(ctx context.Context, dst net.Destination, pc udp.PacketReaderWriter) error {
	return errors.New("http client cannot handler udp")
}

// Process implements proxy.Outbound.Process. We first create a socket tunnel via HTTP CONNECT method, then redirect all inbound traffic to that tunnel.
func (c *Client) HandleFlow(ctx context.Context, dst net.Destination, rw buf.ReaderWriter) error {
	target := dst
	targetAddr := target.NetAddr()
	dialer := c.Dialer

	if target.Network == net.Network_UDP {
		return proxy.ErrUDPNotSupport
	}

	var user *proxyconfig.Account
	var conn net.Conn

	var firstPayload []byte

	if reader, ok := rw.(buf.TimeoutReader); ok {
		// 0-RTT optimization for HTTP/2: If the payload comes very soon, it can be
		// transmitted together. Note we should not get stuck here, as the payload may
		// not exist (considering to access MySQL database via a HTTP proxy, where the
		// server sends hello to the client first).
		waitTime := proxy.FirstPayloadTimeout
		if c.H1SkipWaitForReply {
			// Some server require first write to be present in client hello.
			// Increase timeout to if the client have explicitly requested to skip waiting for reply.
			waitTime = time.Second
		}
		if mbuf, _ := reader.ReadMultiBufferTimeout(waitTime); mbuf != nil {
			mlen := mbuf.Len()
			b := bytespool.Alloc(mlen)
			defer bytespool.Free(b)
			mbuf, _ = buf.SplitBytes(mbuf, b)
			firstPayload = b[:mlen]
			buf.ReleaseMulti(mbuf)
		}
	}

	if err := retry.ExponentialBackoff(5, 100).On(func() error {
		port := c.PortPicker.SelectPort()
		dest := net.TCPDestination(c.Address, net.Port(port))
		user = c.Account

		netConn, firstResp, err := setUpHTTPTunnel(ctx, dest, targetAddr, user, dialer, firstPayload, c.H1SkipWaitForReply)
		if netConn != nil {
			if _, ok := netConn.(*http2Conn); !ok && !c.H1SkipWaitForReply {
				if _, err := netConn.Write(firstPayload); err != nil {
					netConn.Close()
					return err
				}
			}
			if firstResp != nil {
				if err := rw.WriteMultiBuffer(firstResp); err != nil {
					return err
				}
			}
			conn = netConn
		}
		return err
	}); err != nil {
		return errors.New("failed to find an available destination").Base(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			errors.New("failed to closed connection").Base(err)
		}
	}()

	return helper.Relay(ctx, rw, rw, buf.NewReader(conn), buf.NewWriter(conn))
}

func (c *Client) ProxyDial(ctx context.Context, dst net.Destination, initialData buf.MultiBuffer) (i.FlowConn, error) {
	target := dst
	targetAddr := target.NetAddr()
	dialer := c.Dialer

	if target.Network == net.Network_UDP {
		return nil, errors.New("UDP is not supported by HTTP outbound")
	}

	var user *proxyconfig.Account
	var conn net.Conn

	if err := retry.ExponentialBackoff(5, 100).On(func() error {
		port := c.PortPicker.SelectPort()
		dest := net.TCPDestination(c.Address, net.Port(port))
		user = c.Account

		mlen := initialData.Len()
		b := bytespool.Alloc(mlen)
		defer bytespool.Free(b)
		buf.SplitBytes(initialData, b)
		firstPayload := b[:mlen]

		netConn, firstResp, err := setUpHTTPTunnel(ctx, dest, targetAddr, user, dialer, firstPayload, c.H1SkipWaitForReply)
		if netConn != nil {
			if _, ok := netConn.(*http2Conn); !ok && !c.H1SkipWaitForReply {
				if _, err := netConn.Write(firstPayload); err != nil {
					netConn.Close()
					return err
				}
			}
			if firstResp != nil {
				netConn = net.NewMbConn(netConn, firstResp)
			}
			conn = netConn
		}
		return err
	}); err != nil {
		return nil, errors.New("failed to find an available destination").Base(err)
	}

	return proxy.NewFlowConn(proxy.FlowConnOption{
		Reader:      buf.NewReader(conn),
		Writer:      buf.NewWriter(conn),
		SetDeadline: conn,
		Close:       conn.Close,
	}), nil
}

func (c *Client) ListenPacket(ctx context.Context, dst net.Destination) (udp.UdpConn, error) {
	return nil, proxy.ErrUDPNotSupport
}

// setUpHTTPTunnel will create a socket tunnel via HTTP CONNECT method
func setUpHTTPTunnel(ctx context.Context, dest net.Destination, target string, user *proxyconfig.Account, dialer i.Dialer, firstPayload []byte, writeFirstPayloadInH1 bool,
) (net.Conn, buf.MultiBuffer, error) {
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: target},
		Header: make(http.Header),
		Host:   target,
	}

	if user != nil {
		account := user
		auth := account.GetUsername() + ":" + account.GetPassword()
		req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}

	connectHTTP1 := func(rawConn net.Conn) (net.Conn, buf.MultiBuffer, error) {
		req.Header.Set("Proxy-Connection", "Keep-Alive")

		if !writeFirstPayloadInH1 {
			err := req.Write(rawConn)
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
		} else {
			buffer := bytes.NewBuffer(nil)
			err := req.Write(buffer)
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
			_, err = io.Copy(buffer, bytes.NewReader(firstPayload))
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
			_, err = rawConn.Write(buffer.Bytes())
			if err != nil {
				rawConn.Close()
				return nil, nil, err
			}
		}
		bufferedReader := bufio.NewReader(rawConn)
		resp, err := http.ReadResponse(bufferedReader, req)
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			rawConn.Close()
			return nil, nil, errors.New("Proxy responded with non 200 code: " + resp.Status)
		}
		if bufferedReader.Buffered() > 0 {
			payload, err := buf.ReadFrom(io.LimitReader(bufferedReader, int64(bufferedReader.Buffered())))
			if err != nil {
				return nil, nil, errors.New("unable to drain buffer: ").Base(err)
			}
			return rawConn, payload, nil
		}
		return rawConn, nil, nil
	}

	connectHTTP2 := func(rawConn net.Conn, h2clientConn *http2.ClientConn) (net.Conn, error) {
		pr, pw := io.Pipe()
		req.Body = pr

		var pErr error
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			_, pErr = pw.Write(firstPayload)
			wg.Done()
		}()

		resp, err := h2clientConn.RoundTrip(req) // nolint: bodyclose
		if err != nil {
			rawConn.Close()
			return nil, err
		}

		wg.Wait()
		if pErr != nil {
			rawConn.Close()
			return nil, pErr
		}

		if resp.StatusCode != http.StatusOK {
			rawConn.Close()
			return nil, errors.New("Proxy responded with non 200 code: " + resp.Status)
		}
		return newHTTP2Conn(rawConn, pw, resp.Body), nil
	}

	cachedH2Mutex.Lock()
	cachedConn, cachedConnFound := cachedH2Conns[dest]
	cachedH2Mutex.Unlock()

	if cachedConnFound {
		rc, cc := cachedConn.rawConn, cachedConn.h2Conn
		if cc.CanTakeNewRequest() {
			proxyConn, err := connectHTTP2(rc, cc)
			if err != nil {
				return nil, nil, err
			}

			return proxyConn, nil, nil
		}
	}

	rawConn, err := dialer.Dial(ctx, dest)
	if err != nil {
		return nil, nil, err
	}
	log.Ctx(ctx).Debug().Str("laddr", rawConn.LocalAddr().String()).Msg("http client dial ok")
	iConn := rawConn

	nextProto := ""
	if connALPNGetter, ok := iConn.(security.ConnectionApplicationProtocol); ok {
		nextProto, err = connALPNGetter.GetConnectionApplicationProtocol()
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}
	}

	switch nextProto {
	case "", "http/1.1":
		return connectHTTP1(rawConn)
	case "h2":
		t := http2.Transport{}
		h2clientConn, err := t.NewClientConn(rawConn)
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}

		proxyConn, err := connectHTTP2(rawConn, h2clientConn)
		if err != nil {
			rawConn.Close()
			return nil, nil, err
		}

		cachedH2Mutex.Lock()
		if cachedH2Conns == nil {
			cachedH2Conns = make(map[net.Destination]h2Conn)
		}

		cachedH2Conns[dest] = h2Conn{
			rawConn: rawConn,
			h2Conn:  h2clientConn,
		}
		cachedH2Mutex.Unlock()

		return proxyConn, nil, err
	default:
		return nil, nil, errors.New("negotiated unsupported application layer protocol: " + nextProto)
	}
}

func newHTTP2Conn(c net.Conn, pipedReqBody *io.PipeWriter, respBody io.ReadCloser) net.Conn {
	return &http2Conn{Conn: c, in: pipedReqBody, out: respBody}
}

type http2Conn struct {
	net.Conn
	in  *io.PipeWriter
	out io.ReadCloser
}

func (h *http2Conn) Read(p []byte) (n int, err error) {
	return h.out.Read(p)
}

func (h *http2Conn) Write(p []byte) (n int, err error) {
	return h.in.Write(p)
}

func (h *http2Conn) Close() error {
	h.in.Close()
	return h.out.Close()
}
