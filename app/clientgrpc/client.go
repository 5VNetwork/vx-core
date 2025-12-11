package clientgrpc

import (
	"context"
	gotls "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	configs "github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/signal/done"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	// "github.com/shirou/gopsutil/v4/cpu"
	// "github.com/shirou/gopsutil/v4/disk"
	// "github.com/shirou/gopsutil/v4/host"
	// "github.com/shirou/gopsutil/v4/mem"
)

type ClientGrpc struct {
	streamLock        sync.RWMutex
	communicateStream ClientService_CommunicateServer

	Client *client.Client

	Done       *done.Instance
	GrpcConfig *configs.GrpcConfig

	// if runningInService, when flutter side disconnect(meaning the app exits), call OnExit after 2 seconds
	RunningInService bool
	timeoutExit      *time.Timer
	OnExit           func()

	UpdateLantency bool

	server *grpc.Server
	UnimplementedClientServiceServer
}

func (s *ClientGrpc) Start() error {
	var lis net.Listener
	var err error
	var opts []grpc.ServerOption
	// listen unix socket
	if s.GrpcConfig.Port == 0 {
		os.Remove(s.GrpcConfig.Address)
		lis, err = net.ListenUnix("unix", &net.UnixAddr{Name: s.GrpcConfig.Address, Net: "unix"})
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
		if s.GrpcConfig.Uid != 0 && s.GrpcConfig.Gid != 0 {
			err = os.Chown(s.GrpcConfig.Address, int(s.GrpcConfig.Uid), int(s.GrpcConfig.Gid))
		} else {
			err = os.Chown(s.GrpcConfig.Address, os.Getuid(), os.Getgid())
		}
		if err != nil {
			lis.Close()
			os.Remove(s.GrpcConfig.Address)
			return fmt.Errorf("failed to chown: %w", err)
		}
	} else {
		certificate, err := cert.Generate(nil, cert.NotBefore(time.Now().Add(-time.Hour*24*365)),
			cert.NotAfter(time.Now().Add(time.Hour*24*365)))
		if err != nil {
			return fmt.Errorf("failed to generate certificate: %s", err)
		}
		certPEM, keyPEM := certificate.ToPEM()
		cert, err := gotls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("failed to load key pair: %s", err)
		}
		ca := x509.NewCertPool()
		if ok := ca.AppendCertsFromPEM(s.GrpcConfig.ClientCert); !ok {
			return errors.New("failed to parse CA certificate")
		}
		tlsConfig := &gotls.Config{
			ClientCAs:  ca,
			ClientAuth: gotls.RequestClientCert,
			Certificates: []gotls.Certificate{
				cert,
			},
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.GrpcConfig.Address, s.GrpcConfig.Port))
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	}

	opts = append(opts,
		grpc.InitialWindowSize(64*1024),
		grpc.InitialConnWindowSize(128*1024),
	)

	s.server = grpc.NewServer(opts...)
	RegisterClientServiceServer(s.server, s)
	go func() {
		// lis will be closed when this method returns
		if err := s.server.Serve(lis); err != nil {
			log.Error().Err(err).Msg("failed to serve grpc")
		}
	}()
	s.GrpcConfig = nil
	return nil
}

func (s *ClientGrpc) Close() error {
	s.Done.Close()
	// TODO:
	// s.Subscription.Close()
	if s.server != nil {
		go s.server.GracefulStop()
	}
	return nil
}

func (s *ClientGrpc) SetSubscriptionInterval(ctx context.Context, req *SetSubscriptionIntervalRequest) (*SetSubscriptionIntervalResponse, error) {
	log.Debug().Msgf("set subscription interval: %d", req.Interval)
	s.Client.Subscription.SetInterval(time.Duration(req.Interval) * time.Minute)
	return &SetSubscriptionIntervalResponse{}, nil
}

func (s *ClientGrpc) SetAutoSubscriptionUpdate(ctx context.Context, req *SetAutoSubscriptionUpdateRequest) (*Receipt, error) {
	log.Debug().Msgf("set auto subscription update: %t", req.Enable)
	s.Client.Subscription.SetAutoUpdate(req.Enable)
	return &Receipt{}, nil
}
