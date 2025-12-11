package splithttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/net"
	http_proto "github.com/5vnetwork/vx-core/common/protocol/http"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/rs/zerolog/log"
)

type requestHandler struct {
	config    *SplitHttpConfig
	host      string
	path      string
	ln        *Listener
	sessionMu *sync.Mutex
	sessions  sync.Map
	localAddr net.Addr
}

type httpSession struct {
	uploadQueue *uploadQueue
	// for as long as the GET request is not opened by the client, this will be
	// open ("undone"), and the session may be expired within a certain TTL.
	// after the client connects, this becomes "done" and the session lives as
	// long as the GET request.
	isFullyConnected *done.Instance
}

func (h *requestHandler) upsertSession(sessionId string) *httpSession {
	// fast path
	currentSessionAny, ok := h.sessions.Load(sessionId)
	if ok {
		return currentSessionAny.(*httpSession)
	}

	// slow path
	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()

	currentSessionAny, ok = h.sessions.Load(sessionId)
	if ok {
		return currentSessionAny.(*httpSession)
	}

	s := &httpSession{
		uploadQueue:      NewUploadQueue(h.ln.config.GetNormalizedScMaxBufferedPosts()),
		isFullyConnected: done.New(),
	}

	h.sessions.Store(sessionId, s)

	shouldReap := done.New()
	go func() {
		time.Sleep(30 * time.Second)
		shouldReap.Close()
	}()
	go func() {
		select {
		case <-shouldReap.Wait():
			h.sessions.Delete(sessionId)
			s.uploadQueue.Close()
		case <-s.isFullyConnected.Wait():
		}
	}()

	return s
}

func IsValidHTTPHost(request string, config string) bool {
	r := strings.ToLower(request)
	c := strings.ToLower(config)
	if strings.Contains(r, ":") {
		h, _, _ := net.SplitHostPort(r)
		return h == c
	}
	return r == c
}

func (h *requestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if len(h.host) > 0 && !IsValidHTTPHost(request.Host, h.host) {
		log.Info().Msgf("failed to validate host, request: %s, config: %s", request.Host, h.host)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if !strings.HasPrefix(request.URL.Path, h.path) {
		log.Info().Msgf("failed to validate path, request: %s, config: %s", request.URL.Path, h.path)
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	h.config.WriteResponseHeader(writer)

	/*
		clientVer := []int{0, 0, 0}
		x_version := strings.Split(request.URL.Query().Get("x_version"), ".")
		for j := 0; j < 3 && len(x_version) > j; j++ {
			clientVer[j], _ = strconv.Atoi(x_version[j])
		}
	*/

	validRange := h.config.GetNormalizedXPaddingBytes()
	paddingLength := 0

	referrer := request.Header.Get("Referer")
	if referrer != "" {
		if referrerURL, err := url.Parse(referrer); err == nil {
			// Browser dialer cannot control the host part of referrer header, so only check the query
			paddingLength = len(referrerURL.Query().Get("x_padding"))
		}
	} else {
		paddingLength = len(request.URL.Query().Get("x_padding"))
	}

	if int32(paddingLength) < validRange.From || int32(paddingLength) > validRange.To {
		log.Info().Msgf("invalid x_padding length: %d", int32(paddingLength))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionId := ""
	subpath := strings.Split(request.URL.Path[len(h.path):], "/")
	if len(subpath) > 0 {
		sessionId = subpath[0]
	}

	if sessionId == "" && h.config.Mode != "" && h.config.Mode != "auto" && h.config.Mode != "stream-one" && h.config.Mode != "stream-up" {
		log.Info().Msg("stream-one mode is not allowed")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	forwardedAddrs := http_proto.ParseXForwardedFor(request.Header)
	var remoteAddr net.Addr
	var err error
	remoteAddr, err = net.ResolveTCPAddr("tcp", request.RemoteAddr)
	if err != nil {
		remoteAddr = &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		}
	}
	if request.ProtoMajor == 3 {
		remoteAddr = &net.UDPAddr{
			IP:   remoteAddr.(*net.TCPAddr).IP,
			Port: remoteAddr.(*net.TCPAddr).Port,
		}
	}
	if len(forwardedAddrs) > 0 && forwardedAddrs[0].Family().IsIP() {
		remoteAddr = &net.TCPAddr{
			IP:   forwardedAddrs[0].IP(),
			Port: 0,
		}
	}

	var currentSession *httpSession
	if sessionId != "" {
		currentSession = h.upsertSession(sessionId)
	}
	scMaxEachPostBytes := int(h.ln.config.GetNormalizedScMaxEachPostBytes().To)

	if request.Method == "POST" && sessionId != "" { // stream-up, packet-up
		seq := ""
		if len(subpath) > 1 {
			seq = subpath[1]
		}

		if seq == "" {
			if h.config.Mode != "" && h.config.Mode != "auto" && h.config.Mode != "stream-up" {
				log.Info().Msg("stream-up mode is not allowed")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			httpSC := &httpServerConn{
				Instance:       done.New(),
				Reader:         request.Body,
				ResponseWriter: writer,
			}
			err = currentSession.uploadQueue.Push(Packet{
				Reader: httpSC,
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to upload (PushReader)")
				writer.WriteHeader(http.StatusConflict)
			} else {
				writer.Header().Set("X-Accel-Buffering", "no")
				writer.Header().Set("Cache-Control", "no-store")
				writer.WriteHeader(http.StatusOK)
				scStreamUpServerSecs := h.config.GetNormalizedScStreamUpServerSecs()
				if referrer != "" && scStreamUpServerSecs.To > 0 {
					go func() {
						for {
							_, err := httpSC.Write(bytes.Repeat([]byte{'X'}, int(h.config.GetNormalizedXPaddingBytes().rand())))
							if err != nil {
								break
							}
							time.Sleep(time.Duration(scStreamUpServerSecs.rand()) * time.Second)
						}
					}()
				}
				select {
				case <-request.Context().Done():
				case <-httpSC.Wait():
				}
			}
			httpSC.Close()
			return
		}

		if h.config.Mode != "" && h.config.Mode != "auto" && h.config.Mode != "packet-up" {
			log.Info().Msg("packet-up mode is not allowed")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		payload, err := io.ReadAll(io.LimitReader(request.Body, int64(scMaxEachPostBytes)+1))

		if len(payload) > scMaxEachPostBytes {
			log.Info().Msgf("Too large upload. scMaxEachPostBytes is set to %d, but request size exceed it. Adjust scMaxEachPostBytes on the server to be at least as large as client.", scMaxEachPostBytes)
			writer.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		if err != nil {
			log.Error().Err(err).Msg("failed to upload (ReadAll)")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		seqInt, err := strconv.ParseUint(seq, 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("failed to upload (ParseUint)")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = currentSession.uploadQueue.Push(Packet{
			Payload: payload,
			Seq:     seqInt,
		})

		if err != nil {
			log.Error().Err(err).Msg("failed to upload (PushPayload)")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	} else if request.Method == "GET" || sessionId == "" { // stream-down, stream-one
		if sessionId != "" {
			// after GET is done, the connection is finished. disable automatic
			// session reaping, and handle it in defer
			currentSession.isFullyConnected.Close()
			defer h.sessions.Delete(sessionId)
		}

		// magic header instructs nginx + apache to not buffer response body
		writer.Header().Set("X-Accel-Buffering", "no")
		// A web-compliant header telling all middleboxes to disable caching.
		// Should be able to prevent overloading the cache, or stop CDNs from
		// teeing the response stream into their cache, causing slowdowns.
		writer.Header().Set("Cache-Control", "no-store")

		if !h.config.NoSSEHeader {
			// magic header to make the HTTP middle box consider this as SSE to disable buffer
			writer.Header().Set("Content-Type", "text/event-stream")
		}

		writer.WriteHeader(http.StatusOK)
		writer.(http.Flusher).Flush()

		httpSC := &httpServerConn{
			Instance:       done.New(),
			Reader:         request.Body,
			ResponseWriter: writer,
		}
		conn := splitConn{
			writer:     httpSC,
			reader:     httpSC,
			remoteAddr: remoteAddr,
			localAddr:  h.localAddr,
		}
		if sessionId != "" { // if not stream-one
			conn.reader = currentSession.uploadQueue
		}

		h.ln.addConn(&conn)

		// "A ResponseWriter may not be used after [Handler.ServeHTTP] has returned."
		select {
		case <-request.Context().Done():
		case <-httpSC.Wait():
		}

		conn.Close()
	} else {
		log.Info().Msgf("unsupported method: %s", request.Method)
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type httpServerConn struct {
	sync.Mutex
	*done.Instance
	io.Reader // no need to Close request.Body
	http.ResponseWriter
}

func (c *httpServerConn) Write(b []byte) (int, error) {
	c.Lock()
	defer c.Unlock()
	if c.Done() {
		return 0, io.ErrClosedPipe
	}
	n, err := c.ResponseWriter.Write(b)
	if err == nil {
		c.ResponseWriter.(http.Flusher).Flush()
	}
	return n, err
}

func (c *httpServerConn) Close() error {
	c.Lock()
	defer c.Unlock()
	return c.Instance.Close()
}

type Listener struct {
	sync.Mutex
	server     http.Server
	h3server   *http3.Server
	listener   net.Listener
	h3listener *quic.EarlyListener
	config     *SplitHttpConfig
	addConn    func(net.Conn)
	isH3       bool
}

func ListenXH(ctx context.Context, addr net.Destination,
	config *SplitHttpConfig, li i.Listener,
	addConn func(net.Conn)) (*Listener, error) {
	l := &Listener{
		addConn: addConn,
	}
	l.config = config
	handler := &requestHandler{
		config:    l.config,
		host:      l.config.Host,
		path:      l.config.GetNormalizedPath(),
		ln:        l,
		sessionMu: &sync.Mutex{},
		sessions:  sync.Map{},
	}

	listener, err := li.Listen(ctx, addr.Addr())
	if err != nil {
		return nil, err
	}
	l.listener = listener

	// tcp/unix (h1/h2)
	if l.listener != nil {
		handler.localAddr = l.listener.Addr()
		// server can handle both plaintext HTTP/1.1 and h2c
		protocols := new(http.Protocols)
		protocols.SetHTTP1(true)
		protocols.SetUnencryptedHTTP2(true)
		l.server = http.Server{
			Handler:           handler,
			ReadHeaderTimeout: time.Second * 4,
			MaxHeaderBytes:    8192,
			Protocols:         protocols,
		}
		go func() {
			if err := l.server.Serve(l.listener); err != nil {
				log.Error().Err(err).Msg("failed to serve HTTP for XHTTP")
			}
		}()
	}

	return l, err
}

// Addr implements net.Listener.Addr().
func (ln *Listener) Addr() net.Addr {
	if ln.h3listener != nil {
		return ln.h3listener.Addr()
	}
	if ln.listener != nil {
		return ln.listener.Addr()
	}
	return nil
}

// Close implements net.Listener.Close().
func (ln *Listener) Close() error {
	if ln.h3server != nil {
		if err := ln.h3server.Close(); err != nil {
			return err
		}
	} else if ln.listener != nil {
		return ln.listener.Close()
	}
	return errors.New("listener does not have an HTTP/3 server or a net.listener")
}
