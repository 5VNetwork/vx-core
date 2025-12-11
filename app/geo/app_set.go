package geo

import (
	"errors"
	"strings"

	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/strmatcher"
	"github.com/5vnetwork/vx-core/i"
)

type StrmatcherToAppSet struct {
	strmatcher.IndexMatcher
}

func (m *StrmatcherToAppSet) Match(appId string) bool {
	return m.MatchAny(strings.ToLower(appId))
}

type AppSet struct {
	m    strmatcher.IndexMatcher
	tags []string
	h    i.GeoHelper
}

func NewAppSet(tags []string, h i.GeoHelper, appIds ...*configs.AppId) (*AppSet, error) {
	a := &AppSet{
		tags: tags,
		h:    h,
	}
	if len(appIds) > 0 {
		indexMatcher := strmatcher.NewMphIndexMatcher()
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
		a.m = indexMatcher
	}

	return a, nil
}

func (a *AppSet) Match(appId string) bool {
	appId = strings.ToLower(appId)
	if a.m != nil && a.m.MatchAny(appId) {
		return true
	}
	if a.h != nil {
		for _, tag := range a.tags {
			if a.h.MatchAppId(appId, tag) {
				return true
			}
		}
	}
	return false
}

func ToStrMatcher(d *configs.AppId) (strmatcher.Matcher, error) {
	lowerValue := strings.ToLower(d.Value)
	switch d.Type {
	case configs.AppId_Exact:
		return strmatcher.Full.New(lowerValue)
	case configs.AppId_Prefix:
		return strmatcher.Prefix.New(lowerValue)
	case configs.AppId_Keyword:
		return strmatcher.Substr.New(lowerValue)
	default:
		return nil, errors.New("unknown domain type")
	}
}
