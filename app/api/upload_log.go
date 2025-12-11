package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (a *Api) UploadLog(ctx context.Context, req *UploadLogRequest) (*UploadLogResponse, error) {
	ca := x509.NewCertPool()
	if ok := ca.AppendCertsFromPEM(req.Ca); !ok {
		return nil, errors.New("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs:            ca,
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// Parse the peer certificate
			if len(rawCerts) == 0 {
				return errors.New("no certificates provided")
			}

			cert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return err
			}

			// Verify the certificate against our CA
			opts := x509.VerifyOptions{
				Roots: ca,
			}

			_, err = cert.Verify(opts)
			return err
		},
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: transport,
	}

	url := req.Url
	payload := []byte(req.Body)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Version", req.Version)
	httpReq.Header.Set("Authorization", req.Secret)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and print the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Response status: %s", resp.Status)
	log.Debug().Msgf("Response body: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to upload log")
	}

	return &UploadLogResponse{}, nil
}
