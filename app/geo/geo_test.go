package geo_test

// . "github.com/5vnetwork/vx-core/app/geo"

// func TestGeo(t *testing.T) {
// 	// p, _ := os.Executable()
// 	// log.Print("!!!!!!", p)
// 	config := &Config{
// 		GreatDomainSets: []*GreatDomainSet{
// 			{
// 				Name:         "proxy",
// 				OppositeName: "direct",
// 				InNames:      []string{"gfw", "custom-proxy"},
// 			},
// 		},
// 		AtomicDomainSets: []*AtomicDomainSet{
// 			{
// 				Name: "gfw",
// 				Geosite: &geo.GeositeConfig{
// 					Filepath: "../../test/assets/geosite.dat",
// 					Codes:    []string{"gfw"},
// 				},
// 			},
// 			{
// 				Name: "custom-proxy",
// 				Domains: []*geo.Domain{
// 					{
// 						Value: "www.asdf.com",
// 						Type:  geo.Domain_Full,
// 					},
// 				},
// 			},
// 		},
// 		AtomicIpSets: []*AtomicIPSet{
// 			{
// 				Name: "cn",
// 				Geoip: &geo.GeoIPConfig{
// 					Filepath: "../../test/assets/geoip.dat",
// 					Codes:    []string{"cn"},
// 				},
// 			},
// 		},
// 	}
// 	geo := New()
// 	if err := geo.Init(config); err != nil {
// 		t.Fatal(err)
// 	}

// 	if !geo.MatchDomain("www.google.com", "proxy") {
// 		t.Fatal("expected true")
// 	}
// 	if !geo.MatchDomain("www.baidu.com", "direct") {
// 		t.Fatal("expected true")
// 	}
// 	if !geo.MatchDomain("www.asdf.com", "proxy") {
// 		t.Fatal("expected true")
// 	}

// 	if geo.MatchIP([]byte{8, 8, 8, 8}, "cn") {
// 		t.Fatal("expected false")
// 	}

// 	if !geo.MatchIP([]byte{114, 114, 114, 114}, "cn") {
// 		t.Fatal("expected true")
// 	}

// 	if geo.MatchIP([]byte{8, 8, 8, 8}, "non-exist") {
// 		t.Fatal("expected false")
// 	}
// 	if geo.MatchDomain("www.baidu.com", "non-exist") {
// 		t.Fatal("expected false")
// 	}
// }
