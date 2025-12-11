package geo

import (
	"errors"
	"testing"

	"github.com/5vnetwork/vx-core/app/configs"
	cgeo "github.com/5vnetwork/vx-core/common/geo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLoader implements the loader interface for testing
type MockLoader struct {
	loadIPFunc           func(filename, country string) (*cgeo.GeoIP, error)
	loadSiteFunc         func(filename, list string) (*cgeo.GeoSite, error)
	loadDomainsClashFunc func(filename string) ([]*cgeo.Domain, error)
	loadCidrsClashFunc   func(filename string) ([]*cgeo.CIDR, error)
	loadAppsClashFunc    func(filename string) ([]*configs.AppId, error)
}

func (m *MockLoader) LoadIP(filename, country string) (*cgeo.GeoIP, error) {
	if m.loadIPFunc != nil {
		return m.loadIPFunc(filename, country)
	}
	return &cgeo.GeoIP{}, nil
}

func (m *MockLoader) LoadSite(filename, list string) (*cgeo.GeoSite, error) {
	if m.loadSiteFunc != nil {
		return m.loadSiteFunc(filename, list)
	}
	return &cgeo.GeoSite{}, nil
}

func (m *MockLoader) LoadDomainsClash(filename string) ([]*cgeo.Domain, error) {
	if m.loadDomainsClashFunc != nil {
		return m.loadDomainsClashFunc(filename)
	}
	return []*cgeo.Domain{}, nil
}

func (m *MockLoader) LoadCidrsClash(filename string) ([]*cgeo.CIDR, error) {
	if m.loadCidrsClashFunc != nil {
		return m.loadCidrsClashFunc(filename)
	}
	return []*cgeo.CIDR{}, nil
}

func (m *MockLoader) LoadAppsClash(filename string) ([]*configs.AppId, error) {
	if m.loadAppsClashFunc != nil {
		return m.loadAppsClashFunc(filename)
	}
	return []*configs.AppId{}, nil
}

// =============================================================================
// NewGeo Tests
// =============================================================================

func TestNewGeo_EmptyConfig(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.NotNil(t, geo.OppositeDomainTags)
	assert.NotNil(t, geo.DomainSets)
	assert.NotNil(t, geo.OppositeIpTags)
	assert.NotNil(t, geo.IpSets)
	assert.NotNil(t, geo.AppSets)
	assert.Empty(t, geo.OppositeDomainTags)
	assert.Empty(t, geo.DomainSets)
	assert.Empty(t, geo.OppositeIpTags)
	assert.Empty(t, geo.IpSets)
	assert.Empty(t, geo.AppSets)
}

func TestNewGeo_WithOppositeDomainTags(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{
				Name:         "proxy",
				OppositeName: "direct",
				InNames:      []string{},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.OppositeDomainTags, 2)
	assert.Equal(t, "direct", geo.OppositeDomainTags["proxy"])
	assert.Equal(t, "proxy", geo.OppositeDomainTags["direct"])
}

func TestNewGeo_WithOppositeIPTags(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:         "cn",
				OppositeName: "foreign",
				InNames:      []string{},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.OppositeIpTags, 2)
	assert.Equal(t, "foreign", geo.OppositeIpTags["cn"])
	assert.Equal(t, "cn", geo.OppositeIpTags["foreign"])
}

func TestNewGeo_WithMultipleOpposites(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{
				Name:         "proxy",
				OppositeName: "direct",
			},
			{
				Name:         "ads",
				OppositeName: "clean",
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:         "cn",
				OppositeName: "foreign",
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	assert.Len(t, geo.OppositeDomainTags, 4)
	assert.Len(t, geo.OppositeIpTags, 2)
}

func TestNewGeo_WithAtomicDomainSet_DirectDomains(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "test-domains",
				Domains: []*cgeo.Domain{
					{
						Type:  cgeo.Domain_Full,
						Value: "example.com",
					},
					{
						Type:  cgeo.Domain_RootDomain,
						Value: "test.com",
					},
				},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.DomainSets, 1)
	assert.Contains(t, geo.DomainSets, "test-domains")
	assert.NotNil(t, geo.DomainSets["test-domains"])
}

func TestNewGeo_WithAtomicIPSet_DirectCIDRs(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "test-ips",
				Cidrs: []*cgeo.CIDR{
					{
						Ip:     []byte{192, 168, 1, 0},
						Prefix: 24,
					},
					{
						Ip:     []byte{10, 0, 0, 0},
						Prefix: 8,
					},
				},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.IpSets, 1)
	assert.Contains(t, geo.IpSets, "test-ips")
	assert.NotNil(t, geo.IpSets["test-ips"])
}

func TestNewGeo_WithAppSets(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AppSets: []*configs.AppSetConfig{
			{
				Name: "test-apps",
				AppIds: []*configs.AppId{
					{
						Type:  0, // Full type
						Value: "com.example.app",
					},
				},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.AppSets, 1)
	assert.Contains(t, geo.AppSets, "test-apps")
	assert.NotNil(t, geo.AppSets["test-apps"])
}

func TestNewGeo_WithGreatDomainSet_SelfReference(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{
				Name:    "self-ref",
				InNames: []string{"self-ref"}, // References itself
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, geo)
	assert.Contains(t, err.Error(), "cannot contain itself")
}

func TestNewGeo_WithGreatDomainSet_SelfReferenceInExNames(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{
				Name:    "self-ref",
				ExNames: []string{"self-ref"}, // References itself in exclusions
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, geo)
	assert.Contains(t, err.Error(), "cannot contain itself")
}

func TestNewGeo_WithGreatIPSet_MissingReference(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:    "combined",
				InNames: []string{"non-existent"},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, geo)
	assert.Contains(t, err.Error(), "not found")
}

func TestNewGeo_WithGreatIPSet_MissingReferenceInExNames(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:    "combined",
				ExNames: []string{"non-existent"},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, geo)
	assert.Contains(t, err.Error(), "not found")
}

func TestNewGeo_WithValidGreatIPSet(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "base-ips",
				Cidrs: []*cgeo.CIDR{
					{
						Ip:     []byte{192, 168, 0, 0},
						Prefix: 16,
					},
				},
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:    "combined",
				InNames: []string{"base-ips"},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.IpSets, 2) // base-ips + combined
	assert.Contains(t, geo.IpSets, "base-ips")
	assert.Contains(t, geo.IpSets, "combined")
}

func TestNewGeo_WithComplexConfiguration(t *testing.T) {
	// Setup - A more realistic configuration
	config := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name: "gfw-domains",
				Domains: []*cgeo.Domain{
					{
						Type:  cgeo.Domain_RootDomain,
						Value: "google.com",
					},
				},
			},
			{
				Name: "custom-proxy",
				Domains: []*cgeo.Domain{
					{
						Type:  cgeo.Domain_Full,
						Value: "www.example.com",
					},
				},
			},
		},
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{
				Name:         "proxy",
				OppositeName: "direct",
				InNames:      []string{"gfw-domains", "custom-proxy"},
			},
		},
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "cn-ips",
				Cidrs: []*cgeo.CIDR{
					{
						Ip:     []byte{114, 114, 114, 0},
						Prefix: 24,
					},
				},
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:         "cn",
				OppositeName: "foreign",
				InNames:      []string{"cn-ips"},
			},
		},
		AppSets: []*configs.AppSetConfig{
			{
				Name: "social-apps",
				AppIds: []*configs.AppId{
					{
						Type:  0, // Full type
						Value: "com.facebook.app",
					},
				},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)

	// Verify domain sets
	assert.Len(t, geo.DomainSets, 3) // gfw-domains, custom-proxy, proxy
	assert.Contains(t, geo.DomainSets, "gfw-domains")
	assert.Contains(t, geo.DomainSets, "custom-proxy")
	assert.Contains(t, geo.DomainSets, "proxy")

	// Verify opposite domain tags
	assert.Len(t, geo.OppositeDomainTags, 2)
	assert.Equal(t, "direct", geo.OppositeDomainTags["proxy"])
	assert.Equal(t, "proxy", geo.OppositeDomainTags["direct"])

	// Verify IP sets
	assert.Len(t, geo.IpSets, 2) // cn-ips, cn
	assert.Contains(t, geo.IpSets, "cn-ips")
	assert.Contains(t, geo.IpSets, "cn")

	// Verify opposite IP tags
	assert.Len(t, geo.OppositeIpTags, 2)
	assert.Equal(t, "foreign", geo.OppositeIpTags["cn"])
	assert.Equal(t, "cn", geo.OppositeIpTags["foreign"])

	// Verify app sets
	assert.Len(t, geo.AppSets, 1)
	assert.Contains(t, geo.AppSets, "social-apps")
}

func TestNewGeo_WithAtomicDomainSet_UseBloomFilter(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicDomainSets: []*configs.AtomicDomainSetConfig{
			{
				Name:           "bloom-test",
				UseBloomFilter: true,
				Domains: []*cgeo.Domain{
					{
						Type:  cgeo.Domain_Full,
						Value: "test.com",
					},
				},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Contains(t, geo.DomainSets, "bloom-test")
}

func TestNewGeo_WithAtomicIPSet_Inverse(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name:    "inverse-ips",
				Inverse: true,
				Cidrs: []*cgeo.CIDR{
					{
						Ip:     []byte{192, 168, 0, 0},
						Prefix: 16,
					},
				},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Contains(t, geo.IpSets, "inverse-ips")
}

func TestNewGeo_WithGreatIPSet_MultipleInclusions(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "set1",
				Cidrs: []*cgeo.CIDR{
					{Ip: []byte{10, 0, 0, 0}, Prefix: 8},
				},
			},
			{
				Name: "set2",
				Cidrs: []*cgeo.CIDR{
					{Ip: []byte{172, 16, 0, 0}, Prefix: 12},
				},
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:    "combined",
				InNames: []string{"set1", "set2"},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Len(t, geo.IpSets, 3) // set1, set2, combined
}

func TestNewGeo_WithGreatIPSet_WithExclusions(t *testing.T) {
	// Setup
	config := &configs.GeoConfig{
		AtomicIpSets: []*configs.AtomicIPSetConfig{
			{
				Name: "all-ips",
				Cidrs: []*cgeo.CIDR{
					{Ip: []byte{0, 0, 0, 0}, Prefix: 0}, // All IPs
				},
			},
			{
				Name: "private-ips",
				Cidrs: []*cgeo.CIDR{
					{Ip: []byte{192, 168, 0, 0}, Prefix: 16},
				},
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:    "public-ips",
				InNames: []string{"all-ips"},
				ExNames: []string{"private-ips"},
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, geo)
	assert.Contains(t, geo.IpSets, "public-ips")
}

func TestNewGeo_WithEmptyOppositeNames(t *testing.T) {
	// Setup - opposite names are empty strings
	config := &configs.GeoConfig{
		GreatDomainSets: []*configs.GreatDomainSetConfig{
			{
				Name:         "test",
				OppositeName: "", // Empty opposite name
			},
		},
		GreatIpSets: []*configs.GreatIPSetConfig{
			{
				Name:         "test-ip",
				OppositeName: "", // Empty opposite name
			},
		},
	}

	// Act
	geo, err := NewGeo(config)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, geo.OppositeDomainTags, "Should not add opposite tags for empty names")
	assert.Empty(t, geo.OppositeIpTags, "Should not add opposite tags for empty names")
}

// =============================================================================
// AppSetConfigToAppSet Tests
// =============================================================================

func TestAppSetConfigToAppSet_BasicAppIds(t *testing.T) {
	// Setup
	config := &configs.AppSetConfig{
		Name: "test-apps",
		AppIds: []*configs.AppId{
			{
				Type:  0, // Full type
				Value: "com.example.app1",
			},
			{
				Type:  1, // Substr type
				Value: "example",
			},
		},
	}
	loader := &MockLoader{}

	// Act
	appSet, err := AppSetConfigToAppSet(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, appSet)
}

func TestAppSetConfigToAppSet_WithClashFiles(t *testing.T) {
	// Setup
	mockApps := []*configs.AppId{
		{
			Type:  0, // Full type
			Value: "com.clash.app",
		},
	}
	loader := &MockLoader{
		loadAppsClashFunc: func(filename string) ([]*configs.AppId, error) {
			if filename == "apps.yaml" {
				return mockApps, nil
			}
			return nil, errors.New("file not found")
		},
	}
	config := &configs.AppSetConfig{
		Name:       "test-apps",
		ClashFiles: []string{"apps.yaml"},
	}

	// Act
	appSet, err := AppSetConfigToAppSet(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, appSet)
}

func TestAppSetConfigToAppSet_ClashFileError(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadAppsClashFunc: func(filename string) ([]*configs.AppId, error) {
			return nil, errors.New("failed to load clash file")
		},
	}
	config := &configs.AppSetConfig{
		Name:       "test-apps",
		ClashFiles: []string{"invalid.yaml"},
	}

	// Act
	appSet, err := AppSetConfigToAppSet(config, loader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, appSet)
	assert.Contains(t, err.Error(), "failed to extract apps from clash file")
}

func TestAppSetConfigToAppSet_EmptyConfig(t *testing.T) {
	// Setup
	config := &configs.AppSetConfig{
		Name: "empty-apps",
	}
	loader := &MockLoader{}

	// Act
	appSet, err := AppSetConfigToAppSet(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, appSet)
}

// =============================================================================
// AtomicDomainSetToIndexMatcher Tests
// =============================================================================

func TestAtomicDomainSetToIndexMatcher_BasicDomains(t *testing.T) {
	// Setup
	config := &configs.AtomicDomainSetConfig{
		Name: "test",
		Domains: []*cgeo.Domain{
			{
				Type:  cgeo.Domain_Full,
				Value: "example.com",
			},
		},
	}
	loader := &MockLoader{}

	// Act
	matcher, err := AtomicDomainSetToIndexMatcher(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, matcher)
}

func TestAtomicDomainSetToIndexMatcher_WithGeosite(t *testing.T) {
	// Setup
	mockDomains := []*cgeo.Domain{
		{
			Type:  cgeo.Domain_Full,
			Value: "geosite.example.com",
		},
	}
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			if filename == "geosite.dat" && list == "cn" {
				return &cgeo.GeoSite{
					CountryCode: "cn",
					Domain:      mockDomains,
				}, nil
			}
			return nil, errors.New("not found")
		},
	}
	config := &configs.AtomicDomainSetConfig{
		Name: "test",
		Geosite: &configs.GeositeConfig{
			Filepath: "geosite.dat",
			Codes:    []string{"cn"},
		},
	}

	// Act
	matcher, err := AtomicDomainSetToIndexMatcher(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, matcher)
}

func TestAtomicDomainSetToIndexMatcher_WithClashFiles(t *testing.T) {
	// Setup
	mockDomains := []*cgeo.Domain{
		{
			Type:  cgeo.Domain_Full,
			Value: "clash.example.com",
		},
	}
	loader := &MockLoader{
		loadDomainsClashFunc: func(filename string) ([]*cgeo.Domain, error) {
			if filename == "domains.yaml" {
				return mockDomains, nil
			}
			return nil, errors.New("file not found")
		},
	}
	config := &configs.AtomicDomainSetConfig{
		Name:       "test",
		ClashFiles: []string{"domains.yaml"},
	}

	// Act
	matcher, err := AtomicDomainSetToIndexMatcher(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, matcher)
}

func TestAtomicDomainSetToIndexMatcher_ClashFileError(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadDomainsClashFunc: func(filename string) ([]*cgeo.Domain, error) {
			return nil, errors.New("failed to load clash file")
		},
	}
	config := &configs.AtomicDomainSetConfig{
		Name:       "test",
		ClashFiles: []string{"invalid.yaml"},
	}

	// Act
	matcher, err := AtomicDomainSetToIndexMatcher(config, loader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, matcher)
	assert.Contains(t, err.Error(), "failed to extract domains from clash file")
}

// =============================================================================
// AtomicIpSetToIPMatcher Tests
// =============================================================================

func TestAtomicIpSetToIPMatcher_BasicCIDRs(t *testing.T) {
	// Setup
	config := &configs.AtomicIPSetConfig{
		Name: "test",
		Cidrs: []*cgeo.CIDR{
			{
				Ip:     []byte{192, 168, 1, 0},
				Prefix: 24,
			},
		},
	}
	loader := &MockLoader{}

	// Act
	matcher, err := AtomicIpSetToIPMatcher(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, matcher)
}

func TestAtomicIpSetToIPMatcher_WithGeoIP(t *testing.T) {
	// Setup
	mockCIDRs := []*cgeo.CIDR{
		{
			Ip:     []byte{114, 114, 114, 0},
			Prefix: 24,
		},
	}
	loader := &MockLoader{
		loadIPFunc: func(filename, country string) (*cgeo.GeoIP, error) {
			if filename == "geoip.dat" && country == "cn" {
				return &cgeo.GeoIP{
					CountryCode: "cn",
					Cidr:        mockCIDRs,
				}, nil
			}
			return nil, errors.New("not found")
		},
	}
	config := &configs.AtomicIPSetConfig{
		Name: "test",
		Geoip: &configs.GeoIPConfig{
			Filepath: "geoip.dat",
			Codes:    []string{"cn"},
		},
	}

	// Act
	matcher, err := AtomicIpSetToIPMatcher(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, matcher)
}

func TestAtomicIpSetToIPMatcher_WithClashFiles(t *testing.T) {
	// Setup
	mockCIDRs := []*cgeo.CIDR{
		{
			Ip:     []byte{10, 0, 0, 0},
			Prefix: 8,
		},
	}
	loader := &MockLoader{
		loadCidrsClashFunc: func(filename string) ([]*cgeo.CIDR, error) {
			if filename == "ips.yaml" {
				return mockCIDRs, nil
			}
			return nil, errors.New("file not found")
		},
	}
	config := &configs.AtomicIPSetConfig{
		Name:       "test",
		ClashFiles: []string{"ips.yaml"},
	}

	// Act
	matcher, err := AtomicIpSetToIPMatcher(config, loader)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, matcher)
}

func TestAtomicIpSetToIPMatcher_ClashFileError(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadCidrsClashFunc: func(filename string) ([]*cgeo.CIDR, error) {
			return nil, errors.New("failed to load clash file")
		},
	}
	config := &configs.AtomicIPSetConfig{
		Name:       "test",
		ClashFiles: []string{"invalid.yaml"},
	}

	// Act
	matcher, err := AtomicIpSetToIPMatcher(config, loader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, matcher)
	assert.Contains(t, err.Error(), "failed to extract cidrs from clash file")
}

func TestAtomicIpSetToIPMatcher_GeoIPLoadError(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadIPFunc: func(filename, country string) (*cgeo.GeoIP, error) {
			return nil, errors.New("failed to load geoip")
		},
	}
	config := &configs.AtomicIPSetConfig{
		Name: "test",
		Geoip: &configs.GeoIPConfig{
			Filepath: "geoip.dat",
			Codes:    []string{"xx"},
		},
	}

	// Act
	matcher, err := AtomicIpSetToIPMatcher(config, loader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, matcher)
	assert.Contains(t, err.Error(), "failed to load geoip")
}

// =============================================================================
// GeositeConfigToGeoDomains Tests
// =============================================================================

func TestGeositeConfigToGeoDomains_BasicLoad(t *testing.T) {
	// Setup
	mockDomains := []*cgeo.Domain{
		{
			Type:  cgeo.Domain_Full,
			Value: "example.com",
		},
	}
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			return &cgeo.GeoSite{
				CountryCode: list,
				Domain:      mockDomains,
			}, nil
		},
	}
	config := &configs.GeositeConfig{
		Filepath: "geosite.dat",
		Codes:    []string{"test"},
	}

	// Act
	domains, err := GeositeConfigToGeoDomains(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, domains, 1)
	assert.Equal(t, "example.com", domains[0].Value)
}

func TestGeositeConfigToGeoDomains_WithAttributes(t *testing.T) {
	// Setup
	mockDomains := []*cgeo.Domain{
		{
			Type:  cgeo.Domain_Full,
			Value: "ads.example.com",
			Attribute: []*cgeo.Domain_Attribute{
				{Key: "ads"},
			},
		},
		{
			Type:  cgeo.Domain_Full,
			Value: "clean.example.com",
			Attribute: []*cgeo.Domain_Attribute{
				{Key: "clean"},
			},
		},
	}
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			return &cgeo.GeoSite{
				CountryCode: list,
				Domain:      mockDomains,
			}, nil
		},
	}
	config := &configs.GeositeConfig{
		Filepath:   "geosite.dat",
		Codes:      []string{"test"},
		Attributes: []string{"ads"},
	}

	// Act
	domains, err := GeositeConfigToGeoDomains(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, domains, 1)
	assert.Equal(t, "ads.example.com", domains[0].Value)
}

func TestGeositeConfigToGeoDomains_AttributesCaseInsensitive(t *testing.T) {
	// Setup
	mockDomains := []*cgeo.Domain{
		{
			Type:  cgeo.Domain_Full,
			Value: "example.com",
			Attribute: []*cgeo.Domain_Attribute{
				{Key: "ADS"}, // Uppercase in data
			},
		},
	}
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			return &cgeo.GeoSite{
				Domain: mockDomains,
			}, nil
		},
	}
	config := &configs.GeositeConfig{
		Filepath:   "geosite.dat",
		Codes:      []string{"test"},
		Attributes: []string{"ads"}, // Lowercase in filter
	}

	// Act
	domains, err := GeositeConfigToGeoDomains(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, domains, 1)
}

func TestGeositeConfigToGeoDomains_EmptyAttributesIgnored(t *testing.T) {
	// Setup
	mockDomains := []*cgeo.Domain{
		{
			Type:  cgeo.Domain_Full,
			Value: "example.com",
		},
	}
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			return &cgeo.GeoSite{
				Domain: mockDomains,
			}, nil
		},
	}
	config := &configs.GeositeConfig{
		Filepath:   "geosite.dat",
		Codes:      []string{"test"},
		Attributes: []string{"", "  ", ""}, // Empty and whitespace attributes
	}

	// Act
	domains, err := GeositeConfigToGeoDomains(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, domains, 1) // Should return all domains when all attributes are empty
}

func TestGeositeConfigToGeoDomains_LoadError(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			return nil, errors.New("failed to load geosite")
		},
	}
	config := &configs.GeositeConfig{
		Filepath: "geosite.dat",
		Codes:    []string{"invalid"},
	}

	// Act
	domains, err := GeositeConfigToGeoDomains(config, loader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, domains)
}

func TestGeositeConfigToGeoDomains_MultipleCodes(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadSiteFunc: func(filename, list string) (*cgeo.GeoSite, error) {
			return &cgeo.GeoSite{
				CountryCode: list,
				Domain: []*cgeo.Domain{
					{
						Type:  cgeo.Domain_Full,
						Value: list + ".example.com",
					},
				},
			}, nil
		},
	}
	config := &configs.GeositeConfig{
		Filepath: "geosite.dat",
		Codes:    []string{"cn", "us", "jp"},
	}

	// Act
	domains, err := GeositeConfigToGeoDomains(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, domains, 3)
}

// =============================================================================
// GeoIpConfigToCidrs Tests
// =============================================================================

func TestGeoIpConfigToCidrs_BasicLoad(t *testing.T) {
	// Setup
	mockCIDRs := []*cgeo.CIDR{
		{
			Ip:     []byte{192, 168, 1, 0},
			Prefix: 24,
		},
	}
	loader := &MockLoader{
		loadIPFunc: func(filename, country string) (*cgeo.GeoIP, error) {
			return &cgeo.GeoIP{
				CountryCode: country,
				Cidr:        mockCIDRs,
			}, nil
		},
	}
	config := &configs.GeoIPConfig{
		Filepath: "geoip.dat",
		Codes:    []string{"cn"},
	}

	// Act
	cidrs, err := GeoIpConfigToCidrs(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, cidrs, 1)
}

func TestGeoIpConfigToCidrs_LoadError(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadIPFunc: func(filename, country string) (*cgeo.GeoIP, error) {
			return nil, errors.New("not found")
		},
	}
	config := &configs.GeoIPConfig{
		Filepath: "geoip.dat",
		Codes:    []string{"invalid"},
	}

	// Act
	cidrs, err := GeoIpConfigToCidrs(config, loader)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, cidrs)
	assert.Contains(t, err.Error(), "failed to load geoip")
}

func TestGeoIpConfigToCidrs_MultipleCodes(t *testing.T) {
	// Setup
	loader := &MockLoader{
		loadIPFunc: func(filename, country string) (*cgeo.GeoIP, error) {
			return &cgeo.GeoIP{
				CountryCode: country,
				Cidr: []*cgeo.CIDR{
					{
						Ip:     []byte{1, 1, 1, 0},
						Prefix: 24,
					},
				},
			}, nil
		},
	}
	config := &configs.GeoIPConfig{
		Filepath: "geoip.dat",
		Codes:    []string{"cn", "us", "jp"},
	}

	// Act
	cidrs, err := GeoIpConfigToCidrs(config, loader)

	// Assert
	require.NoError(t, err)
	assert.Len(t, cidrs, 3)
}
