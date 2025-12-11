package router

import (
	"context"

	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/i"
)

// case insensitive
// type AppIdMatcher struct {
// 	strmatcher.IndexMatcher
// }

// func (m *AppIdMatcher) Apply(ctx SessionInfo) bool {
// 	appId := ctx.GetAppId()
// 	if appId == "" {
// 		return false
// 	}
// 	return m.MatchAny(strings.ToLower(appId))
// }

//	func ToIndexMatcher(appIds []*configs.AppId) (strmatcher.IndexMatcher, error) {
//		indexMatcher := strmatcher.NewMphIndexMatcher()
//		for _, appId := range appIds {
//			matcher, err := ToStrMatcher(appId)
//			if err != nil {
//				return nil, err
//			}
//			indexMatcher.Add(matcher)
//		}
//		if err := indexMatcher.Build(); err != nil {
//			return nil, err
//		}
//		return indexMatcher, nil
//	}
//
//	func ToStrMatcher(d *configs.AppId) (strmatcher.Matcher, error) {
//		lowerValue := strings.ToLower(d.Value)
//		switch d.Type {
//		case configs.AppId_Exact:
//			return strmatcher.Full.New(lowerValue)
//		case configs.AppId_Prefix:
//			return strmatcher.Prefix.New(lowerValue)
//		case configs.AppId_Keyword:
//			return strmatcher.Substr.New(lowerValue)
//		default:
//			return nil, errors.New("unknown domain type")
//		}
//	}
type AppIdMatcher struct {
	AppSet i.AppSet
}

func (m *AppIdMatcher) Apply(c context.Context, info *session.Info, rw interface{}) (interface{}, bool) {
	appId := info.GetAppId()
	if appId == "" {
		return rw, false
	}
	return rw, m.AppSet.Match(appId)
}
