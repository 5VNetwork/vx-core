package memconservative

import (
	"os"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/geo"
)

type GeoIPCache map[string]*geo.GeoIP

func (g GeoIPCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g GeoIPCache) Get(key string) *geo.GeoIP {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g GeoIPCache) Set(key string, value *geo.GeoIP) {
	if g == nil {
		g = make(map[string]*geo.GeoIP)
	}
	g[key] = value
}

func (g GeoIPCache) Unmarshal(filePath, code string) (*geo.GeoIP, error) {
	// asset := platform.GetAssetLocation(filename)
	idx := strings.ToLower(filePath + ":" + code)
	if g.Has(idx) {
		return g.Get(idx), nil
	}

	geoipBytes, err := Decode(filePath, code)
	switch err {
	case nil:
		var geoip geo.GeoIP
		if err := proto.Unmarshal(geoipBytes, &geoip); err != nil {
			return nil, err
		}
		g.Set(idx, &geoip)
		return &geoip, nil

	case errCodeNotFound:
		return nil, errors.New("country code ", code, " not found in ", filePath)

	case errFailedToReadBytes, errFailedToReadExpectedLenBytes,
		errInvalidGeodataFile, errInvalidGeodataVarintLength:
		errors.New("failed to decode geoip file: ", filePath, ", fallback to the original ReadFile method")
		geoipBytes, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		var geoipList geo.GeoIPList
		if err := proto.Unmarshal(geoipBytes, &geoipList); err != nil {
			return nil, err
		}
		for _, geoip := range geoipList.GetEntry() {
			if strings.EqualFold(code, geoip.GetCountryCode()) {
				g.Set(idx, geoip)
				return geoip, nil
			}
		}

	default:
		return nil, err
	}

	return nil, errors.New("country code ", code, " not found in ", filePath)
}

type GeoSiteCache map[string]*geo.GeoSite

func (g GeoSiteCache) Has(key string) bool {
	return !(g.Get(key) == nil)
}

func (g GeoSiteCache) Get(key string) *geo.GeoSite {
	if g == nil {
		return nil
	}
	return g[key]
}

func (g GeoSiteCache) Set(key string, value *geo.GeoSite) {
	if g == nil {
		g = make(map[string]*geo.GeoSite)
	}
	g[key] = value
}

func (g GeoSiteCache) Unmarshal(filepath, code string) (*geo.GeoSite, error) {
	idx := strings.ToLower(filepath + ":" + code)
	if g.Has(idx) {
		return g.Get(idx), nil
	}

	geositeBytes, err := Decode(filepath, code)
	switch err {
	case nil:
		var geosite geo.GeoSite
		if err := proto.Unmarshal(geositeBytes, &geosite); err != nil {
			return nil, err
		}
		g.Set(idx, &geosite)
		return &geosite, nil

	case errCodeNotFound:
		return nil, errors.New("list ", code, " not found in ", filepath)

	case errFailedToReadBytes, errFailedToReadExpectedLenBytes,
		errInvalidGeodataFile, errInvalidGeodataVarintLength:
		errors.New("failed to decode geoip file: ", filepath, ", fallback to the original ReadFile method")
		geositeBytes, err = os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		var geositeList geo.GeoSiteList
		if err := proto.Unmarshal(geositeBytes, &geositeList); err != nil {
			return nil, err
		}
		for _, geosite := range geositeList.GetEntry() {
			if strings.EqualFold(code, geosite.GetCountryCode()) {
				g.Set(idx, geosite)
				return geosite, nil
			}
		}

	default:
		return nil, err
	}

	return nil, errors.New("list ", code, " not found in ", filepath)
}
