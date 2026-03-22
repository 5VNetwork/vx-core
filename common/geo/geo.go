package geo

import (
	"errors"

	commongeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	"github.com/5vnetwork/vx-core/common/strmatcher"
)

type (
	GeoIP   = commongeo.GeoIP
	GeoSite = commongeo.GeoSite
	Domain  = commongeo.Domain
	CIDR    = commongeo.CIDR
)

func ToStrMatcher(d *commongeo.Domain) (strmatcher.Matcher, error) {
	switch d.Type {
	case commongeo.Domain_Full:
		return strmatcher.Full.New(d.Value)
	case commongeo.Domain_RootDomain:
		return strmatcher.Domain.New(d.Value)
	case commongeo.Domain_Plain:
		return strmatcher.Substr.New(d.Value)
	case commongeo.Domain_Regex:
		return strmatcher.Regex.New(d.Value)
	default:
		return nil, errors.New("unknown domain type")
	}
}

func ToMphIndexMatcher(domainMatchings []*commongeo.Domain,
	opts ...strmatcher.MphIndexMatcherOption) (strmatcher.IndexMatcher, error) {
	indexMatcher := strmatcher.NewMphIndexMatcher(opts...)
	for _, d := range domainMatchings {
		matcher, err := ToStrMatcher(d)
		if err != nil {
			return nil, err
		}
		indexMatcher.Add(matcher)
	}
	if err := indexMatcher.Build(); err != nil {
		return nil, err
	}
	return indexMatcher, nil
}
