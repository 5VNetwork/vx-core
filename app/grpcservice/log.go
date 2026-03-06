// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package grpcservice

import (
	"context"

	"github.com/5vnetwork/vx-core/app/userlogger"

	"github.com/rs/zerolog/log"
)

// TODO?
// func (s *ClientGrpc) ChangeLogLevel(ctx context.Context, in *ChangeLogLevelRequest) (*ChangeLogLevelResponse, error) {
// 	return &ChangeLogLevelResponse{}, nil
// }

// func (s *Controller) LogStream(in *LogStreamRequest, stream Service_LogStreamServer) error {
// 	log.Debug().Msg("log stream called")
// 	cw := &logWriter{ch: make(chan *buf.Buffer, 500)}
// 	defer cw.Close()
// 	if err := s.GetLogger().AddOutputToGlobalLogger(cw); err != nil {
// 		return fmt.Errorf("cannot add output to logger: %w", err)
// 	}
// 	defer s.GetLogger().RemoveOutputFromGlobalLogger(cw)
// 	for {
// 		if s.closed {
// 			return nil
// 		}
// 		select {
// 		case <-stream.Context().Done():
// 			log.Debug().Msg("log stream context done")
// 			return nil
// 		case b, ok := <-cw.ch:
// 			if !ok {
// 				return nil
// 			}
// 			if err := stream.Send(&LogMessage{Message: b.String()}); err != nil {
// 				log.Error().Err(err).Msg("failed to send log message")
// 				b.Release()
// 				return err
// 			}
// 			b.Release()
// 		}
// 	}
// }

func (s *GrpcService) ToggleUserLog(ctx context.Context, in *ToggleUserLogRequest) (*ToggleUserLogResponse, error) {
	s.Client.UserLogger.SetEnabled(in.Enable)
	return &ToggleUserLogResponse{}, nil
}

func (s *GrpcService) ToggleLogAppId(ctx context.Context, in *ToggleLogAppIdRequest) (*ToggleLogAppIdResponse, error) {
	s.Client.UserLogger.LogAppId.Store(in.Enable)
	return &ToggleLogAppIdResponse{}, nil
}

func (s *GrpcService) UserLogStream(in *UserLogStreamRequest, stream GrpcService_UserLogStreamServer) error {
	log.Debug().Msg("user log stream request received")
	ul := s.Client.UserLogger
	ul.SetEnabled(true)
	slice := make([]*userlogger.UserLogMessage, 100)
	for {
		if s.Done.Done() {
			return nil
		}
		select {
		case <-s.Done.Wait():
			return nil
		case <-stream.Context().Done():
			return nil
		default:
			n, err := ul.ReadLog(stream.Context(), slice)
			if err != nil {
				return err
			}
			for _, msg := range slice[:n] {
				if err := stream.Send(msg); err != nil {
					log.Error().Err(err).Msg("failed to send user log message")
					return err
				}
			}
		}
	}
}
