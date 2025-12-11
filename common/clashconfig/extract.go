package clashconfig

import (
	"bufio"
	"bytes"
	"io"
	"net/netip"
	"strings"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/geo"
	"gopkg.in/yaml.v3"
)

type PayloadConfig struct {
	Payload []string `yaml:"payload"`
}

// ExtractDomainsFromClashRules parses files containing domain rules and extracts geo.Domain entries.
// It supports both plain text format (DOMAIN and DOMAIN-SUFFIX rules) and YAML format with payload array.
func ExtractDomainsFromClashRules(reader io.Reader) ([]*geo.Domain, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Try to parse as YAML first
	if domains, err := parseYAMLDomainFormat(content); err == nil && len(domains) > 0 {
		return domains, nil
	}

	// Fallback to plain text format
	return parsePlainTextDomainFormat(content)
}

// parseDomainRule parses a single domain rule string and returns a geo.Domain if valid
func parseDomainRule(rule string) *geo.Domain {
	rule = strings.TrimSpace(rule)

	// Skip comments and empty lines
	if rule == "" || strings.HasPrefix(rule, "#") {
		return nil
	}

	// Parse DOMAIN rules
	if strings.HasPrefix(rule, "DOMAIN,") {
		value := strings.TrimPrefix(rule, "DOMAIN,")
		return &geo.Domain{
			Type:  geo.Domain_Full,
			Value: value,
		}
	} else if strings.HasPrefix(rule, "DOMAIN-SUFFIX,") {
		value := strings.TrimPrefix(rule, "DOMAIN-SUFFIX,")
		return &geo.Domain{
			Type:  geo.Domain_RootDomain,
			Value: value,
		}
	} else if strings.HasPrefix(rule, "DOMAIN-KEYWORD,") {
		value := strings.TrimPrefix(rule, "DOMAIN-KEYWORD,")
		return &geo.Domain{
			Type:  geo.Domain_Plain,
			Value: value,
		}
	} else if strings.HasPrefix(rule, ".") {
		return &geo.Domain{
			Type:  geo.Domain_Plain,
			Value: rule,
		}
	} else if strings.HasPrefix(rule, "+") {
		return &geo.Domain{
			Type:  geo.Domain_RootDomain,
			Value: strings.TrimPrefix(rule, "+."),
		}
	} else if strings.Contains(rule, "*") {
		// Handle any wildcard pattern like *.domain.com, *.*.domain.com, a.*.domain.com
		regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(rule, ".", "\\."), "*", "[^.]+") + "$"
		return &geo.Domain{
			Type:  geo.Domain_Regex,
			Value: regexPattern,
		}
	} else if !strings.Contains(rule, ",") && !strings.Contains(rule, "/") {
		return &geo.Domain{
			Type:  geo.Domain_Plain,
			Value: rule,
		}
	}
	return nil
}

// parseYAMLDomainFormat parses YAML content with payload array
func parseYAMLDomainFormat(content []byte) ([]*geo.Domain, error) {
	var config PayloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var domains []*geo.Domain
	for _, rule := range config.Payload {
		if domain := parseDomainRule(rule); domain != nil {
			domains = append(domains, domain)
		}
	}

	return domains, nil
}

// parsePlainTextDomainFormat parses plain text format with line-by-line rules
func parsePlainTextDomainFormat(content []byte) ([]*geo.Domain, error) {
	var domains []*geo.Domain
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		if domain := parseDomainRule(scanner.Text()); domain != nil {
			domains = append(domains, domain)
		}
	}

	return domains, scanner.Err()
}

// parseCidrRule parses a single CIDR rule string and returns a geo.CIDR if valid
func parseCidrRule(rule string) *geo.CIDR {
	rule = strings.TrimSpace(rule)

	// Skip comments and empty lines
	if rule == "" || strings.HasPrefix(rule, "#") {
		return nil
	}

	if !strings.Contains(rule, "/") {
		return nil
	}

	if strings.Contains(rule, ",") {
		parts := strings.Split(rule, ",")
		if len(parts) < 2 {
			return nil
		}

		if parts[0] == "IP-CIDR" || parts[0] == "IP-CIDR6" {
			cidr := parts[1]
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				return nil
			}
			return &geo.CIDR{
				Ip:     prefix.Addr().AsSlice(),
				Prefix: uint32(prefix.Bits()),
			}
		}
	} else {
		prefix, err := netip.ParsePrefix(rule)
		if err != nil {
			return nil
		}
		return &geo.CIDR{
			Ip:     prefix.Addr().AsSlice(),
			Prefix: uint32(prefix.Bits()),
		}
	}

	return nil
}

// parseYAMLCidrFormat parses YAML content with payload array
func parseYAMLCidrFormat(content []byte) ([]*geo.CIDR, error) {
	var config PayloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var cidrs []*geo.CIDR
	for _, rule := range config.Payload {
		if cidr := parseCidrRule(rule); cidr != nil {
			cidrs = append(cidrs, cidr)
		}
	}

	return cidrs, nil
}

// parsePlainTextCidrFormat parses plain text format with line-by-line rules
func parsePlainTextCidrFormat(content []byte) ([]*geo.CIDR, error) {
	var cidrs []*geo.CIDR
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		if cidr := parseCidrRule(scanner.Text()); cidr != nil {
			cidrs = append(cidrs, cidr)
		}
	}

	return cidrs, scanner.Err()
}

// ExtractCidrFromClashRules parses files containing CIDR rules and extracts geo.CIDR entries.
// It supports both plain text format (IP-CIDR and IP-CIDR6 rules) and YAML format with payload array.
func ExtractCidrFromClashRules(reader io.Reader) ([]*geo.CIDR, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Try to parse as YAML first
	if cidrs, err := parseYAMLCidrFormat(content); err == nil && len(cidrs) > 0 {
		return cidrs, nil
	}

	// Fallback to plain text format
	return parsePlainTextCidrFormat(content)
}

// parseAppRule parses a single app rule string and returns a configs.AppId if valid
func parseAppRule(rule string) *configs.AppId {
	rule = strings.TrimSpace(rule)

	// Skip comments and empty lines
	if rule == "" || strings.HasPrefix(rule, "#") {
		return nil
	}

	parts := strings.Split(rule, ",")
	if len(parts) != 2 {
		return nil
	}

	app := parts[1]
	if parts[0] == "PROCESS-NAME" {
		return &configs.AppId{
			Value: app,
			Type:  configs.AppId_Keyword,
		}
	} else if parts[0] == "PROCESS-PATH" {
		return &configs.AppId{
			Value: app,
			Type:  configs.AppId_Prefix,
		}
	}

	return nil
}

// parseYAMLAppFormat parses YAML content with payload array
func parseYAMLAppFormat(content []byte) ([]*configs.AppId, error) {
	var config PayloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var apps []*configs.AppId
	for _, rule := range config.Payload {
		if app := parseAppRule(rule); app != nil {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

// parsePlainTextAppFormat parses plain text format with line-by-line rules
func parsePlainTextAppFormat(content []byte) ([]*configs.AppId, error) {
	var apps []*configs.AppId
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		if app := parseAppRule(scanner.Text()); app != nil {
			apps = append(apps, app)
		}
	}

	return apps, scanner.Err()
}

// ExtractAppsFromClashRules parses files containing app rules and extracts configs.AppId entries.
// It supports both plain text format (PROCESS-NAME and PROCESS-PATH rules) and YAML format with payload array.
func ExtractAppsFromClashRules(reader io.Reader) ([]*configs.AppId, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Try to parse as YAML first
	if apps, err := parseYAMLAppFormat(content); err == nil && len(apps) > 0 {
		return apps, nil
	}

	// Fallback to plain text format
	return parsePlainTextAppFormat(content)
}
