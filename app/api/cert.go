package api

import (
	"context"
	"time"

	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
)

func (a *Api) GenerateCert(ctx context.Context, req *GenerateCertRequest) (*GenerateCertResponse, error) {
	crt, err := cert.Generate(nil,
		cert.DNSNames(req.Domain),
		cert.CommonName(req.Domain),
		cert.NotBefore(time.Now().Add(-time.Hour*24*365)),
		cert.NotAfter(time.Now().Add(time.Hour*24*365)))
	if err != nil {
		return nil, err
	}

	certificatePem, keyPem := crt.ToPEM()
	certHash, err := cert.GetCertHash(certificatePem)
	if err != nil {
		return nil, err
	}

	return &GenerateCertResponse{
		Cert:     certificatePem,
		Key:      keyPem,
		CertHash: []byte(certHash),
	}, nil
}

func (a *Api) GetCertDomain(ctx context.Context,
	req *GetCertDomainRequest) (*GetCertDomainResponse, error) {
	domain, err := cert.ExtractDomainFromCertificate(req.Cert)
	if err != nil {
		return nil, err
	}
	return &GetCertDomainResponse{
		Domain: domain,
	}, nil
}
