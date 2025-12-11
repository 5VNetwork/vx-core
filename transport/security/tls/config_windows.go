package tls

import (
	"crypto/x509"
)

func (c *TlsConfig) getRootCA() (*x509.CertPool, error) {
	if c.DisableSystemRoot {
		return CertsToCertPool(c.RootCas)
	}
	return nil, nil
}
