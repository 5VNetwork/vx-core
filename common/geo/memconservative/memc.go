package memconservative

import (
	"runtime"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/geo"
)

var (
	Loader = NewMemConservativeLoader()
)

type MemConservativeLoader struct {
	geoipcache   GeoIPCache
	geositecache GeoSiteCache
}

func (m *MemConservativeLoader) LoadIP(filepath, country string) (*geo.GeoIP, error) {
	defer runtime.GC()
	geoip, err := m.geoipcache.Unmarshal(filepath, country)
	if err != nil {
		return nil, errors.New("failed to decode geodata file: ", filepath).Base(err)
	}
	return geoip, nil
}

func (m *MemConservativeLoader) LoadSite(filepath, list string) (*geo.GeoSite, error) {
	defer runtime.GC()
	geosite, err := m.geositecache.Unmarshal(filepath, list)
	if err != nil {
		return nil, errors.New("failed to decode geodata file: ", filepath).Base(err)
	}
	return geosite, nil
}

func NewMemConservativeLoader() *MemConservativeLoader {
	return &MemConservativeLoader{make(map[string]*geo.GeoIP), make(map[string]*geo.GeoSite)}
}
