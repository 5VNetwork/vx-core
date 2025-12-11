package strmatcher

// import (
// 	"strings"

// 	"github.com/aelaguiz/mph"
// )

// type MatcherGroupMph0 struct {
// 	builder *mph.CHDBuilder
// 	chd     *mph.CHD
// }

// var onlyValue = []byte{1}

// func NewMatcherGroupMph0() *MatcherGroupMph0 {
// 	m := &MatcherGroupMph0{
// 		builder: mph.Builder(),
// 	}
// 	return m
// }

// // AddDomainMatcher implements MatcherGroupForDomain.
// func (g *MatcherGroupMph0) AddDomainMatcher(matcher DomainMatcher, value uint32) {
// 	pattern := strings.ToLower(matcher.Pattern())
// 	g.builder.Add([]byte(pattern), onlyValue)
// }

// func (g *MatcherGroupMph0) AddFullMatcher(matcher FullMatcher, value uint32) {
// 	pattern := strings.ToLower(matcher.Pattern())
// 	g.builder.Add([]byte(pattern), onlyValue)
// }

// func (g *MatcherGroupMph0) Build() error {
// 	chd, err := g.builder.Build()
// 	if err != nil {
// 		return err
// 	}
// 	g.chd = chd
// 	g.builder = nil
// 	return nil
// }

// // Match implements MatcherGroup.Match.
// func (g *MatcherGroupMph0) Match(input string) []uint32 {
// 	return nil
// }

// func (g *MatcherGroupMph0) MatchAny(input string) bool {
// 	if g.chd == nil || len(input) == 0 || input[len(input)-1] == '.' {
// 		return false
// 	}

// 	for i := len(input) - 1; i >= 0; i-- {
// 		if input[i] == '.' && i+1 < len(input) {
// 			if b := g.chd.Get([]byte(input[i+1:])); b != nil {
// 				return true
// 			}
// 		}
// 	}

// 	if b := g.chd.Get([]byte(input)); b != nil {
// 		return true
// 	}

// 	return false
// }
