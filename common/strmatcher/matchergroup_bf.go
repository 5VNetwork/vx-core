package strmatcher

import "github.com/bits-and-blooms/bloom/v3"

type MatcherGroupBF struct {
	num uint
	bf  *bloom.BloomFilter
}

func NewMatcherGroupBF(num uint) *MatcherGroupBF {
	m := &MatcherGroupBF{
		bf: bloom.NewWithEstimates(num, 0.01),
	}
	return m
}

func (g *MatcherGroupBF) AddDomainMatcher(matcher DomainMatcher, value uint32) {
	g.bf = g.bf.Add([]byte(matcher.Pattern()))
}

func (g *MatcherGroupBF) AddFullMatcher(matcher FullMatcher, value uint32) {
	g.bf = g.bf.Add([]byte(matcher.Pattern()))
}

func (g *MatcherGroupBF) Build() error {

	return nil
}

func (g *MatcherGroupBF) Match(input string) []uint32 {
	return nil
}

func (g *MatcherGroupBF) MatchAny(input string) bool {
	if g.bf == nil || len(input) == 0 || input[len(input)-1] == '.' {
		return false
	}

	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == '.' && i+1 < len(input) {
			if g.bf.Test([]byte(input[i+1:])) {
				return true
			}
		}
	}

	if g.bf.Test([]byte(input)) {
		return true
	}

	return false
}
