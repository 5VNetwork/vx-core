package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	http_proto "github.com/5vnetwork/vx-core/common/protocol/http"
	"github.com/5vnetwork/vx-core/common/signal/done"
)

type Listener struct {
	server *http.Server
	local  net.Addr
	config *HttpConfig
	h      func(net.Conn)
}

func (l *Listener) Addr() net.Addr {
	return l.local
}

func (l *Listener) Close() error {
	return l.server.Close()
}

type flushWriter struct {
	w io.Writer
	d *done.Instance
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	if fw.d.Done() {
		return 0, io.ErrClosedPipe
	}

	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func (l *Listener) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	if len(l.config.Host) != 0 && !l.config.isValidHost(host) {
		writer.WriteHeader(404)
		return
	}
	path := l.config.getNormalizedPath()
	if !strings.HasPrefix(request.URL.Path, path) {
		writer.WriteHeader(404)
		return
	}

	writer.Header().Set("Cache-Control", "no-store")

	for _, httpHeader := range l.config.Header {
		for _, httpHeaderValue := range httpHeader.Value {
			writer.Header().Set(httpHeader.Name, httpHeaderValue)
		}
	}

	writer.WriteHeader(200)
	if f, ok := writer.(http.Flusher); ok {
		f.Flush()
	}

	remoteAddr := l.Addr()
	dest, err := net.ParseDestination(request.RemoteAddr)
	if err != nil {
		log.Err(err).Msg("failed to parse request remote addr" + request.RemoteAddr)
	} else {
		remoteAddr = &net.TCPAddr{
			IP:   dest.Address.IP(),
			Port: int(dest.Port),
		}
	}

	forwardedAddress := http_proto.ParseXForwardedFor(request.Header)
	if len(forwardedAddress) > 0 && forwardedAddress[0].Family().IsIP() {
		remoteAddr = &net.TCPAddr{
			IP:   forwardedAddress[0].IP(),
			Port: 0,
		}
	}

	done := done.New()
	conn := net.NewConnection(
		net.ConnectionOutput(request.Body),
		net.ConnectionInput(flushWriter{w: writer, d: done}),
		net.ConnectionOnClose(common.ChainedClosable{done, request.Body}),
		net.ConnectionLocalAddr(l.Addr()),
		net.ConnectionRemoteAddr(remoteAddr),
	)
	l.h(conn)
	<-done.Wait()
}

func Listen(ctx context.Context, addr net.Destination, c *HttpConfig,
	li i.Listener, h func(net.Conn)) (*Listener, error) {
	listener := &Listener{
		local:  addr.Addr(),
		config: c,
		h:      h,
	}

	var server *http.Server
	h2s := &http2.Server{}
	server = &http.Server{
		Addr:              addr.NetAddr(),
		Handler:           h2c.NewHandler(listener, h2s),
		ReadHeaderTimeout: time.Second * 4,
	}

	listener.server = server
	go func() {
		streamListener, err := li.Listen(ctx, addr.Addr())
		if err != nil {
			log.Err(err).Msg("failed to listen on " + addr.NetAddr())
			return
		}
		err = server.Serve(streamListener)
		if err != nil {
			log.Err(err).Msg("stop to serve")
		}
	}()

	return listener, nil
}
