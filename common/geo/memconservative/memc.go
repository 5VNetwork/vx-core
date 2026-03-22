package memconservative

import (
	"runtime"

	commongeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	"github.com/5vnetwork/vx-core/common/errors"
)

var (
	Loader = NewMemConservativeLoader()
)

type MemConservativeLoader struct {
	geoipcache   GeoIPCache
	geositecache GeoSiteCache
}

func (m *MemConservativeLoader) LoadIP(filepath, country string) (*commongeo.GeoIP, error) {
	defer runtime.GC()
	geoip, err := m.geoipcache.Unmarshal(filepath, country)
	if err != nil {
		return nil, errors.New("failed to decode geodata file: ", filepath).Base(err)
	}
	return geoip, nil
}

func (m *MemConservativeLoader) LoadSite(filepath, list string) (*commongeo.GeoSite, error) {
	defer runtime.GC()
	geosite, err := m.geositecache.Unmarshal(filepath, list)
	if err != nil {
		return nil, errors.New("failed to decode geodata file: ", filepath).Base(err)
	}
	return geosite, nil
}

func NewMemConservativeLoader() *MemConservativeLoader {
	return &MemConservativeLoader{make(map[string]*commongeo.GeoIP), make(map[string]*commongeo.GeoSite)}
}
