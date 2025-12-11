package buildclient

import (
	"fmt"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/clientgrpc"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/app/router/selector"
	"github.com/5vnetwork/vx-core/app/tester"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
)

func Handler(config *configs.TmConfig, fc *Builder, cc *client.Client) error {
	d := dispatcher.New()
	d.DestinationOverride = config.GetDispatcher().GetDestinationOverride()
	d.FallbackToProxy = config.GetDispatcher().GetFallbackToProxy()
	d.FallbackToDomain = config.GetDispatcher().GetFallbackToDomain()
	d.Sniff = config.GetDispatcher().GetSniff()
	d.OutStats = outbound.NewOutStats()

	cc.Dispatcher = d
	fc.requireFeature(func(om *outbound.Manager, p *policy.Policy,
		ul *userlogger.UserLogger) {
		d.Om = om
		d.TimeoutPolicy = p
		d.StatsPolicy = p
		d.UserLogger = ul
	})
	fc.requireOptionalFeatures(func(id *dns.Dns) {
		d.FakeDns = cc.AllFakeDns
		d.Dns = cc.IPResolverForRequestAddress
	})

	selectors := selector.NewSelectors()
	d.Sm = selectors
	common.Must(fc.addComponent(selectors))
	cc.Selectors = selectors
	if len(config.GetSelectors().GetSelectors()) > 0 {
		err := fc.requireFeature(func(dispatcher *dispatcher.Dispatcher, tester *tester.Tester,
			db client.Db, grpc *clientgrpc.ClientGrpc, om *outbound.Manager) error {
			for _, selectorConfig := range config.Selectors.Selectors {
				landHandlers := make([]*xsqlite.OutboundHandler, 0, len(selectorConfig.LandHandlers))
				for _, landHandlerId := range selectorConfig.LandHandlers {
					handler := db.GetHandler(int(landHandlerId))
					if handler == nil {
						return fmt.Errorf("land handler %d not found", landHandlerId)
					}
					landHandlers = append(landHandlers, handler)
				}
				selectors.AddSelector(selector.NewSelector(selector.SelectorConfig{
					SelectorConfig:            selectorConfig,
					CreateHandler:             cc.CreateHandler,
					HandlerErrorChangeSubject: dispatcher,
					Tester:                    tester,
					Database:                  db,
					OnHandlerBeingUsedChange: func(s []string) {
						grpc.OnHandlerBeingUsedUpdated(selectorConfig.Tag, s)
					},
					LandHandlers:    landHandlers,
					OutboundManager: om,
				}))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// router
	err := fc.requireFeature(func(om *outbound.Manager, g i.GeoHelper, ipr i.IPResolver) error {
		r, err := router.NewRouter(&router.RouterConfig{
			RouterConfig:    config.Router,
			GeoHelper:       g,
			OutboundManager: om,
			Selectors:       selectors,
			IpResolver:      cc.IPResolverForRequestAddress,
		})
		if err != nil {
			return err
		}
		routerWrapper := &router.RouterWrapper{}
		routerWrapper.UpdateRouter(r)
		d.Router = routerWrapper
		cc.Router = routerWrapper
		fc.addComponent(routerWrapper)

		return nil
	})
	if err != nil {
		return err
	}
	// Not notify ui about handler error for now
	// fc.requireOptionalFeatures(func(g *clientgrpc.ClientGrpc) {
	// 	d.AddHandlerErrorObserver(g)
	// })
	common.Must(fc.addComponent(d))
	return nil
}
