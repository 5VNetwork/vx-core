package tls

import (
	"crypto/hmac"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport/security"

	"github.com/rs/zerolog/log"
)

var globalSessionCache = tls.NewLRUClientSessionCache(128)

type Engine struct {
	config    *TlsConfig
	tlsConfig *tls.Config

	// ech
	dnsServer i.ECHResolver
}

type EngineConfig struct {
	Config    *TlsConfig
	DnsServer i.ECHResolver
}

// TODO: prebuild tls.Config
func NewEngine(config EngineConfig) (*Engine, error) {
	tlsConfig, err := config.Config.GetTLSConfig()
	if err != nil {
		return nil, err
	}
	return &Engine{config: config.Config, tlsConfig: tlsConfig,
		dnsServer: config.DnsServer}, nil
}

func (c *Engine) GetTLSConfig(opts ...security.Option) *tls.Config {
	tlsConfig := c.tlsConfig.Clone()

	if len(opts) != 0 {
		var options []Option
		for _, o := range opts {
			switch v := o.(type) {
			case security.OptionWithALPN:
				if len(tlsConfig.NextProtos) == 0 {
					if c.config.Imitate != "" {
						if c.config.ForceAlpn == ForceALPN_TRANSPORT_PREFERENCE_TAKE_PRIORITY {
							options = append(options, WithNextProtocol(v.ALPNs))
						}
					} else {
						options = append(options, WithNextProtocol(v.ALPNs))
					}
				}
			case security.OptionWithDestination:
				if tlsConfig.ServerName == "" {
					options = append(options, WithDestination(v.Dest))
				}
			default:
				panic("unreachable")
			}
		}

		for _, o := range options {
			o(tlsConfig)
		}
	}

	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	c.ApplyECH(tlsConfig)

	return tlsConfig
}

func (c *Engine) GetClientConn(conn net.Conn, opts ...security.Option) (net.Conn, error) {
	tlsConfig := c.GetTLSConfig(opts...)

	if c.config.Imitate == "" {
		tlsConn := tls.Client(conn, tlsConfig)
		return &Conn{tlsConn}, nil
	} else { //utls
		return c.config.GetUClient(conn, tlsConfig)
	}
}

// GetTLSConfig converts this Config into tls.Config.
func (c *TlsConfig) GetTLSConfig(opts ...Option) (*tls.Config, error) {
	rootCA, err := c.getRootCA()
	if err != nil {
		return nil, err
	}

	if c == nil {
		return &tls.Config{
			ClientSessionCache:     globalSessionCache,
			RootCAs:                rootCA,
			InsecureSkipVerify:     false,
			NextProtos:             nil,
			SessionTicketsDisabled: true,
		}, nil
	}

	clientCA, err := CertsToCertPool(c.RootCas)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		ClientSessionCache:     globalSessionCache,
		RootCAs:                rootCA,
		InsecureSkipVerify:     c.AllowInsecure,
		NextProtos:             c.NextProtocol,
		SessionTicketsDisabled: !c.EnableSessionResumption,
		VerifyPeerCertificate:  c.VerifyPeerCert,
		ClientCAs:              clientCA,
	}

	if len(c.EchKey) != 0 {
		KeySets, err := ConvertToGoECHKeys(c.EchKey)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal ECHKeySetList: %w", err)
		}
		config.EncryptedClientHelloKeys = KeySets
	}

	for _, opt := range opts {
		opt(config)
	}

	config.Certificates, err = BuildCertificates(c.Certificates)
	if err != nil {
		return nil, err
	}
	config.BuildNameToCertificate()

	if len(c.IssueCas) > 0 {
		config.GetCertificate = getGetCertificateFunc(config, c.IssueCas)
	}

	if sn := c.parseServerName(); len(sn) > 0 {
		config.ServerName = sn
	}

	// if len(config.NextProtos) == 0 {
	// 	config.NextProtos = []string{"h2", "http/1.1"}
	// }

	if c.VerifyClientCertificate {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if len(c.MasterKeyLog) > 0 && c.MasterKeyLog != "none" {
		writer, err := os.OpenFile(c.MasterKeyLog, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			log.Warn().Err(err).Msg("failed to open master key log file")
		} else {
			config.KeyLogWriter = writer
		}
	}

	return config, nil
}

func (c *TlsConfig) parseServerName() string {
	if c.IsExperiment8357() {
		return c.ServerName[len(exp8357):]
	}

	return c.ServerName
}

const exp8357 = "experiment:8357"

func (c *TlsConfig) IsExperiment8357() bool {
	return strings.HasPrefix(c.ServerName, exp8357)
}

func (c *TlsConfig) VerifyPeerCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if c.PinnedPeerCertificateChainSha256 != nil {
		hashValue := GenerateCertChainHash(rawCerts)
		for _, v := range c.PinnedPeerCertificateChainSha256 {
			if hmac.Equal(hashValue, v) {
				return nil
			}
		}
		return fmt.Errorf("peer cert is unrecognized: %v", base64.StdEncoding.EncodeToString(hashValue))
	}
	return nil
}

func isCertificateExpired(c *tls.Certificate) bool {
	if c.Leaf == nil && len(c.Certificate) > 0 {
		if pc, err := x509.ParseCertificate(c.Certificate[0]); err == nil {
			c.Leaf = pc
		}
	}

	// If leaf is not there, the certificate is probably not used yet. We trust user to provide a valid certificate.
	return c.Leaf != nil && c.Leaf.NotAfter.Before(time.Now().Add(time.Minute*2))
}

func getGetCertificateFunc(c *tls.Config, ca []*Certificate) func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var access sync.RWMutex

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		domain := hello.ServerName
		certExpired := false

		access.RLock()
		certificate, found := c.NameToCertificate[domain]
		access.RUnlock()

		if found {
			if !isCertificateExpired(certificate) {
				return certificate, nil
			}
			certExpired = true
		}

		if certExpired {
			newCerts := make([]tls.Certificate, 0, len(c.Certificates))

			access.Lock()
			for _, certificate := range c.Certificates {
				cert := certificate
				if !isCertificateExpired(&cert) {
					newCerts = append(newCerts, cert)
				} else if cert.Leaf != nil {
					expTime := cert.Leaf.NotAfter.Format(time.RFC3339)
					log.Info().Msgf("old certificate for %s (expire on %s) discard", domain, expTime)
				}
			}

			c.Certificates = newCerts
			access.Unlock()
		}

		var issuedCertificate *tls.Certificate

		// Create a new certificate from existing CA if possible
		for _, rawCert := range ca {
			newCert, err := issueCertificate(rawCert, domain)
			if err != nil {
				log.Warn().Str("domain", domain).Msg("failed to issue new certificate")
				continue
			}
			parsed, err := x509.ParseCertificate(newCert.Certificate[0])
			if err == nil {
				newCert.Leaf = parsed
				expTime := parsed.NotAfter.Format(time.RFC3339)
				log.Info().Msgf("new certificate for %s (expire on %s) issued", domain, expTime)
			} else {
				log.Warn().Str("domain", domain).Msg("failed to parse new certificate")
			}

			access.Lock()
			c.Certificates = append(c.Certificates, *newCert)
			issuedCertificate = &c.Certificates[len(c.Certificates)-1]
			access.Unlock()
			break
		}

		if issuedCertificate == nil {
			return nil, fmt.Errorf("failed to create a new certificate for %v", domain)
		}

		access.Lock()
		c.BuildNameToCertificate()
		access.Unlock()

		return issuedCertificate, nil
	}
}

func issueCertificate(rawCA *Certificate, domain string) (*tls.Certificate, error) {
	parent, err := cert.ParseCertificate(rawCA.Certificate, rawCA.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse raw certificate %w", err)
	}
	newCert, err := cert.Generate(parent, cert.CommonName(domain), cert.DNSNames(domain))
	if err != nil {
		return nil, fmt.Errorf("failed to generate new certificate for %v %w", domain, err)
	}
	newCertPEM, newKeyPEM := newCert.ToPEM()
	cert, err := tls.X509KeyPair(newCertPEM, newKeyPEM)
	return &cert, err
}

// ParseCertificate converts a cert.Certificate to Certificate.
// func ParseCertificate(c *cert.Certificate) *CertificateAndKey {
// 	if c != nil {
// 		certPEM, keyPEM := c.ToPEM()
// 		return &CertificateAndKey{
// 			certificate: certPEM,
// 			key:         keyPEM,
// 		}
// 	}
// 	return nil
// }

func CertsToCertPool(certs [][]byte) (*x509.CertPool, error) {
	root := x509.NewCertPool()
	for _, cert := range certs {
		if !root.AppendCertsFromPEM(cert) {
			return nil, errors.New("failed to append cert")
		}
	}
	return root, nil
}

// ParseCertificate converts a cert.Certificate to Certificate.
func ParseCertificate(c *cert.Certificate) *Certificate {
	if c != nil {
		certPEM, keyPEM := c.ToPEM()
		return &Certificate{
			Certificate: certPEM,
			Key:         keyPEM,
		}
	}
	return nil
}
