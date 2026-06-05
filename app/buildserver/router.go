//go:build server || test

package buildserver

import (
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/router"
	"github.com/5vnetwork/vx-core/i"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In
	Config          *configs.RouterConfig
	OutboundManager *outbound.Manager
	GeoHelper       *geo.Geo
	IpResolver      i.IPResolver `name:"request_domain_resolver"`
}
type RouterResult struct {
	fx.Out
	Router i.Router
}

func NewRouter(params RouterParams) (RouterResult, error) {
	router, err := router.NewRouter(&router.RouterConfig{
		RouterConfig:    params.Config,
		OutboundManager: params.OutboundManager,
		GeoHelper:       params.GeoHelper,
		IpResolver:      params.IpResolver,
	})
	if err != nil {
		return RouterResult{}, fmt.Errorf("failed to create router: %w", err)
	}

	return RouterResult{Router: router}, nil
}
