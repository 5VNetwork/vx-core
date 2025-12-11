package strmatcher_test

// import (
// 	"runtime"
// 	"testing"

// 	"github.com/5vnetwork/vx-core/app/configs"
// 	appgeo "github.com/5vnetwork/vx-core/app/geo"
// 	"github.com/5vnetwork/vx-core/common"
// 	"github.com/5vnetwork/vx-core/common/geo/stdloader"
// 	. "github.com/5vnetwork/vx-core/common/strmatcher"
// 	"github.com/5vnetwork/vx-core/test"
// 	"github.com/stretchr/testify/assert"
// )

// func TestMphMatcherGroup1(t *testing.T) {
// 	cases1 := []struct {
// 		pattern string
// 		mType   Type
// 		input   string
// 		output  bool
// 	}{
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Domain,
// 			input:   "www.v2fly.org",
// 			output:  true,
// 		},
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Domain,
// 			input:   "v2fly.org",
// 			output:  true,
// 		},
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Domain,
// 			input:   "www.v3fly.org",
// 			output:  false,
// 		},
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Domain,
// 			input:   "2fly.org",
// 			output:  false,
// 		},
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Domain,
// 			input:   "xv2fly.org",
// 			output:  false,
// 		},
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Full,
// 			input:   "v2fly.org",
// 			output:  true,
// 		},
// 		{
// 			pattern: "v2fly.org",
// 			mType:   Full,
// 			input:   "xv2fly.org",
// 			output:  false,
// 		},
// 	}
// 	for _, test := range cases1 {
// 		mph := NewMatcherGroupMph1(100)
// 		matcher, err := test.mType.New(test.pattern)
// 		common.Must(err)
// 		common.Must(AddMatcherToGroup(mph, matcher, 0))
// 		mph.Build()
// 		if m := mph.MatchAny(test.input); m != test.output {
// 			t.Error("unexpected output: ", m, " for test case ", test)
// 		}
// 	}
// 	{
// 		cases2Input := []struct {
// 			pattern string
// 			mType   Type
// 		}{
// 			{
// 				pattern: "163.com",
// 				mType:   Domain,
// 			},
// 			{
// 				pattern: "m.126.com",
// 				mType:   Full,
// 			},
// 			{
// 				pattern: "3.com",
// 				mType:   Full,
// 			},
// 		}
// 		mph := NewMatcherGroupMph1(100)
// 		for _, test := range cases2Input {
// 			matcher, err := test.mType.New(test.pattern)
// 			common.Must(err)
// 			common.Must(AddMatcherToGroup(mph, matcher, 0))
// 		}
// 		mph.Build()
// 		cases2Output := []struct {
// 			pattern string
// 			res     bool
// 		}{
// 			{
// 				pattern: "126.com",
// 				res:     false,
// 			},
// 			{
// 				pattern: "m.163.com",
// 				res:     true,
// 			},
// 			{
// 				pattern: "mm163.com",
// 				res:     false,
// 			},
// 			{
// 				pattern: "m.126.com",
// 				res:     true,
// 			},
// 			{
// 				pattern: "163.com",
// 				res:     true,
// 			},
// 			{
// 				pattern: "63.com",
// 				res:     false,
// 			},
// 			{
// 				pattern: "oogle.com",
// 				res:     false,
// 			},
// 			{
// 				pattern: "vvgoogle.com",
// 				res:     false,
// 			},
// 		}
// 		for _, test := range cases2Output {
// 			if m := mph.MatchAny(test.pattern); m != test.res {
// 				t.Error("unexpected output: ", m, " for test case ", test)
// 			}
// 		}
// 	}
// 	{
// 		cases3Input := []struct {
// 			pattern string
// 			mType   Type
// 		}{
// 			{
// 				pattern: "video.google.com",
// 				mType:   Domain,
// 			},
// 			{
// 				pattern: "gle.com",
// 				mType:   Domain,
// 			},
// 		}
// 		mph := NewMatcherGroupMph1(100)
// 		for _, test := range cases3Input {
// 			matcher, err := test.mType.New(test.pattern)
// 			common.Must(err)
// 			common.Must(AddMatcherToGroup(mph, matcher, 0))
// 		}
// 		mph.Build()
// 		cases3Output := []struct {
// 			pattern string
// 			res     bool
// 		}{
// 			{
// 				pattern: "google.com",
// 				res:     false,
// 			},
// 		}
// 		for _, test := range cases3Output {
// 			if m := mph.MatchAny(test.pattern); m != test.res {
// 				t.Error("unexpected output: ", m, " for test case ", test)
// 			}
// 		}
// 	}
// }

// func TestA(t *testing.T) {
// 	test.InitZeroLog()
// 	l := stdloader.NewStandartLoader()
// 	i, err := appgeo.AtomicDomainSetToIndexMatcher(&configs.AtomicDomainSetConfig{
// 		Geosite: &configs.GeositeConfig{
// 			Codes:    []string{"cn", "apple-cn", "google-cn", "tld-cn", "private", "category-games"},
// 			Filepath: "../../infra/geo/simplified_geosite.dat",
// 		},
// 	}, l)
// 	common.Must(err)

// 	l = nil
// 	common.Log()
// 	runtime.GC()
// 	common.Log()

// 	assert.Equal(t, true, i.MatchAny("cn"))
// 	assert.Equal(t, true, i.MatchAny("baidu.com"))
// 	assert.Equal(t, true, i.MatchAny("xiaomi.com"))
// 	assert.Equal(t, true, i.MatchAny("bilibili.com"))
// 	assert.Equal(t, true, i.MatchAny("v2ray.com.cn"))
// 	assert.Equal(t, true, i.MatchAny("douyin.com"))
// 	assert.Equal(t, false, i.MatchAny("google.com"))
// 	assert.Equal(t, false, i.MatchAny("x.org"))
// 	assert.Equal(t, false, i.MatchAny("facebook.com"))
// 	common.Log()
// }
