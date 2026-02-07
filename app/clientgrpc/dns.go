// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package clientgrpc

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

func (d *ClientGrpc) EnableFakeDns() error {
	if d.Client.AllFakeDns != nil {
		log.Info().Msg("fake dns enabled")
		d.Client.SetFakeDnsEnabled(true)
	}
	return nil
}

func (d *ClientGrpc) DisableFakeDns() error {
	if d.Client.AllFakeDns != nil {
		log.Info().Msg("fake dns disabled")
		d.Client.SetFakeDnsEnabled(false)
	}
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
