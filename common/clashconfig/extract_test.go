package clashconfig_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/5vnetwork/vx-core/common/clashconfig"
	"github.com/5vnetwork/vx-core/common/geo"
)

func TestExtractDomainsFromClashRulesYaml1(t *testing.T) {
	file, err := os.Open("1.yaml")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	domains, err := clashconfig.ExtractDomainsFromClashRules(file)
	if err != nil {
		t.Fatalf("failed to extract domains: %v", err)
	}
	fmt.Println(domains)

	if len(domains) != 4 {
		t.Fatalf("expected 4 domains, got %d", len(domains))
	}

	m, err := geo.ToMphIndexMatcher(domains)
	if err != nil {
		t.Fatalf("failed to create mph index matcher: %v", err)
	}

	if !m.MatchAny("a.com") {
		t.Fatalf("failed to match a.com")
	}
	if !m.MatchAny("b.com") {
		t.Fatalf("failed to match b.com")
	}
	if m.MatchAny("c.com") {
		t.Fatalf("should not match c.com")
	}
	if !m.MatchAny("a.b.com") {
		t.Fatalf("failed to match a.b.com")
	}
	if !m.MatchAny("a.b.c.d.com") {
		t.Fatalf("failed to match a.b.c.d.com")
	}
	if m.MatchAny("a.c.d.com") {
		t.Fatalf("should not match a.c.d.com")
	}
	if m.MatchAny("e.a.b.c.d.com") {
		t.Fatalf("should not match e.a.b.c.d.com")
	}
	if m.MatchAny("e.com") {
		t.Fatalf("should not match e.com")
	}
}

func TestExtractDomainsFromClashRulesYaml2(t *testing.T) {
	file, err := os.Open("2.yaml")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	domains, err := clashconfig.ExtractDomainsFromClashRules(file)
	if err != nil {
		t.Fatalf("failed to extract domains: %v", err)
	}
	fmt.Println(domains)

	if len(domains) != 3 {
		t.Fatalf("expected 4 domains, got %d", len(domains))
	}

	m, err := geo.ToMphIndexMatcher(domains)
	if err != nil {
		t.Fatalf("failed to create mph index matcher: %v", err)
	}

	if !m.MatchAny("voice.telephony.goog") {
		t.Fatalf("failed to match voice.telephony.goog")
	}
	if !m.MatchAny("0emm.com") {
		t.Fatalf("failed to match 0emm.com")
	}
	if !m.MatchAny("appspot.com") {
		t.Fatalf("failed to match appspot.com")
	}
}

func TestExtractDomainsFromClashRules3(t *testing.T) {
	file, err := os.Open("3.list")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	domains, err := clashconfig.ExtractDomainsFromClashRules(file)
	if err != nil {
		t.Fatalf("failed to extract domains: %v", err)
	}
	fmt.Println(domains)

	if len(domains) != 3 {
		t.Fatalf("expected 4 domains, got %d", len(domains))
	}

	m, err := geo.ToMphIndexMatcher(domains)
	if err != nil {
		t.Fatalf("failed to create mph index matcher: %v", err)
	}

	if !m.MatchAny("voice.telephony.goog") {
		t.Fatalf("failed to match voice.telephony.goog")
	}
	if !m.MatchAny("0emm.com") {
		t.Fatalf("failed to match 0emm.com")
	}
	if !m.MatchAny("appspot.com") {
		t.Fatalf("failed to match appspot.com")
	}
}

func TestExtractCidrsFromClashRules1(t *testing.T) {
	file, err := os.Open("1.yaml")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	cidrs, err := clashconfig.ExtractCidrFromClashRules(file)
	if err != nil {
		t.Fatalf("failed to extract domains: %v", err)
	}

	if len(cidrs) != 2 {
		t.Fatalf("expected 2 cidrs, got %d", len(cidrs))
	}

}

func TestExtractCidrsFromClashRules2(t *testing.T) {
	file, err := os.Open("2.yaml")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	cidrs, err := clashconfig.ExtractCidrFromClashRules(file)
	if err != nil {
		t.Fatalf("failed to extract domains: %v", err)
	}

	if len(cidrs) != 2 {
		t.Fatalf("expected 2 cidrs, got %d", len(cidrs))
	}

}

func TestExtractCidrsFromClashRules3(t *testing.T) {
	file, err := os.Open("3.list")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	cidrs, err := clashconfig.ExtractCidrFromClashRules(file)
	if err != nil {
		t.Fatalf("failed to extract domains: %v", err)
	}

	if len(cidrs) != 2 {
		t.Fatalf("expected 2 cidrs, got %d", len(cidrs))
	}
}
