package tls

import (
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
)

type rootCertsCache struct {
	sync.Mutex
	pool *x509.CertPool
}

func (c *rootCertsCache) load() (*x509.CertPool, error) {
	c.Lock()
	defer c.Unlock()

	if c.pool != nil {
		return c.pool, nil
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	c.pool = pool
	return pool, nil
}

var rootCerts rootCertsCache

func getRootCA(c *TlsConfig) (*x509.CertPool, error) {
	if c.GetDisableSystemRoot() {
		return CertsToCertPool(c.GetRootCas())
	}

	if len(c.GetRootCas()) == 0 {
		return rootCerts.load()
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("system root %w", err)
	}
	for _, cert := range c.GetRootCas() {
		if !pool.AppendCertsFromPEM(cert) {
			return nil, errors.New("failed to append cert to root")
		}
	}
	return pool, err
}
