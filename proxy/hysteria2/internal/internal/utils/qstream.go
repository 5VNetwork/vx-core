package utils

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/apernet/quic-go"
)

// QStream is a wrapper of quic.Stream that handles Close() in a way that
// makes more sense to us. By default, quic.Stream's Close() only closes
// the write side of the stream, not the read side. And if there is unread
// data, the stream is not really considered closed until either the data
// is drained or CancelRead() is called.
// References:
// - https://github.com/libp2p/go-libp2p/blob/master/p2p/transport/quic/stream.go
// - https://github.com/quic-go/quic-go/issues/3558
// - https://github.com/quic-go/quic-go/issues/1599
type QStream struct {
	Stream *quic.Stream
}

func (s *QStream) StreamID() quic.StreamID {
	return s.Stream.StreamID()
}

func (s *QStream) Read(p []byte) (n int, err error) {
	n, err = s.Stream.Read(p)
	return n, normalizeStreamError(err)
}

func (s *QStream) CancelRead(code quic.StreamErrorCode) {
	s.Stream.CancelRead(code)
}

func (s *QStream) SetReadDeadline(t time.Time) error {
	return s.Stream.SetReadDeadline(t)
}

func (s *QStream) Write(p []byte) (n int, err error) {
	return s.Stream.Write(p)
}

// CloseWrite half-closes the write side (QUIC STREAM FIN) without cancelling
// reads. Relay() calls this when the destination TCP connection EOFs, so the
// client can still receive the HTTP response after finishing an upload.
func (s *QStream) CloseWrite() error {
	return s.Stream.Close()
}

func (s *QStream) Close() error {
	s.Stream.CancelRead(0)
	return s.Stream.Close()
}

func (s *QStream) CancelWrite(code quic.StreamErrorCode) {
	s.Stream.CancelWrite(code)
}

func (s *QStream) Context() context.Context {
	return s.Stream.Context()
}

func (s *QStream) SetWriteDeadline(t time.Time) error {
	return s.Stream.SetWriteDeadline(t)
}

func (s *QStream) SetDeadline(t time.Time) error {
	return s.Stream.SetDeadline(t)
}

// normalizeStreamError maps a QUIC stream cancellation with application error
// code 0 (the "no error" / normal close code, e.g. from our own Close() ->
// CancelRead(0), or a peer closing cleanly) to io.EOF. This keeps benign stream
// teardown from surfacing as a scary "stream N canceled by local with error
// code 0" error to consumers, while still propagating real (non-zero) errors.
func normalizeStreamError(err error) error {
	if err == nil {
		return nil
	}
	var se *quic.StreamError
	if errors.As(err, &se) && se.ErrorCode == 0 {
		return io.EOF
	}
	return err
}
