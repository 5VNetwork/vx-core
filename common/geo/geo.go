package geo

import (
	"errors"

	"github.com/5vnetwork/vx-core/common/strmatcher"
)

// type loader interface {
// 	LoadIP(filename, country string) ([]*CIDR, error)
// 	LoadSite(filename, list string) ([]*Domain, error)
// }

func ToStrMatcher(d *Domain) (strmatcher.Matcher, error) {
	switch d.Type {
	case Domain_Full:
		return strmatcher.Full.New(d.Value)
	case Domain_RootDomain:
		return strmatcher.Domain.New(d.Value)
	case Domain_Plain:
		return strmatcher.Substr.New(d.Value)
	case Domain_Regex:
		return strmatcher.Regex.New(d.Value)
	default:
		return nil, errors.New("unknown domain type")
	}
}

func ToMphIndexMatcher(domainMatchings []*Domain, opts ...strmatcher.MphIndexMatcherOption) (strmatcher.IndexMatcher, error) {
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
