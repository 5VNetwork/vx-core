//go:build server || test

package buildserver

import (
	"context"
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/geo/geosync"
	"github.com/5vnetwork/vx-core/i"
	"go.uber.org/fx"
)

type GeoHelperParams struct {
	fx.In
	Config *configs.GeoConfig
}
type GeoHelperResult struct {
	fx.Out
	Geo       *geo.Geo
	GeoHelper i.GeoHelper
}

func NewGeoHelper(params GeoHelperParams) (GeoHelperResult, error) {
	gw := &geo.Geo{}
	if err := geosync.PrefetchGeoDataFiles(context.Background(), params.Config); err != nil {
		return GeoHelperResult{}, fmt.Errorf("geo url prefetch: %w", err)
	}
	if err := gw.UpdateGeo(params.Config); err != nil {
		return GeoHelperResult{}, fmt.Errorf("failed to create geo helper: %w", err)
	}
	return GeoHelperResult{Geo: gw, GeoHelper: gw}, nil
}
