package tls

import (
	"crypto/tls"
	"os"
)

func BuildCertificates(certConfigs []*Certificate) ([]tls.Certificate, error) {
	certs := make([]tls.Certificate, 0, len(certConfigs))
	for _, entry := range certConfigs {
		var certBytes []byte
		var keyBytes []byte
		var err error
		if entry.Certificate != nil {
			certBytes = entry.Certificate
		}
		if entry.Key != nil {
			keyBytes = entry.Key
		}
		if entry.CertificateFilepath != "" {
			certBytes, err = os.ReadFile(entry.CertificateFilepath)
			if err != nil {
				return nil, err
			}
		}
		if entry.KeyFilepath != "" {
			keyBytes, err = os.ReadFile(entry.KeyFilepath)
			if err != nil {
				return nil, err
			}
		}
		keyPair, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, keyPair)
	}
	return certs, nil
}
