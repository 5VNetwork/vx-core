//go:build server || test

package buildserver

import (
	"context"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/geo/geosync"
	"go.uber.org/fx"
)

type GeoSyncParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *configs.GeoConfig
	GeoHelper *geo.Geo
}

type GeoSyncResult struct {
	fx.Out
	GeoSync *geosync.GeoSync
}

func NewGeoSync(lc fx.Lifecycle, p GeoSyncParams) (GeoSyncResult, error) {
	gs := geosync.New(p.GeoHelper)
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			gs.Reconfigure(p.Config)
			return gs.Start()
		},
		OnStop: func(_ context.Context) error {
			return gs.Close()
		},
	})
	return GeoSyncResult{
		GeoSync: gs,
	}, nil
}
