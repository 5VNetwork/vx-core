package clashconfig

import (
	"bufio"
	"bytes"
	"io"
	"net/netip"
	"strings"

	commongeo "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/common/geo"
	vxrouter "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/router"
	"gopkg.in/yaml.v3"
)

type PayloadConfig struct {
	Payload []string `yaml:"payload"`
}

// ExtractDomainsFromClashRules parses files containing domain rules and extracts commongeo.Domain entries.
// It supports both plain text format (DOMAIN and DOMAIN-SUFFIX rules) and YAML format with payload array.
func ExtractDomainsFromClashRules(reader io.Reader) ([]*commongeo.Domain, error) {
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

// parseDomainRule parses a single domain rule string and returns a commongeo.Domain if valid
func parseDomainRule(rule string) *commongeo.Domain {
	rule = strings.TrimSpace(rule)

	// Skip comments and empty lines
	if rule == "" || strings.HasPrefix(rule, "#") {
		return nil
	}

	// Parse DOMAIN rules
	if strings.HasPrefix(rule, "DOMAIN,") {
		value := strings.TrimPrefix(rule, "DOMAIN,")
		return &commongeo.Domain{
			Type:  commongeo.Domain_Full,
			Value: value,
		}
	} else if strings.HasPrefix(rule, "DOMAIN-SUFFIX,") {
		value := strings.TrimPrefix(rule, "DOMAIN-SUFFIX,")
		return &commongeo.Domain{
			Type:  commongeo.Domain_RootDomain,
			Value: value,
		}
	} else if strings.HasPrefix(rule, "DOMAIN-KEYWORD,") {
		value := strings.TrimPrefix(rule, "DOMAIN-KEYWORD,")
		return &commongeo.Domain{
			Type:  commongeo.Domain_Plain,
			Value: value,
		}
	} else if strings.HasPrefix(rule, ".") {
		return &commongeo.Domain{
			Type:  commongeo.Domain_Plain,
			Value: rule,
		}
	} else if strings.HasPrefix(rule, "+") {
		return &commongeo.Domain{
			Type:  commongeo.Domain_RootDomain,
			Value: strings.TrimPrefix(rule, "+."),
		}
	} else if strings.Contains(rule, "*") {
		// Handle any wildcard pattern like *.domain.com, *.*.domain.com, a.*.domain.com
		regexPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(rule, ".", "\\."), "*", "[^.]+") + "$"
		return &commongeo.Domain{
			Type:  commongeo.Domain_Regex,
			Value: regexPattern,
		}
	} else if !strings.Contains(rule, ",") && !strings.Contains(rule, "/") {
		return &commongeo.Domain{
			Type:  commongeo.Domain_Plain,
			Value: rule,
		}
	}
	return nil
}

// parseYAMLDomainFormat parses YAML content with payload array
func parseYAMLDomainFormat(content []byte) ([]*commongeo.Domain, error) {
	var config PayloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var domains []*commongeo.Domain
	for _, rule := range config.Payload {
		if domain := parseDomainRule(rule); domain != nil {
			domains = append(domains, domain)
		}
	}

	return domains, nil
}

// parsePlainTextDomainFormat parses plain text format with line-by-line rules
func parsePlainTextDomainFormat(content []byte) ([]*commongeo.Domain, error) {
	var domains []*commongeo.Domain
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		if domain := parseDomainRule(scanner.Text()); domain != nil {
			domains = append(domains, domain)
		}
	}

	return domains, scanner.Err()
}

// parseCidrRule parses a single CIDR rule string
func parseCidrRule(rule string) (*commongeo.CIDR, error) {
	rule = strings.TrimSpace(rule)

	// Skip comments and empty lines
	if rule == "" || strings.HasPrefix(rule, "#") {
		return nil, nil
	}

	if !strings.Contains(rule, "/") {
		return nil, nil
	}

	if strings.Contains(rule, ",") {
		parts := strings.Split(rule, ",")
		if len(parts) < 2 {
			return nil, nil
		}

		if parts[0] == "IP-CIDR" || parts[0] == "IP-CIDR6" {
			cidr := parts[1]
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				return nil, err
			}
			return &commongeo.CIDR{
				Ip:     prefix.Addr().AsSlice(),
				Prefix: uint32(prefix.Bits()),
			}, nil
		}
	} else {
		prefix, err := netip.ParsePrefix(rule)
		if err != nil {
			return nil, err
		}
		return &commongeo.CIDR{
			Ip:     prefix.Addr().AsSlice(),
			Prefix: uint32(prefix.Bits()),
		}, nil
	}

	return nil, nil
}

// parseYAMLCidrFormat parses YAML content with payload array
func parseYAMLCidrFormat(content []byte) ([]*commongeo.CIDR, error) {
	var config PayloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var cidrs []*commongeo.CIDR
	for _, rule := range config.Payload {
		cidr, err := parseCidrRule(rule)
		if err != nil {
			return nil, err
		}
		if cidr != nil {
			cidrs = append(cidrs, cidr)
		}
	}

	return cidrs, nil
}

// parsePlainTextCidrFormat parses plain text format with line-by-line rules
func parsePlainTextCidrFormat(content []byte) ([]*commongeo.CIDR, error) {
	var cidrs []*commongeo.CIDR
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		cidr, err := parseCidrRule(scanner.Text())
		if err != nil {
			return nil, err
		}
		if cidr != nil {
			cidrs = append(cidrs, cidr)
		}
	}

	return cidrs, scanner.Err()
}

// ExtractCidrFromClashRules parses files containing CIDR rules and extracts commongeo.CIDR entries.
// It supports both plain text format (IP-CIDR and IP-CIDR6 rules) and YAML format with payload array.
func ExtractCidrFromClashRules(reader io.Reader) ([]*commongeo.CIDR, error) {
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
func parseAppRule(rule string) *vxrouter.AppId {
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
		return &vxrouter.AppId{
			Value: app,
			Type:  vxrouter.AppId_Keyword,
		}
	} else if parts[0] == "PROCESS-PATH" {
		return &vxrouter.AppId{
			Value: app,
			Type:  vxrouter.AppId_Prefix,
		}
	}

	return nil
}

// parseYAMLAppFormat parses YAML content with payload array
func parseYAMLAppFormat(content []byte) ([]*vxrouter.AppId, error) {
	var config PayloadConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	var apps []*vxrouter.AppId
	for _, rule := range config.Payload {
		if app := parseAppRule(rule); app != nil {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

// parsePlainTextAppFormat parses plain text format with line-by-line rules
func parsePlainTextAppFormat(content []byte) ([]*vxrouter.AppId, error) {
	var apps []*vxrouter.AppId
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
func ExtractAppsFromClashRules(reader io.Reader) ([]*vxrouter.AppId, error) {
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
