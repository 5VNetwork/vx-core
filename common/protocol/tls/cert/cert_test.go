package cert

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/errors"
	"github.com/5vnetwork/vx-core/common/task"
)

func TestGenerate(t *testing.T) {
	err := generate(nil, true, true, "ca")
	if err != nil {
		t.Fatal(err)
	}
}

func generate(domainNames []string, isCA bool, jsonOutput bool, fileOutput string) error {
	commonName := "V2Ray Root CA"
	organization := "V2Ray Inc"

	expire := time.Hour * 3

	var opts []Option
	if isCA {
		opts = append(opts, Authority(isCA))
		opts = append(opts, KeyUsage(x509.KeyUsageCertSign|x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature))
	}

	opts = append(opts, NotAfter(time.Now().Add(expire)))
	opts = append(opts, CommonName(commonName))
	if len(domainNames) > 0 {
		opts = append(opts, DNSNames(domainNames...))
	}
	opts = append(opts, Organization(organization))

	cert, err := Generate(nil, opts...)
	if err != nil {
		return errors.New("failed to generate TLS certificate").Base(err)
	}

	if jsonOutput {
		printJSON(cert)
	}

	if len(fileOutput) > 0 {
		if err := printFile(cert, fileOutput); err != nil {
			return err
		}
	}

	return nil
}

type jsonCert struct {
	Certificate []string `json:"certificate"`
	Key         []string `json:"key"`
}

func printJSON(certificate *Certificate) {
	certPEM, keyPEM := certificate.ToPEM()
	jCert := &jsonCert{
		Certificate: strings.Split(strings.TrimSpace(string(certPEM)), "\n"),
		Key:         strings.Split(strings.TrimSpace(string(keyPEM)), "\n"),
	}
	content, err := json.MarshalIndent(jCert, "", "  ")
	common.Must(err)
	os.Stdout.Write(content)
	os.Stdout.WriteString("\n")
}

func printFile(certificate *Certificate, name string) error {
	certPEM, keyPEM := certificate.ToPEM()
	return task.Run(context.Background(), func() error {
		return writeFile(certPEM, name+"_cert.pem")
	}, func() error {
		return writeFile(keyPEM, name+"_key.pem")
	})
}

func writeFile(content []byte, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return common.Error2(f.Write(content))
}

func TestExtractDomainFromCertificate_WithDNSNames(t *testing.T) {
	// Generate certificate with DNSNames
	cert := MustGenerate(nil,
		DNSNames("example.com", "www.example.com", "api.example.com"),
		CommonName("example.com"),
	)

	certPEM, _ := cert.ToPEM()

	domain, err := ExtractDomainFromCertificate(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if domain != "example.com" {
		t.Errorf("expected domain 'example.com', got '%s'", domain)
	}
}

func TestExtractDomainFromCertificate_WithCommonNameOnly(t *testing.T) {
	// Generate certificate with only CommonName (no DNSNames)
	cert := MustGenerate(nil,
		CommonName("test.example.com"),
	)

	certPEM, _ := cert.ToPEM()

	domain, err := ExtractDomainFromCertificate(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if domain != "test.example.com" {
		t.Errorf("expected domain 'test.example.com', got '%s'", domain)
	}
}

func TestExtractDomainFromCertificate_PrefersDNSNamesOverCommonName(t *testing.T) {
	// Generate certificate with both DNSNames and CommonName
	// Should prefer DNSNames
	cert := MustGenerate(nil,
		DNSNames("primary.example.com"),
		CommonName("secondary.example.com"),
	)

	certPEM, _ := cert.ToPEM()

	domain, err := ExtractDomainFromCertificate(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return DNSNames (primary) not CommonName (secondary)
	if domain != "primary.example.com" {
		t.Errorf("expected domain 'primary.example.com' (from DNSNames), got '%s'", domain)
	}
}

func TestExtractDomainFromCertificate_NoDomain(t *testing.T) {
	// Generate certificate without any domain
	cert := MustGenerate(nil,
		Organization("Test Org"),
	)

	certPEM, _ := cert.ToPEM()

	domain, err := ExtractDomainFromCertificate(certPEM)
	if err == nil {
		t.Errorf("expected error for certificate with no domain, got domain: '%s'", domain)
	}

	if !strings.Contains(err.Error(), "no domain found") {
		t.Errorf("expected error message to contain 'no domain found', got: %v", err)
	}
}

func TestExtractDomainFromCertificate_InvalidPEM(t *testing.T) {
	// Test with invalid PEM data
	invalidPEM := []byte("-----BEGIN INVALID-----\ninvalid data\n-----END INVALID-----")

	domain, err := ExtractDomainFromCertificate(invalidPEM)
	if err == nil {
		t.Errorf("expected error for invalid PEM, got domain: '%s'", domain)
	}

	if !strings.Contains(err.Error(), "failed to decode PEM block") && !strings.Contains(err.Error(), "failed to parse certificate") {
		t.Errorf("expected error about PEM decode or parse failure, got: %v", err)
	}
}

func TestExtractDomainFromCertificate_EmptyPEM(t *testing.T) {
	// Test with empty PEM
	emptyPEM := []byte("")

	domain, err := ExtractDomainFromCertificate(emptyPEM)
	if err == nil {
		t.Errorf("expected error for empty PEM, got domain: '%s'", domain)
	}
}

func TestExtractDomainFromCertificate_NonPEMData(t *testing.T) {
	// Test with non-PEM data
	nonPEM := []byte("this is not PEM data at all")

	domain, err := ExtractDomainFromCertificate(nonPEM)
	if err == nil {
		t.Errorf("expected error for non-PEM data, got domain: '%s'", domain)
	}
}

func TestExtractDomainFromCertificate_MultipleDNSNames(t *testing.T) {
	// Test that it returns the first DNS name when multiple are present
	cert := MustGenerate(nil,
		DNSNames("first.com", "second.com", "third.com"),
	)

	certPEM, _ := cert.ToPEM()

	domain, err := ExtractDomainFromCertificate(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the first DNS name
	if domain != "first.com" {
		t.Errorf("expected domain 'first.com' (first DNS name), got '%s'", domain)
	}
}

func TestExtractDomainFromCertificate_WildcardDomain(t *testing.T) {
	// Test with wildcard domain
	cert := MustGenerate(nil,
		DNSNames("*.example.com", "example.com"),
	)

	certPEM, _ := cert.ToPEM()

	domain, err := ExtractDomainFromCertificate(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the first DNS name (wildcard)
	if domain != "*.example.com" {
		t.Errorf("expected domain '*.example.com', got '%s'", domain)
	}
}
