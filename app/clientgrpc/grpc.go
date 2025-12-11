//go:build !server || test

package clientgrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/5vnetwork/vx-core/common/buf"
	"github.com/rs/zerolog/log"
)

type logWriter struct {
	closed bool
	ch     chan *buf.Buffer
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, errors.New("logWriter closed")
	}
	b := buf.New()
	b.Write(p)
	select {
	case w.ch <- b:
	default:
		log.Warn().Msg("log channel is full, drop log message")
		b.Release()
	}
	return len(p), nil
}

func (w *logWriter) Close() error {
	w.closed = true
	for b := range w.ch {
		b.Release()
	}
	close(w.ch)
	return nil
}

// should not be called concurrently
func (s *ClientGrpc) SwitchFakeDns(ctx context.Context, in *SwitchFakeDnsRequest) (*SwitchFakeDnsResponse, error) {
	if in.Enable {
		if err := s.EnableFakeDns(); err != nil {
			return nil, fmt.Errorf("failed to enable fake dns: %w", err)
		}
	} else {
		if err := s.DisableFakeDns(); err != nil {
			return nil, fmt.Errorf("failed to disable fake dns: %w", err)
		}
	}
	return &SwitchFakeDnsResponse{}, nil
}
