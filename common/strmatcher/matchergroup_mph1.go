package strmatcher

// import (
// 	"strings"
// 	// "github.com/alecthomas/mph"
// 	"github.com/rs/zerolog/log"
// )

// type MatcherGroupMph1 struct {
// 	strings []string
// 	table   *Table
// }

// func NewMatcherGroupMph1(num int) *MatcherGroupMph1 {
// 	m := &MatcherGroupMph1{
// 		strings: make([]string, 0, num),
// 	}
// 	return m
// }

// // AddDomainMatcher implements MatcherGroupForDomain.
// func (g *MatcherGroupMph1) AddDomainMatcher(matcher DomainMatcher, value uint32) {
// 	pattern := strings.ToLower(matcher.Pattern())
// 	g.strings = append(g.strings, pattern)
// }

// func (g *MatcherGroupMph1) AddFullMatcher(matcher FullMatcher, value uint32) {
// 	pattern := strings.ToLower(matcher.Pattern())
// 	g.strings = append(g.strings, pattern)
// }

// func (g *MatcherGroupMph1) Build() error {
// 	log.Debug().Int("num", len(g.strings)).Msg("Building mph1")
// 	g.table = Build(g.strings)
// 	g.strings = nil
// 	return nil
// }

// // Match implements MatcherGroup.Match.
// func (g *MatcherGroupMph1) Match(input string) []uint32 {
// 	return nil
// }

// func (g *MatcherGroupMph1) MatchAny(input string) bool {
// 	if g.table == nil || len(input) == 0 || input[len(input)-1] == '.' {
// 		return false
// 	}

// 	for i := len(input) - 1; i >= 0; i-- {
// 		if input[i] == '.' && i+1 < len(input) {
// 			if _, found := g.table.Lookup(input[i+1:]); found {
// 				return true
// 			}
// 		}
// 	}

// 	if _, found := g.table.Lookup(input); found {
// 		return true
// 	}

// 	return false
// }
