package grpcserver

import (
	"crypto/tls"
	gotls "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcServer struct {
	Server *grpc.Server
	Config *configs.GrpcConfig
}

func NewGrpcServer(config *configs.GrpcConfig) (*GrpcServer, error) {
	var opts []grpc.ServerOption

	if config.Port != 0 {
		certificate, err := cert.Generate(nil, cert.NotBefore(time.Now().Add(-time.Hour*24*365)),
			cert.NotAfter(time.Now().Add(time.Hour*24*365)))
		if err != nil {
			return nil, fmt.Errorf("failed to generate certificate: %s", err)
		}
		certPEM, keyPEM := certificate.ToPEM()
		cert, err := gotls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to load key pair: %s", err)
		}
		ca := x509.NewCertPool()
		if ok := ca.AppendCertsFromPEM(config.ClientCert); !ok {
			return nil, errors.New("failed to parse CA certificate")
		}
		tlsConfig := &tls.Config{
			ClientCAs:  ca,
			ClientAuth: gotls.RequestClientCert,
			Certificates: []gotls.Certificate{
				cert,
			},
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	opts = append(opts,
		grpc.InitialWindowSize(64*1024),
		grpc.InitialConnWindowSize(128*1024),
	)

	server := grpc.NewServer(opts...)
	return &GrpcServer{Server: server, Config: config}, nil
}

func (s *GrpcServer) Start() error {
	var lis net.Listener
	var err error
	config := s.Config

	// listen unix socket
	if config.Port == 0 {
		os.Remove(config.Address)
		lis, err = net.ListenUnix("unix", &net.UnixAddr{Name: config.Address, Net: "unix"})
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
		if config.Uid != 0 && config.Gid != 0 {
			err = os.Chown(config.Address, int(config.Uid), int(config.Gid))
		} else {
			err = os.Chown(config.Address, os.Getuid(), os.Getgid())
		}
		if err != nil {
			lis.Close()
			os.Remove(config.Address)
			return fmt.Errorf("failed to chown: %w", err)
		}
	} else {
		lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", config.Address, config.Port))
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	}
	go func() {
		if err := s.Server.Serve(lis); err != nil {
			log.Error().Err(err).Msg("failed to serve grpc")
		}
	}()
	return nil
}

func (s *GrpcServer) Stop() error {
	go s.Server.GracefulStop()
	return nil
}
