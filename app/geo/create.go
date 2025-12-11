package geo

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	configs "github.com/5vnetwork/vx-core/app/configs"

	cgeo "github.com/5vnetwork/vx-core/common/geo"
	"github.com/5vnetwork/vx-core/common/geo/memloader"
	"github.com/5vnetwork/vx-core/common/geo/stdloader"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/5vnetwork/vx-core/i"
)

func NewGeo(config *configs.GeoConfig) (*Geo, error) {
	geo := &Geo{}

	oppositeDomainTags := make(map[string]string)
	for _, domainGroup := range config.GetGreatDomainSets() {
		if domainGroup.OppositeName != "" {
			oppositeDomainTags[domainGroup.Name] = domainGroup.OppositeName
			oppositeDomainTags[domainGroup.OppositeName] = domainGroup.Name
		}
	}
	oppositeIpTags := make(map[string]string)
	for _, ipGroup := range config.GetGreatIpSets() {
		if ipGroup.OppositeName != "" {
			oppositeIpTags[ipGroup.Name] = ipGroup.OppositeName
			oppositeIpTags[ipGroup.OppositeName] = ipGroup.Name
		}
	}

	var l loader
	if runtime.GOOS == "ios" {
		l = memloader.New()
	} else {
		l = stdloader.NewStandartLoader()
	}

	domainSets := make(map[string]i.DomainSet)
	for _, atomicSet := range config.GetAtomicDomainSets() {
		matcher, err := AtomicDomainSetToIndexMatcher(atomicSet, l)
		if err != nil {
			return nil, fmt.Errorf("failed to create domain matcher: %w", err)
		}
		// common.Log()
		domainSets[atomicSet.Name] = &IndexMatcherToDomainSet{
			IndexMatcher: matcher}
	}
	for _, greatSet := range config.GetGreatDomainSets() {
		d, err := getGreatDomainSet(greatSet, geo)
		if err != nil {
			return nil, fmt.Errorf("failed to create domain matcher: %w", err)
		}
		// common.Log()
		domainSets[greatSet.Name] = d
	}
	// log.Debug().Msg("domain loaded")
	runtime.GC()

	// ip sets
	ipSets := make(map[string]i.IPSet)
	for _, atomicSet := range config.GetAtomicIpSets() {
		ipMatcher, err := AtomicIpSetToIPMatcher(atomicSet, l)
		if err != nil {
			return nil, fmt.Errorf("failed to create ip matcher: %w", err)
		}
		ipSets[atomicSet.Name] = ipMatcher
	}
	for _, greatIpSet := range config.GetGreatIpSets() {
		m, err := getGreatIPSet(greatIpSet, ipSets)
		if err != nil {
			return nil, fmt.Errorf("failed to create ip matcher: %w", err)
		}
		ipSets[greatIpSet.Name] = m
	}
	// app sets
	appSets := make(map[string]i.AppSet)
	for _, appSet := range config.GetAppSets() {
		set, err := AppSetConfigToAppSet(appSet, l)
		if err != nil {
			return nil, fmt.Errorf("failed to create app set: %w", err)
		}
		appSets[appSet.Name] = set
	}

	l = nil
	runtime.GC()

	geo.OppositeDomainTags = oppositeDomainTags
	geo.DomainSets = domainSets
	geo.OppositeIpTags = oppositeIpTags
	geo.IpSets = ipSets
	geo.AppSets = appSets

	return geo, nil
}

func AppSetConfigToAppSet(c *configs.AppSetConfig, l loader) (i.AppSet, error) {
	indexMatcher := strmatcher.NewMphIndexMatcher()
	var appIds []*configs.AppId
	appIds = append(appIds, c.AppIds...)
	for _, clashFile := range c.ClashFiles {
		values, err := l.LoadAppsClash(clashFile)
		if err != nil {
			return nil, fmt.Errorf("failed to extract apps from clash file: %w", err)
		}
		appIds = append(appIds, values...)
	}
	for _, appId := range appIds {
		matcher, err := ToStrMatcher(appId)
		if err != nil {
			return nil, err
		}
		indexMatcher.Add(matcher)
	}
	if err := indexMatcher.Build(); err != nil {
		return nil, err
	}
	return &StrmatcherToAppSet{
		IndexMatcher: indexMatcher,
	}, nil
}

type loader interface {
	LoadIP(filename, country string) (*cgeo.GeoIP, error)
	LoadSite(filename, list string) (*cgeo.GeoSite, error)
	LoadDomainsClash(filename string) ([]*cgeo.Domain, error)
	LoadCidrsClash(filename string) ([]*cgeo.CIDR, error)
	LoadAppsClash(filename string) ([]*configs.AppId, error)
}

func AtomicDomainSetToIndexMatcher(atomicSet *configs.AtomicDomainSetConfig, l loader) (strmatcher.IndexMatcher, error) {
	var domains []*cgeo.Domain
	var err error
	if geosite := atomicSet.Geosite; geosite != nil {
		domains, err = GeositeConfigToGeoDomains(geosite, l)
		if err != nil {
			return nil, fmt.Errorf("geosite.ToGeoDomains failed: %w", err)
		}
	}
	if atomicSet.Domains != nil {
		for _, domain := range atomicSet.Domains {
			domains = append(domains, &cgeo.Domain{
				Type:  cgeo.Domain_Type(domain.Type),
				Value: domain.Value,
			})
		}
	}
	if atomicSet.ClashFiles != nil {
		for _, clashFile := range atomicSet.ClashFiles {
			values, err := l.LoadDomainsClash(clashFile)
			if err != nil {
				return nil, fmt.Errorf("failed to extract domains from clash file: %w", err)
			}
			domains = append(domains, values...)
		}
	}
	var opts []strmatcher.MphIndexMatcherOption
	if atomicSet.UseBloomFilter {
		opts = append(opts, strmatcher.WithSufficMatcherGroup(strmatcher.NewMatcherGroupBF(uint(len(domains)+1000))))
	}
	matcher, err := cgeo.ToMphIndexMatcher(domains, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to convert []*geo.Domain to matcher: %w", err)
	}
	// common.Log()
	return matcher, nil
}

func GeositeConfigToGeoDomains(c *configs.GeositeConfig, l loader) ([]*cgeo.Domain, error) {
	var geoDomains []*cgeo.Domain
	for _, code := range c.Codes {
		l, err := l.LoadSite(c.Filepath, code)
		if err != nil {
			return nil, err
		}
		// common.Log()
		geoDomains = append(geoDomains, l.Domain...)
	}

	// trim the attrList
	trimmedAttrList := make([]string, 0, len(c.Attributes))
	for _, attr := range c.Attributes {
		trimmedAttr := strings.ToLower(strings.TrimSpace(attr))
		if len(trimmedAttr) == 0 {
			continue
		}
		trimmedAttrList = append(trimmedAttrList, trimmedAttr)
	}
	if len(trimmedAttrList) == 0 {
		return geoDomains, nil
	} else {
		var domainsToBeUsed []*cgeo.Domain
		//TODO use a function
		// filter the domains
		filteredDomains := make([]*cgeo.Domain, 0, len(geoDomains))
		matched := false
		for _, domain := range geoDomains {
			for _, attribute := range domain.Attribute {
				for _, attr := range trimmedAttrList {
					if strings.EqualFold(attribute.GetKey(), attr) {
						filteredDomains = append(filteredDomains, domain)
						matched = true
						break
					}
				}
				if matched {
					matched = false
					break
				}
			}
		}
		domainsToBeUsed = append(domainsToBeUsed, filteredDomains...)
		return domainsToBeUsed, nil
	}
}

func GeoIpConfigToCidrs(config *configs.GeoIPConfig, l loader) ([]*cgeo.CIDR, error) {
	var cidrs []*cgeo.CIDR
	for _, code := range config.GetCodes() {
		l, err := l.LoadIP(config.Filepath, code)
		if err != nil {
			return nil, fmt.Errorf("failed to load geoip: %s", code)
		}
		cidrs = append(cidrs, l.Cidr...)
	}
	return cidrs, nil
}

func AtomicIpSetToIPMatcher(c *configs.AtomicIPSetConfig, l loader) (*cgeo.IPMatcher, error) {
	var cidrs []*cgeo.CIDR
	if geoip := c.Geoip; geoip != nil {
		for _, code := range geoip.GetCodes() {
			l, err := l.LoadIP(geoip.Filepath, code)
			if err != nil {
				return nil, fmt.Errorf("failed to load geoip: %s", code)
			}
			cidrs = append(cidrs, l.Cidr...)
		}
	}
	if c.Cidrs != nil {
		for _, cidr := range c.Cidrs {
			cidrs = append(cidrs, &cgeo.CIDR{
				Ip:     cidr.Ip,
				Prefix: cidr.Prefix,
			})
		}
	}
	if c.ClashFiles != nil {
		for _, clashFile := range c.ClashFiles {
			values, err := l.LoadCidrsClash(clashFile)
			if err != nil {
				return nil, fmt.Errorf("failed to extract cidrs from clash file: %w", err)
			}
			cidrs = append(cidrs, values...)
		}
	}
	m, err := cgeo.NewIPMatcherFromGeoCidrs(cidrs, c.Inverse)
	if err != nil {
		return nil, fmt.Errorf("failed to create ip matcher: %w", err)
	}
	return m, nil
}

func getGreatIPSet(c *configs.GreatIPSetConfig, ipSets map[string]i.IPSet) (*GreatIPSet, error) {
	var inMatchers []i.IPSet
	var exMatcher []i.IPSet
	for _, n := range c.InNames {
		if ipSets[n] == nil {
			return nil, fmt.Errorf("ip matcher not found: %s", n)
		}
		inMatchers = append(inMatchers, ipSets[n])
	}
	for _, n := range c.ExNames {
		if ipSets[n] == nil {
			return nil, fmt.Errorf("ip matcher not found: %s", n)
		}
		exMatcher = append(exMatcher, ipSets[n])
	}
	ipMatcher := NewGreatIPSet(inMatchers, exMatcher)

	return ipMatcher, nil
}

func getGreatDomainSet(c *configs.GreatDomainSetConfig, geo *Geo) (*GreatDomainSet, error) {
	for _, n := range c.InNames {
		if n == c.Name {
			return nil, errors.New("a great domain set cannot contain itself")
		}
	}
	for _, n := range c.ExNames {
		if n == c.Name {
			return nil, errors.New("a great domain set cannot contain itself")
		}
	}
	dm := NewGreatDomainSet(c.InNames, c.ExNames, geo)

	return dm, nil
}
