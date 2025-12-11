package stdloader

import (
	"errors"
	"os"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/clashconfig"
	"github.com/5vnetwork/vx-core/common/geo"
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
)

type StandartLoader struct {
	ipCache   map[string]*geo.GeoIPList
	siteCache map[string]*geo.GeoSiteList
}

func NewStandartLoader() *StandartLoader {
	return &StandartLoader{
		ipCache:   make(map[string]*geo.GeoIPList),
		siteCache: make(map[string]*geo.GeoSiteList),
	}
}

func (s *StandartLoader) LoadIP(filepath, countryCode string) (*geo.GeoIP, error) {
	geoipList := s.ipCache[filepath]
	if geoipList == nil {
		geoipList = &geo.GeoIPList{}
		geoipBytes, err := os.ReadFile(filepath)
		if err != nil {
			return nil, errors.New("failed to open file: " + filepath)
		}
		log.Debug().Float64("size", float64(len(geoipBytes))/1024/1024).Msg("geoip file size")
		if err := proto.Unmarshal(geoipBytes, geoipList); err != nil {
			return nil, err
		}
		s.ipCache[filepath] = geoipList
	}

	for _, geoip := range geoipList.Entry {
		if strings.EqualFold(geoip.CountryCode, countryCode) {
			return geoip, nil
		}
	}

	return nil, errors.New("country not found in " + filepath + ": " + countryCode)
}

func (s *StandartLoader) LoadSite(filepath, siteName string) (*geo.GeoSite, error) {
	geositeList := s.siteCache[filepath]

	if geositeList == nil {
		geositeList = &geo.GeoSiteList{}
		geositebytes, err := os.ReadFile(filepath)
		if err != nil {
			return nil, errors.New("failed to open file: " + filepath)
		}
		log.Debug().Float64("size", float64(len(geositebytes))/1024/1024).Msg("geosite file size")
		if err := proto.Unmarshal(geositebytes, geositeList); err != nil {
			return nil, err
		}
		s.siteCache[filepath] = geositeList
	}
	for _, site := range geositeList.Entry {
		if strings.EqualFold(site.CountryCode, siteName) {
			return site, nil
		}
	}

	return nil, errors.New("list not found in " + filepath + ": " + siteName)
}

func (s *StandartLoader) LoadDomainsClash(filepath string) ([]*geo.Domain, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return clashconfig.ExtractDomainsFromClashRules(file)
}

func (s *StandartLoader) LoadCidrsClash(filepath string) ([]*geo.CIDR, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return clashconfig.ExtractCidrFromClashRules(file)
}

func (s *StandartLoader) LoadAppsClash(filepath string) ([]*configs.AppId, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return clashconfig.ExtractAppsFromClashRules(file)
}
