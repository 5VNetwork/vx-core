package memloader

import (
	"os"
	"runtime"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/clashconfig"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/geo"
	"google.golang.org/protobuf/proto"
)

var (
	Loader = New()
)

type MemLoader struct {
}

func (m *MemLoader) LoadIP(filePath, code string) (*geo.GeoIP, error) {
	defer runtime.GC()
	geoipBytes, err := Decode(filePath, code)
	switch err {
	case nil:
		var geoip geo.GeoIP
		if err := proto.Unmarshal(geoipBytes, &geoip); err != nil {
			return nil, err
		}
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
				return geoip, nil
			}
		}

	default:
		return nil, err
	}

	return nil, errors.New("country code ", code, " not found in ", filePath)
}

func (m *MemLoader) LoadSite(filepath, code string) (*geo.GeoSite, error) {
	defer runtime.GC()

	geositeBytes, err := Decode(filepath, code)
	switch err {
	case nil:
		var geosite geo.GeoSite
		if err := proto.Unmarshal(geositeBytes, &geosite); err != nil {
			return nil, err
		}
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
				return geosite, nil
			}
		}

	default:
		return nil, err
	}

	return nil, errors.New("list ", code, " not found in ", filepath)
}

func (s *MemLoader) LoadDomainsClash(filepath string) ([]*geo.Domain, error) {
	defer runtime.GC()

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return clashconfig.ExtractDomainsFromClashRules(file)
}

func (s *MemLoader) LoadCidrsClash(filepath string) ([]*geo.CIDR, error) {
	defer runtime.GC()

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return clashconfig.ExtractCidrFromClashRules(file)
}

func (s *MemLoader) LoadAppsClash(filepath string) ([]*configs.AppId, error) {
	defer runtime.GC()

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return clashconfig.ExtractAppsFromClashRules(file)
}

func New() *MemLoader {
	return &MemLoader{}
}
