package api

import (
	"context"
	"os"
	"testing"
)

//	func TestGeoIP(t *testing.T) {
//		api := &Api{
//			ApiServerConfig: &ApiServerConfig{
//				GeoipPath: "../../test/assets/geoip.dat",
//			},
//		}
//		rsp, err := api.GeoIP(context.Background(), &GeoIPRequest{
//			Ips: []string{"5.62.60.5"},
//		})
//		if err != nil {
//			t.Fatal(err)
//		}
//		if rsp.Countries[0] != "AD" {
//			t.Fatal("expect AD, but actually", rsp.Countries[0])
//		}
//	}
// func TestGeoIPReturnError(t *testing.T) {
// 	api := &Api{
// 		ApiServerConfig: &ApiServerConfig{
// 			GeoipPath: "../../test/assets/nonexist.dat",
// 		},
// 	}

// 	_, err := api.GeoIP(context.Background(), &GeoIPRequest{
// 		Ips: []string{"5.62.60.5"},
// 	})
// 	if err == nil {
// 		t.Fatal("expect error, but actually nil")
// 	}
// 	if err != ErrGeoIPFileNotFound {
// 		t.Fatal("expect error, but actually", err)
// 	}
// }

func TestProcessGeoFiles(t *testing.T) {
	// copy files
	// Copy the original geosite.dat and geoip.dat files to temporary locations
	// so we don't modify the original test files
	geositeSrc := "../../test/assets/geosite.dat"
	geoipSrc := "../../test/assets/geoip.dat"

	// Read the source files
	geositeData, err := os.ReadFile(geositeSrc)
	if err != nil {
		t.Fatalf("Failed to read geosite source file: %v", err)
	}

	geoipData, err := os.ReadFile(geoipSrc)
	if err != nil {
		t.Fatalf("Failed to read geoip source file: %v", err)
	}

	// Create backup copies
	err = os.WriteFile(geositeSrc+".bak", geositeData, 0644)
	if err != nil {
		t.Fatalf("Failed to create geosite backup: %v", err)
	}
	defer os.Remove(geositeSrc + ".bak")

	err = os.WriteFile(geoipSrc+".bak", geoipData, 0644)
	if err != nil {
		t.Fatalf("Failed to create geoip backup: %v", err)
	}
	defer os.Remove(geoipSrc + ".bak")

	api := &Api{}
	_, err = api.ProcessGeoFiles(context.Background(), &ProcessGeoFilesRequest{
		GeositeCodes:   []string{"cn"},
		GeoipCodes:     []string{"private"},
		GeositePath:    "../../test/assets/geosite.dat.bak",
		GeoipPath:      "../../test/assets/geoip.dat.bak",
		DstGeositePath: "../../test/assets/geosite_simple.dat.bak",
		DstGeoipPath:   "../../test/assets/geoip_simple.dat.bak",
	})
	if err != nil {
		t.Fatal(err)
	}
	// verify files existed
	// Verify that the files were created
	files := []string{
		"../../test/assets/geosite_simple.dat.bak",
		"../../test/assets/geoip_simple.dat.bak",
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist, but it doesn't", file)
		}
	}
	// remove the files
	os.Remove(files[0])
	os.Remove(files[1])
}
