package clientgrpc

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/i"

	"github.com/rs/zerolog/log"
)

func (s *ClientGrpc) UpdateRouter(ctx context.Context, in *UpdateRouterRequest) (*UpdateRouterResponse, error) {
	log.Info().Msg("update router")

	err := s.updateRouter(in.RouterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update router: %w", err)
	}

	return &UpdateRouterResponse{}, nil
}

func (s *ClientGrpc) ChangeRoutingMode(ctx context.Context, in *ChangeRoutingModeRequest) (*ChangeRoutingModeResponse, error) {
	log.Debug().Msg("ChangeRoutingMode")
	err := s.Client.Geo.UpdateGeo(in.GeoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create geo: %w", err)
	}
	log.Debug().Msg("geo updated")
	if err := s.updateRouter(in.RouterConfig); err != nil {
		return nil, fmt.Errorf("failed to updateRouter: %w", err)
	}

	log.Debug().Msg("routing mode changed")
	return &ChangeRoutingModeResponse{}, nil
}
func (s *ClientGrpc) updateRouter(config *configs.RouterConfig) error {
	newRouter, err := router.NewRouter(&router.RouterConfig{
		RouterConfig:    config,
		OutboundManager: s.Client.OutboundManager,
		GeoHelper:       s.Client.Geo,
		Selectors:       s.Client.Selectors,
		IpResolver:      s.Client.IPResolverForRequestAddress,
	})
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}
	s.Client.Router.UpdateRouter(newRouter)
	return nil
}

func (s *ClientGrpc) ChangeOutbound(ctx context.Context, in *ChangeOutboundRequest) (*ChangeOutboundResponse, error) {
	log.Debug().Msg("ChangeOutbound")
	// s.AutoOutbound = in.GetAutoOutbound()
	// s.Policy.SetOutboundStats(s.AutoOutbound)
	om := s.Client.OutboundManager
	handlers := make([]i.Outbound, 0, len(in.GetHandlers()))
	for _, handler := range in.GetHandlers() {
		h, err := s.Client.CreateHandler(handler, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbound handler: %w", err)
		}
		handlers = append(handlers, h)
	}
	if in.GetDeleteAll() {
		om.ReplaceHandlers(handlers...)
	} else {
		om.RemoveHandlers(in.GetTags())
		om.AddHandlers(handlers...)
	}
	return &ChangeOutboundResponse{}, nil
}

func (s *ClientGrpc) CurrentOutbound(ctx context.Context, in *CurrentOutboundRequest) (*CurrentOutboundResponse, error) {
	om := s.Client.OutboundManager
	tags := make([]string, 0, len(outbound.GetAllProxyhandlers(om)))
	for _, h := range outbound.GetAllProxyhandlers(om) {
		tags = append(tags, h.Tag())
	}
	return &CurrentOutboundResponse{
		OutboundTags: tags,
	}, nil
}

func (s *ClientGrpc) NotifyHandlerChange(context.Context, *HandlerChangeNotify) (*HandlerChangeNotifyResponse, error) {
	log.Info().Msg("NotifyHandlerChange")
	s.Client.Selectors.OnHandlerChanged()
	return &HandlerChangeNotifyResponse{}, nil
}

func (s *ClientGrpc) ChangeSelector(ctx context.Context, in *ChangeSelectorRequest) (*ChangeSelectorResponse, error) {
	log.Info().Msg("ChangeSelector")
	if in.SelectorToRemove != "" {
		s.Client.Selectors.RemoveSelector(in.SelectorToRemove)
	}
	if in.DeleteAll {
		s.Client.Selectors.RemoveAllSelectors()
	}
	for _, selectorConfig := range in.GetSelectorsToAdd() {
		landHandlers := make([]*xsqlite.OutboundHandler, 0, len(selectorConfig.LandHandlers))
		for _, landHandlerId := range selectorConfig.LandHandlers {
			handler := s.Client.DB.GetHandler(int(landHandlerId))
			if handler == nil {
				return nil, fmt.Errorf("land handler %d not found", landHandlerId)
			}
			landHandlers = append(landHandlers, handler)
		}
		s.Client.Selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
			SelectorConfig:            selectorConfig,
			CreateHandler:             s.Client.CreateHandler,
			HandlerErrorChangeSubject: s.Client.Dispatcher,
			Tester:                    s.Client.Tetser,
			Database:                  s.Client.DB,
			OutboundManager:           s.Client.OutboundManager,
			OnHandlerBeingUsedChange: func(v []string) {
				s.OnHandlerBeingUsedUpdated(selectorConfig.Tag, v)
			},
			LandHandlers: landHandlers,
		}))
	}
	return &ChangeSelectorResponse{}, nil
}

func (s *ClientGrpc) UpdateSelectorBalancer(ctx context.Context, in *UpdateSelectorBalancerRequest) (*Receipt, error) {
	log.Info().Msg("UpdateSelectorBalancer")

	var balancer selector.Balancer
	switch in.BalanceStrategy {
	case configs.SelectorConfig_RANDOM:
		balancer = selector.NewRandomBanlancer()
	case configs.SelectorConfig_MEMORY:
		balancer = selector.NewMemoryBalancer()
	}
	se := s.Client.Selectors.GetSelector(in.Tag)
	if se == nil {
		return nil, fmt.Errorf("selector not found: %s", in.Tag)
	}
	se.UpdateBalancer(balancer)
	return &Receipt{}, nil
}

func (s *ClientGrpc) UpdateSelectorFilter(ctx context.Context, in *UpdateSelectorFilterRequest) (*Receipt, error) {
	log.Info().Msg("UpdateSelectorFilter")
	se := s.Client.Selectors.GetSelector(in.Tag)
	if se == nil {
		return nil, fmt.Errorf("selector not found: %s", in.Tag)
	}
	var filter selector.Filter
	if in.SelectFromOm {
		filter = selector.NewOmFilter(in.GetFilter(), s.Client.OutboundManager)
	} else {
		filter = selector.NewDbFilter(s.Client.DB, in.GetFilter(),
			se.LandHandlers, s.Client.CreateHandler)
	}
	se.UpdateFilter(filter)
	return &Receipt{}, nil
}

func (s *ClientGrpc) SetOutboundHandlerSpeed(ctx context.Context, in *SetOutboundHandlerSpeedRequest) (*SetOutboundHandlerSpeedResponse, error) {
	log.Debug().Str("tag", in.GetTag()).Int32("speed", in.GetSpeed()).Msg("SetOutboundHandlerSpeed")
	s.Client.Selectors.OnHandlerSpeedChanged(in.GetTag(), in.GetSpeed())
	return &SetOutboundHandlerSpeedResponse{}, nil
}

// func (s *ClientGrpc) SetLandHandler(ctx context.Context, in *SetLandHandlerRequest) (*SetLandHandlerResponse, error) {
// 	log.Debug().Msg("SetLandHandler")
// 	if len(in.GetHandlers()) == 0 {
// 		s.Client.Lock()
// 		s.Client.LandHandlerDF = nil
// 		s.Client.Unlock()
// 	} else {
// 		ch, err := outbound.NewChainHandler(&outbound.ChainHandlerConfig{
// 			ChainHandlerConfig: &configs.ChainHandlerConfig{
// 				Handlers: in.GetHandlers(),
// 			},
// 			Policy:         s.Client.Policy,
// 			UseDnsDialer:   true,
// 			IPResolver:     s.Client.IPResolver,
// 			DefaultNICInfo: s.Client.DefaultNICInfo,
// 			DF:             s.Client.DialerFactory,
// 		})
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create land handler: %w", err)
// 		}
// 		s.Client.Lock()
// 		s.Client.LandHandlerDF = &transport.HandlerDialerFactory{Handler: ch}
// 		s.Client.Unlock()
// 	}
// 	return &SetLandHandlerResponse{}, nil
// }

// func (s *Controller) SpeedTest(req *SpeedTestRequest, in Service_SpeedTestServer) error {
// 	wg := new(errgroup.Group)
// 	for _, t := range req.GetHandlers() {
// 		wg.Go(func() error {
// 			h, err := create.NewOutHandler(t, s.getDialerFactory(), s.getPolicy(), s.getStats())
// 			if err != nil {
// 				log.Err(err).Str("tag", t.GetTag()).Msg("failed to create handler")
// 				if err := in.Send(&SpeedTestResponse{
// 					Ok:  false,
// 					Tag: t.GetTag(),
// 				}); err != nil {
// 					log.Err(err).Msg("failed to send speed test response")
// 					return err
// 				}
// 				return nil
// 			}
// 			rst, err := st.Run(in.Context(), speedtest.WithDoer(outbound.HandlerToHttpClient(h)))
// 			if err != nil {
// 				log.Err(err).Str("tag", t.GetTag()).Msg("failed to run speed test")
// 				if err := in.Send(&SpeedTestResponse{
// 					Ok:  false,
// 					Tag: t.GetTag(),
// 				}); err != nil {
// 					log.Err(err).Msg("failed to send speed test response")
// 					return err
// 				}
// 				return nil
// 			}
// 			if err := in.Send(&SpeedTestResponse{
// 				Ok:   true,
// 				Tag:  h.Tag(),
// 				Up:   float32(rst.Upload),
// 				Down: float32(rst.Download),
// 				Ping: uint32(rst.Latency),
// 			}); err != nil {
// 				log.Err(err).Msg("failed to send speed test response")
// 				return err
// 			}
// 			return nil
// 		})
// 	}
// 	return wg.Wait()
// }
