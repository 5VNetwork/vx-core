package api

import (
	"bytes"
	context "context"
	"fmt"
	"net"
	"os"
	"time"

	ag "github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/common/clashconfig"
	"github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/common/geo/memconservative"
	"github.com/5vnetwork/vx-core/common/geo/memloader"
	"github.com/5vnetwork/vx-core/common/signal"
	"google.golang.org/protobuf/proto"
)

// var ErrGeoIPFileNotFound = errors.New("geo ip file does not exist")

// TODO: more efficient
func (a *Api) GeoIP(ctx context.Context, req *GeoIPRequest) (*GeoIPResponse, error) {
	a.geoLock.Lock()

	var ipMatcher *geoIpMatcher
	if a.geoIpMatcher == nil {
		// // check if geo ip file exists
		// if _, err := os.Stat(a.GeoipPath); os.IsNotExist(err) {
		// 	a.geoLock.Unlock()
		// 	return nil, ErrGeoIPFileNotFound
		// }
		geositebytes, err := os.ReadFile(a.GeoipPath)
		if err != nil {
			a.geoLock.Unlock()
			return nil, fmt.Errorf("failed to open geo ip file: %w", err)
		}
		var geositeList geo.GeoIPList
		if err := proto.Unmarshal(geositebytes, &geositeList); err != nil {
			a.geoLock.Unlock()
			return nil, err
		}

		ipMatcher = &geoIpMatcher{
			matchersmap: make(map[string]*geo.IPMatcher),
		}
		for _, geoip := range geositeList.Entry {
			if geoip.CountryCode == "private" || len(geoip.CountryCode) != 2 {
				continue
			}
			ipMatcher.matchersmap[geoip.CountryCode], err = geo.NewIPMatcherFromGeoCidrs(
				geoip.Cidr, false)
			if err != nil {
				a.geoLock.Unlock()
				return nil, err
			}
		}
		a.geoIpMatcher = ipMatcher
		a.timeoutChecker = signal.NewActivityChecker(func() {
			a.geoLock.Lock()
			a.geoIpMatcher = nil
			a.timeoutChecker = nil
			a.geoLock.Unlock()
		}, 10*time.Second)
	} else {
		ipMatcher = a.geoIpMatcher
	}

	a.timeoutChecker.Update()
	a.geoLock.Unlock()

	rsp := &GeoIPResponse{
		Countries: make([]string, len(req.Ips)),
	}
	for i, ipStr := range req.Ips {
		ip := net.ParseIP(ipStr)
		if ip != nil {
			country, found := ipMatcher.Match(ip)
			if found {
				rsp.Countries[i] = country
				continue
			}
		}
		rsp.Countries[i] = ""
	}
	return rsp, nil
}

type geoIpMatcher struct {
	matchersmap map[string]*geo.IPMatcher
}

func (g *geoIpMatcher) Match(ip net.IP) (string, bool) {
	for country, matcher := range g.matchersmap {
		if matcher.Match(ip) {
			return country, true
		}
	}
	return "", false
}

func (a *Api) ProcessGeoFiles(ctx context.Context, req *ProcessGeoFilesRequest) (*ProcessGeoFilesResponse, error) {
	l := memconservative.NewMemConservativeLoader()
	geositeList := &geo.GeoSiteList{
		Entry: []*geo.GeoSite{},
	}
	geoIpList := &geo.GeoIPList{
		Entry: []*geo.GeoIP{},
	}
	for _, code := range req.GeositeCodes {
		site, err := l.LoadSite(req.GeositePath, code)
		if err != nil {
			return nil, err
		}
		// log.Debug().Int("len", len(site.Domain)).Str("code", code).Msg("geosite cidr")
		// if code == "cn" {
		// 	numFull := 0
		// 	numDomain := 0
		// 	for _, domain := range site.Domain {
		// 		if domain.Type == geo.Domain_Full {
		// 			numFull++
		// 		}
		// 		if domain.Type == geo.Domain_RootDomain {
		// 			numDomain++
		// 		}
		// 		log.Debug().Str("domain", domain.Value).Int("type", int(domain.Type)).Msg("geosite domain")
		// 	}
		// 	log.Debug().Int("numFull", numFull).Int("numDomain", numDomain).Msg("geosite domain")
		// }
		geositeList.Entry = append(geositeList.Entry, site)
	}
	for _, code := range req.GeoipCodes {
		cidr, err := l.LoadIP(req.GeoipPath, code)
		if err != nil {
			return nil, err
		}
		// log.Debug().Int("len", len(cidr.Cidr)).Str("code", code).Msg("geoip cidr")
		geoIpList.Entry = append(geoIpList.Entry, cidr)
	}

	// write into files
	// Marshal the geo data to protobuf format
	geositeBytes, err := proto.Marshal(geositeList)
	if err != nil {
		return nil, err
	}

	geoipBytes, err := proto.Marshal(geoIpList)
	if err != nil {
		return nil, err
	}

	tempGeosite := req.DstGeositePath + ".tmp"
	err = os.WriteFile(tempGeosite, geositeBytes, 0644)
	if err != nil {
		return nil, err
	}
	err = os.Rename(tempGeosite, req.DstGeositePath)
	if err != nil {
		// Clean up temporary file if rename fails
		os.Remove(tempGeosite)
		return nil, err
	}

	tempGeoIpFile := req.DstGeoipPath + ".tmp"
	err = os.WriteFile(tempGeoIpFile, geoipBytes, 0644)
	if err != nil {
		return nil, err
	}
	err = os.Rename(tempGeoIpFile, req.DstGeoipPath)
	if err != nil {
		// Clean up temporary file if rename fails
		os.Remove(tempGeoIpFile)
		return nil, err
	}

	return &ProcessGeoFilesResponse{}, nil
}

func (a *Api) ParseGeositeConfig(ctx context.Context, req *ParseGeositeConfigRequest) (*ParseGeositeConfigResponse, error) {
	l := memloader.New()
	var domains []*geo.Domain
	var err error
	domains, err = ag.GeositeConfigToGeoDomains(req.Config, l)
	if err != nil {
		return nil, err
	}

	return &ParseGeositeConfigResponse{
		Domains: domains,
	}, nil
}

func (a *Api) ParseGeoIPConfig(ctx context.Context, req *ParseGeoIPConfigRequest) (*ParseGeoIPConfigResponse, error) {
	l := memloader.New()
	var cidrs []*geo.CIDR
	var err error
	cidrs, err = ag.GeoIpConfigToCidrs(req.Config, l)
	if err != nil {
		return nil, err
	}

	return &ParseGeoIPConfigResponse{
		Cidrs: cidrs,
	}, nil
}

func (a *Api) ParseClashRuleFile(c context.Context, req *ParseClashRuleFileRequest) (*ParseClashRuleFileResponse, error) {
	reader := bytes.NewReader(req.Content)
	apps, err := clashconfig.ExtractAppsFromClashRules(reader)
	if err != nil {
		return nil, err
	}
	reader = bytes.NewReader(req.Content)
	cidrs, err := clashconfig.ExtractCidrFromClashRules(reader)
	if err != nil {
		return nil, err
	}
	reader = bytes.NewReader(req.Content)
	domains, err := clashconfig.ExtractDomainsFromClashRules(reader)
	if err != nil {
		return nil, err
	}
	return &ParseClashRuleFileResponse{
		AppIds:  apps,
		Cidrs:   cidrs,
		Domains: domains,
	}, nil
}
