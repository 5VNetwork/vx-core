package geo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/platform/file"

	"github.com/golang/protobuf/proto"
)

func init() {
	wd, err := os.Getwd()
	common.Must(err)

	assetsPath := filepath.Join(wd, "..", "..", "test", "assets")
	// geoipPath := filepath.Join(assetsPath, "geoip.dat")

	os.Setenv("x.location.asset", assetsPath)

	// if _, err := os.Stat(geoipPath); err != nil && errors.Is(err, fs.ErrNotExist) {
	// 	common.Must(os.MkdirAll(assetsPath, 0o755))
	// 	geoipBytes, err := net.FetchHTTPContent(geoipURL)
	// 	common.Must(err)
	// 	common.Must(file.WriteFile(geoipPath, geoipBytes))
	// }
}

func loadGeoIP(country string) ([]*CIDR, error) {
	path, _ := os.LookupEnv("x.location.asset")
	bytes, err := file.ReadFile(filepath.Join(path, "geoip.dat"))
	if err != nil {
		return nil, err
	}
	var geoipList GeoIPList
	if err := proto.Unmarshal(bytes, &geoipList); err != nil {
		return nil, err
	}

	for _, geoip := range geoipList.Entry {
		if strings.EqualFold(geoip.CountryCode, country) {
			return geoip.Cidr, nil
		}
	}

	panic("country not found: " + country)
}

// func TestBuildIPMatcherListWithRepeat(t *testing.T) {
// 	geoIPs := []*GeoIP{
// 		{
// 			CountryCode: "CN",
// 		},
// 		{CountryCode: "CN"},
// 	}
// 	IPMatchers, _ := BuildIPMatcherList(geoIPs)
// 	if len(IPMatchers) != 1 {
// 		t.Errorf("expecting 1 ipmather, but got %d", len(IPMatchers))
// 	}
// }

func TestIPMatcher(t *testing.T) {
	cidrList := []*CIDR{
		{Ip: []byte{0, 0, 0, 0}, Prefix: 8},
		{Ip: []byte{10, 0, 0, 0}, Prefix: 8},
		{Ip: []byte{100, 64, 0, 0}, Prefix: 10},
		{Ip: []byte{127, 0, 0, 0}, Prefix: 8},
		{Ip: []byte{169, 254, 0, 0}, Prefix: 16},
		{Ip: []byte{172, 16, 0, 0}, Prefix: 12},
		{Ip: []byte{192, 0, 0, 0}, Prefix: 24},
		{Ip: []byte{192, 0, 2, 0}, Prefix: 24},
		{Ip: []byte{192, 168, 0, 0}, Prefix: 16},
		{Ip: []byte{192, 18, 0, 0}, Prefix: 15},
		{Ip: []byte{198, 51, 100, 0}, Prefix: 24},
		{Ip: []byte{203, 0, 113, 0}, Prefix: 24},
		{Ip: []byte{8, 8, 8, 8}, Prefix: 32},
		{Ip: []byte{91, 108, 4, 0}, Prefix: 16},
	}

	ipMatcher, err := NewIPMatcherFromGeoCidrs(cidrList, false)
	if err != nil {
		t.Error("cannot create ipmatcher")
	}

	testCases := []struct {
		Input  string
		Output bool
	}{
		{
			Input:  "192.168.1.1",
			Output: true,
		},
		{
			Input:  "192.0.0.0",
			Output: true,
		},
		{
			Input:  "192.0.1.0",
			Output: false,
		},
		{
			Input:  "0.1.0.0",
			Output: true,
		},
		{
			Input:  "1.0.0.1",
			Output: false,
		},
		{
			Input:  "8.8.8.7",
			Output: false,
		},
		{
			Input:  "8.8.8.8",
			Output: true,
		},
		{
			Input:  "2001:cdba::3257:9652",
			Output: false,
		},
		{
			Input:  "91.108.255.254",
			Output: true,
		},
	}

	for _, testCase := range testCases {
		ip := net.ParseAddress(testCase.Input).IP()
		actual := ipMatcher.Match(ip)
		if actual != testCase.Output {
			t.Error("expect input", testCase.Input, "to be", testCase.Output, ", but actually", actual)
		}
	}
}

func TestIPReverseMatcher(t *testing.T) {
	cidrList := []*CIDR{
		{Ip: []byte{8, 8, 8, 8}, Prefix: 32},
		{Ip: []byte{91, 108, 4, 0}, Prefix: 16},
	}

	ipMatcher, err := NewIPMatcherFromGeoCidrs(cidrList, true)
	if err != nil {
		t.Error("cannot create ipmatcher")
	}

	testCases := []struct {
		Input  string
		Output bool
	}{
		{
			Input:  "8.8.8.8",
			Output: false,
		},
		{
			Input:  "2001:cdba::3257:9652",
			Output: true,
		},
		{
			Input:  "91.108.255.254",
			Output: false,
		},
	}

	for _, testCase := range testCases {
		ip := net.ParseAddress(testCase.Input).IP()
		actual := ipMatcher.Match(ip)
		if actual != testCase.Output {
			t.Error("expect input", testCase.Input, "to be", testCase.Output, ", but actually", actual)
		}
	}
}

func TestIPMatcher4CN(t *testing.T) {
	ips, err := loadGeoIP("CN")
	common.Must(err)

	ipMatcher, err := NewIPMatcherFromGeoCidrs(ips, false)
	if err != nil {
		t.Error("cannot create ipmatcher")
	}

	if ipMatcher.Match([]byte{8, 8, 8, 8}) {
		t.Error("expect CN geoip doesn't contain 8.8.8.8, but actually does")
	}
}

func TestIPMatcher6US(t *testing.T) {
	ips, err := loadGeoIP("US")
	common.Must(err)

	ipMatcher, err := NewIPMatcherFromGeoCidrs(ips, false)
	if err != nil {
		t.Error("cannot create ipmatcher")
	}

	if !ipMatcher.Match(net.ParseAddress("2001:4860:4860::8888").IP()) {
		t.Error("expect US geoip contain 2001:4860:4860::8888, but actually not")
	}
}

// func TestIPMatcherProxy(t *testing.T) {
// 	var allIps []*CIDR
// 	var err error
// 	var ips []*CIDR
// 	// ips, err := loadGeoIP("telegram")
// 	// common.Must(err)
// 	// allIps = append(allIps, ips...)

// 	ips, err = loadGeoIP("google")
// 	common.Must(err)
// 	allIps = append(allIps, ips...)

// 	// ips, err = loadGeoIP("facebook")
// 	// common.Must(err)
// 	// allIps = append(allIps, ips...)

// 	// ips, err = loadGeoIP("twitter")
// 	// common.Must(err)
// 	// allIps = append(allIps, ips...)

// 	// ips, err = loadGeoIP("tor")
// 	// common.Must(err)
// 	// allIps = append(allIps, ips...)

// 	ipMatcher, err := NewIPMatcherFromGeoCidrs(allIps, false)
// 	if err != nil {
// 		t.Error("cannot create ipmatcher")
// 	}

// 	match := ipMatcher.Match(net.ParseIP("2600:1901:0:6d85::"))
// 	if !match {
// 		t.Error("expect 2600:1901:0:6d85:: to be matched, but actually not")
// 	}
// }

func BenchmarkIPMatcher4CN(b *testing.B) {
	ips, err := loadGeoIP("CN")
	common.Must(err)

	ipMatcher, err := NewIPMatcherFromGeoCidrs(ips, false)
	if err != nil {
		b.Error("cannot create ipmatcher")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ipMatcher.Match([]byte{8, 8, 8, 8})
	}
}

func BenchmarkIPMatcher6US(b *testing.B) {
	ips, err := loadGeoIP("US")
	common.Must(err)

	ipMatcher, err := NewIPMatcherFromGeoCidrs(ips, false)
	if err != nil {
		b.Error("cannot create ipmatcher")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ipMatcher.Match(net.ParseAddress("2001:4860:4860::8888").IP())
	}
}
