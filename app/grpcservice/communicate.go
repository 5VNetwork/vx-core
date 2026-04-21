// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package grpcservice

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/5vnetwork/vx-core/common/units"

	"github.com/5vnetwork/vx-core/i"
	"github.com/rs/zerolog/log"
)

func (s *GrpcService) setCommunicateStream(stream GrpcService_CommunicateServer) {
	s.streamLock.Lock()
	defer s.streamLock.Unlock()
	s.communicateStream = stream
}

func (s *GrpcService) getCommunicateStream() GrpcService_CommunicateServer {
	s.streamLock.RLock()
	defer s.streamLock.RUnlock()
	return s.communicateStream
}

func (s *GrpcService) Communicate(in *CommunicateRequest, stream GrpcService_CommunicateServer) error {
	s.setCommunicateStream(stream)
	defer s.setCommunicateStream(nil)

	if s.RunningInService {
		if s.timeoutExit != nil {
			s.timeoutExit.Stop()
			s.timeoutExit = nil
		}
	}

	// without this, ui might not know the handler being used promptly
	s.notifyHandlerBeingUsed()

	select {
	case <-stream.Context().Done():
		if s.RunningInService {
			// if after two seconds, flutter app is not connected, exit the service
			s.timeoutExit = time.AfterFunc(1*time.Second, func() {
				s.OnExit()
			})
		}
		return nil
	case <-s.Done.Wait():
		return nil
	}
}

// TODO: Lock?
func (s *GrpcService) OnSubscriptionUpdated() {
	if proxyHandlers := GetAllProxyhandlers(s.Client.OutboundManager); len(proxyHandlers) > 0 {
		var handlers []i.Outbound
		for _, ph := range proxyHandlers {
			id, err := strconv.Atoi(ph.Tag())
			if err != nil {
				log.Error().Err(err).Str("tag", ph.Tag()).Msg("failed to convert tag to id")
				continue
			}
			handler := s.Client.DB.GetHandler(id)
			if handler == nil {
				log.Error().Err(err).Msg("failed to get outbound handler")
				continue
			}
			h, err := s.Client.HandlerFactory.CreateHandler(handler.ToConfig())
			if err != nil {
				log.Error().Err(err).Msg("create outbound handler")
				continue
			}
			if h == nil {
				log.Warn().Str("tag", ph.Tag()).Msg("outbound handler is nil")
				continue
			}
			handlers = append(handlers, h)
		}
		ReplaceHandlers(s.Client.OutboundManager, handlers...)
	} else {
		s.Client.Selectors.OnHandlerChanged()
	}
	// TODO: replace node set
	// handlers, err := s.Client.DB.GetAllHandlers()
	// if err != nil {
	// 	log.Error().Err(err).Msg("get all handlers")
	// 	return
	// }
	// domains := make([]string, 0, len(handlers))
	// ips := make([]string, 0, len(handlers))
	// for _, handler := range handlers {
	// 	c := handler.ToConfig()
	// 	var address string
	// 	if c.GetOutbound() != nil {
	// 		address = c.GetOutbound().Address
	// 	} else if c.GetChain() != nil {
	// 		address = c.GetChain().Handlers[0].Address
	// 	}
	// 	if net.ParseAddress(address).Family().IsDomain() {
	// 		domains = append(domains, address)
	// 	} else {
	// 		ips = append(ips, address)
	// 	}
	// }

	// notify ui
	stream := s.getCommunicateStream()
	if stream == nil {
		return
	}
	err := stream.Send(&CommunicateMessage{
		Message: &CommunicateMessage_SubscriptionUpdate{
			SubscriptionUpdate: &SubscriptionUpdated{},
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("notify subscription updated")
	}
}

var handler4BeingUsed atomic.Value

func init() {
	handler4BeingUsed.Store("")
}

func (s *GrpcService) OnSelectedHandlersChanged(selector string, handlers []string) {
	if selector == "代理" {
		log.Debug().Msg("handler being used updated")
		if len(handlers) == 1 {
			handler4BeingUsed.Store(handlers[0])
			s.notifyHandlerBeingUsed()
		} else {
			handler4BeingUsed.Store("")
			s.notifyHandlerBeingUsed()
		}
	}
}

func (s *GrpcService) notifyHandlerBeingUsed() {
	stream := s.getCommunicateStream()
	if stream == nil {
		return
	}
	erro := stream.Send(&CommunicateMessage{
		Message: &CommunicateMessage_HandlerBeingUsed{
			HandlerBeingUsed: &HandlerBeingUsed{
				Tag4: handler4BeingUsed.Load().(string),
			},
		},
	})
	if erro != nil {
		log.Error().Err(erro).Msg("failed to notify handler being used")
	}
}

func (s *GrpcService) PingResult(tag string, ping int) {
	if !s.UpdateLantency {
		return
	}
	// store result into DB
	id, err := strconv.Atoi(tag)
	if err == nil {
		// TODO: update one field not replace all fields
		err := s.Client.DB.UpdateHandler(id, map[string]interface{}{
			"ping":           ping,
			"ok":             ping,
			"ping_test_time": time.Now().Unix(),
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to update handler")
		}
		s.notifyHandlerChange(id)
	}
}

func (s *GrpcService) UsableResult(tag string, ok bool) {
	// Store result into DB
	id, err := strconv.Atoi(tag)
	if err == nil {
		handler := s.Client.DB.GetHandler(id)
		if handler != nil {
			var newOk int
			if ok {
				newOk = 1
			} else {
				newOk = -1
			}
			if newOk != handler.Ok {
				err := s.Client.DB.UpdateHandler(id, map[string]interface{}{
					"ok": newOk,
				})
				if err != nil {
					log.Error().Err(err).Msg("failed to update handler")
				}
				s.notifyHandlerChange(id)
			}
		}
	}
}

func (s *GrpcService) SpeedResult(tag string, speed int64) {
	// store result into DB
	id, err := strconv.Atoi(tag)
	if err == nil {
		var speedToWrite float64
		if speed > 0 {
			speedToWrite = units.BytesToMb(speed)
		} else {
			speedToWrite = -1
		}
		err := s.Client.DB.UpdateHandler(id, map[string]interface{}{
			"speed":           speedToWrite,
			"ok":              int(speed),
			"speed_test_time": time.Now().Unix(),
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to update handler")
		}
		s.notifyHandlerChange(id)
	}
}

func (s *GrpcService) IPv6Result(tag string, ok bool) {
	if ok {
		// store result into DB
		id, err := strconv.Atoi(tag)
		if err == nil {
			support := 0
			if ok {
				support = 1
			} else {
				support = -1
			}
			err := s.Client.DB.UpdateHandler(id, map[string]interface{}{
				"support6":           support,
				"support6_test_time": time.Now().Unix(),
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to update handler")
			}
			s.notifyHandlerChange(id)
		}
	}
}

// notify ui about handler change
func (s *GrpcService) notifyHandlerChange(id int) {
	stream := s.getCommunicateStream()
	if stream == nil {
		return
	}
	erro := stream.Send(&CommunicateMessage{
		Message: &CommunicateMessage_HandlerUpdated{
			HandlerUpdated: &HandlerUpdated{
				Id: int64(id),
			},
		},
	})
	if erro != nil {
		log.Error().Err(erro).Msg("failed to notify handler updated")
	}
}
