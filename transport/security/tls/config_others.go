//go:build !windows

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

func (c *TlsConfig) getRootCA() (*x509.CertPool, error) {
	if c.DisableSystemRoot {
		return CertsToCertPool(c.RootCas)
	}

	if len(c.RootCas) == 0 {
		return rootCerts.load()
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("system root %w", err)
	}
	for _, cert := range c.RootCas {
		if !pool.AppendCertsFromPEM(cert) {
			return nil, errors.New("failed to append cert to root")
		}
	}
	return pool, err
}
