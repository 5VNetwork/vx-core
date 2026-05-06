//go:build server || test

package buildserver

import (
	"fmt"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/geo"
	"go.uber.org/fx"
)

type GeoHelperParams struct {
	fx.In
	Config *configs.GeoConfig
}
type GeoHelperResult struct {
	fx.Out
	GeoHelper *geo.Geo
}

func NewGeoHelper(params GeoHelperParams) (GeoHelperResult, error) {
	geoHelper, err := geo.NewGeo(params.Config)
	if err != nil {
		return GeoHelperResult{}, fmt.Errorf("failed to create geo helper: %w", err)
	}
	return GeoHelperResult{GeoHelper: geoHelper}, nil
}
