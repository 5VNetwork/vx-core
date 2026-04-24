// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package api

import (
	context "context"
	gotls "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"runtime"
	sync "sync"
	"time"

	"github.com/5vnetwork/vx-core/app/create"
	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/sysproxy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/proxy/freedom"
	"github.com/5vnetwork/vx-core/transport/dlhelper"

	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/util/downloader"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/dispatcher"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/session"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	vxlog "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/log"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TODO: set log level
type Api struct {
	*ApiServerConfig
	server *grpc.Server
	UnimplementedApiServer

	log *logger.Logger
	// geo
	geoLock        sync.Mutex
	geoIpMatcher   *geoIpMatcher
	timeoutChecker *signal.ActivityChecker

	// server secrets
	secretLock sync.Mutex
	secrets    map[uint32]string //key is id, value is secret

	// only used when vpn is on
	dialFactory transport.DialerFactory
	mon         i.DefaultInterfaceInfo
	// used to resolve ip address of out handlers when vpn is on
	ipResolver  i.IPResolver
	echResolver i.ECHResolver

	dbLock sync.Mutex
	db     *gorm.DB

	vpnOn bool

	sshClientCacheLock sync.Mutex
	sshClientCache     map[string]*SshClientCacheItem

	// mac only
	sysProxy *sysproxy.SysProxy
}

type ApiOption func(*Api)

func WithDefaultInterfaceMonitor(mon i.DefaultInterfaceInfo) ApiOption {
	return func(a *Api) {
		a.mon = mon
	}
}

func StartApiServer(config *ApiServerConfig, options ...ApiOption) (*Api, error) {
	// set log
	logger, err := logger.SetLog(config.LogConfig)
	if err != nil {
		return nil, err
	}

	api := &Api{
		ApiServerConfig: config,
		sshClientCache:  make(map[string]*SshClientCacheItem),
		log:             logger,
	}
	for _, option := range options {
		option(api)
	}

	var lis net.Listener
	var opts []grpc.ServerOption
	_, _, err = net.SplitHostPort(config.ListenAddr)
	if err != nil { // unix addr
		listenAddr := config.ListenAddr
		log.Debug().Msgf("xapi listenAddr %s", listenAddr)
		os.Remove(listenAddr)
		lis, err = net.ListenUnix("unix", &net.UnixAddr{
			Name: listenAddr, Net: "unix"})
		if err != nil {
			return nil, err
		}
		err = os.Chown(listenAddr, os.Getuid(), os.Getgid())
		if err != nil {
			lis.Close()
			os.Remove(listenAddr)
			return nil, err
		}
	} else {
		lis, err = net.Listen("tcp", config.ListenAddr)
		if err != nil {
			return nil, err
		}
	}

	certificate, err := cert.Generate(nil,
		cert.NotBefore(time.Now().Add(-time.Hour*24*365)),
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
	tlsConfig := &gotls.Config{
		ClientCAs:  ca,
		ClientAuth: gotls.RequestClientCert,
		Certificates: []gotls.Certificate{
			cert,
		},
	}
	opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))

	// subscription
	if config.DbPath != "" {
		db, err := gorm.Open(sqlite.Open(config.DbPath),
			&gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect database: %w", err)
		}
		if runtime.GOOS != "android" {
			db.Exec("PRAGMA journal_mode = WAL")
		}
		db.Exec("PRAGMA foreign_keys = ON")
		db.Exec("PRAGMA busy_timeout = 5000")
		api.db = db
	}

	server := grpc.NewServer(opts...)
	api.server = server

	// in android, if app is added to vpn black list, all dns queries does not go through tun.
	nameServersDirect := []net.AddressPort{
		{
			Address: net.AliyunDns4,
			Port:    53,
		},
		{
			Address: net.CfDns4,
			Port:    53,
		},
	}
	if config.BindToDefaultNic {
		if api.mon == nil {
			mon, err := getNicMonitor(config.TunName)
			if err != nil {
				return nil, err
			}
			err = mon.(common.Startable).Start()
			if err != nil {
				return nil, err
			}
			api.mon = mon
		}

		dialer := transport.NewBindToDefaultNICDialer(api.mon, &dlhelper.SocketSetting{})
		freedomHandler := freedom.New(dialer, dialer, "direct", nil)
		dnsServer1 := idns.NewDnsServerConcurrent(idns.DnsServerConcurrentOption{
			NameserverAddrs: nameServersDirect,
			Name:            "api-dns-server",
			Handler:         freedomHandler,
			Dispatcher:      dispatcher.NewPacketDispatcher(context.Background(), freedomHandler),
		})
		resolver := idns.NewDnsServerToResolver(dnsServer1)
		api.echResolver = resolver
		api.ipResolver = resolver
		api.dialFactory = transport.NewDialerFactoryImp(transport.DialerFactoryOption{
			BindToDefaultNIC:        true,
			IpResolver:              api.ipResolver,
			DefaultInterfaceMonitor: api.mon,
		})
		dnsServer1.Start()
	} else {
		freedomHandler := freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "direct", nil)
		dnsServer1 := idns.NewDnsServerConcurrent(idns.DnsServerConcurrentOption{
			NameserverAddrs: nameServersDirect,
			Name:            "api-dns-server",
			Handler:         freedomHandler,
			Dispatcher:      dispatcher.NewPacketDispatcher(context.Background(), freedomHandler),
		})
		resolver := idns.NewDnsServerToResolver(dnsServer1)
		api.echResolver = resolver
		api.ipResolver = &idns.DnsResolver{}
		api.dialFactory = transport.DefaultDialerFactory()
		dnsServer1.Start()
	}

	RegisterApiServer(server, api)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Error().Err(err).Msg("failed to serve grpc")
		}
	}()
	return api, nil
}

func (a *Api) Stop() {
	if a.mon != nil {
		common.Close(a.mon)
	}
	if a.echResolver != nil {
		common.Close(a.echResolver)
	}
	if a.sysProxy != nil {
		a.sysProxy.Close()
	}
	a.server.Stop()
	a.log.Close()
}

func (a *Api) SetLog(ctx context.Context, req *vxlog.LoggerConfig) (*Receipt, error) {
	if a.log != nil {
		a.log.Close()
	}
	log, err := logger.SetLog(req)
	if err != nil {
		return nil, err
	}
	a.log = log
	return &Receipt{}, nil
}

func (a *Api) getDialerFactory() transport.DialerFactory {
	// since on android, when vpn is off, if try to bind to default nic, and other
	// vpn is on, all connections will fail. Therefore, when tm vpn is off (other vpn might be on),
	// do not try to bind to default nic
	// if runtime.GOOS == "android" {
	// 	if a.vpnOn {
	// 		return a.dialFactory
	// 	} else {
	// 		return transport.DefaultDialerFactory()
	// 	}
	// }
	return a.dialFactory
}

func (a *Api) getIPResolver() i.IPResolver {
	return a.ipResolver
}

func (a *Api) UpdateTmStatus(ctx context.Context, req *UpdateTmStatusRequest) (*Receipt, error) {
	a.vpnOn = req.On
	return &Receipt{}, nil
}

func (a *Api) GetDb() *gorm.DB {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()
	return a.db
}

func (a *Api) SetSecret(id uint32, secret string) {
	a.secretLock.Lock()
	defer a.secretLock.Unlock()
	a.secrets[id] = secret
}

func (a *Api) Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	return a.ApiDownload(req, a.getDialerFactory())
}

func (a *Api) HandlerUsable(ctx context.Context, req *HandlerUsableRequest) (*HandlerUsableResponse, error) {
	log.Debug().Msgf("HandlerUsable for: %v", util.GetTag(req.Handler))
	for i := 0; i < 3; i++ {
		rsp := a.HandlerTest(ctx, req)
		if rsp.Ping > 0 {
			return &rsp, nil
		}
	}
	return &HandlerUsableResponse{
		Ping: -1,
		Ip:   "",
	}, nil
}

func (a *Api) SpeedTest(req *SpeedTestRequest, in Api_SpeedTestServer) error {
	wg := new(errgroup.Group)
	url := util.SpeedtestURL1

	for _, t := range req.GetHandlers() {
		wg.Go(func() error {
			log.Debug().Msgf("SpeedTest for: %v", util.GetTag(t))

			rsp := &SpeedTestResponse{
				Tag: util.GetTag(t),
			}
			h, err := create.NewHandler(&create.HandlerConfig{
				HandlerConfig:               t,
				DialerFactory:               a.getDialerFactory(),
				Policy:                      policy.New(),
				IPResolver:                  a.getIPResolver(),
				EchResolver:                 a.echResolver,
				IPResolverForRequestAddress: a.getIPResolver(),
			})
			if err != nil {
				log.Debug().Err(err).Str("tag", util.GetTag(t)).Msg("failed to create outbound handler")
				rsp.Down = -1
			} else {
				logger := log.With().Uint32("sid", uint32(session.NewID())).
					Str("handler", util.GetTag(t)).Logger()
				logger.Debug().Msg("speed test start")
				ctx := logger.WithContext(in.Context())
				rst, err := util.Speedtest(ctx, url, h)
				if err != nil {
					log.Err(err).Msg("failed to speed test")
					rsp.Down = -1
				} else {
					rsp.Down = int32(rst)
				}
			}
			if err := in.Send(rsp); err != nil {
				log.Err(err).Msg("failed to send speed test response")
				return err
			}
			return nil
		})
	}
	return wg.Wait()
}

// func (a *Api) XStatusChangeNotify(ctx context.Context, req *XStatusChangeNotifyRequest) (*XStatusChangeNotifyResponse, error) {
// 	log.Debug().Msgf("xstatus change notify: %d", req.Status)
// 	switch req.Status {
// 	case XStatusChangeNotifyRequest_STATUS_START:
// 		a.subscriptionManager.Close()
// 	case XStatusChangeNotifyRequest_STATUS_STOP:
// 		a.subscriptionManager.Start()
// 	}
// 	return &XStatusChangeNotifyResponse{}, nil
// }

// func (a *Api) SetSubscriptionInterval(ctx context.Context, req *SetSubscriptionIntervalRequest) (*SetSubscriptionIntervalResponse, error) {
// 	if req.Interval > 0 {
// 		a.subscriptionManager.SetInterval(time.Duration(req.Interval) * time.Minute)
// 	}
// 	return &SetSubscriptionIntervalResponse{}, nil
// }

func (a *Api) UpdateSubscription(ctx context.Context, req *UpdateSubscriptionRequest) (*UpdateSubscriptionResponse, error) {
	log.Debug().Msg("UpdateSubscription")
	db := a.GetDb()
	if db == nil {
		return nil, errors.New("database not open")
	}

	handlers := make([]i.Outbound, 0, len(req.Handlers))
	// for android, if there are other vpn apps running, bind to default nic will not work, resulting in failed download.
	// therefore, add a plain handler first
	// if runtime.GOOS == "android" {
	// 	// add freedom handler
	// 	freedomHandler := freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "direct", &dns.DnsResolver{})
	// 	handlers = append(handlers, freedomHandler)
	// }
	for _, h := range req.Handlers {
		handler, err := create.NewHandler(&create.HandlerConfig{
			HandlerConfig: h,
			DialerFactory: a.getDialerFactory(),
			Policy:        policy.New(),
			IPResolver:    a.getIPResolver(),
			EchResolver:   a.echResolver,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to create outbound handler")
			continue
		}
		handlers = append(handlers, handler)
	}

	if req.All {
		result := subscription.UpdateSubscriptions(db, downloader.NewDownloader0(handlers))
		return &UpdateSubscriptionResponse{
			Success:      int32(result.SuccessSub),
			Fail:         int32(result.FailedSub),
			SuccessNodes: int32(result.SuccessNodes),
			FailedNodes:  result.FailedNodes,
			ErrorReasons: result.ErrorReasons,
		}, nil
	} else {
		var sub *xsqlite.Subscription
		db.First(&sub, req.Id)
		if sub == nil {
			return nil, fmt.Errorf("subscription not found: %d", req.Id)
		}
		successNodes, failedNodes, err := subscription.UpdateSubscription(sub, db, downloader.NewDownloader0(handlers))
		if err != nil {
			return &UpdateSubscriptionResponse{
				Success:      int32(0),
				Fail:         int32(1),
				SuccessNodes: int32(0),
				ErrorReasons: map[string]string{sub.Name: err.Error()},
			}, nil
		}
		return &UpdateSubscriptionResponse{
			Success:      int32(1),
			Fail:         int32(0),
			SuccessNodes: int32(successNodes),
			FailedNodes:  failedNodes,
		}, nil
	}
}

func (a *Api) FetchSubscriptionContent(ctx context.Context, req *FetchSubscriptionContentRequest) (*FetchSubscriptionContentResponse, error) {
	handlers := make([]i.Outbound, 0, len(req.Handlers))
	for _, h := range req.Handlers {
		handler, err := create.NewHandler(&create.HandlerConfig{
			HandlerConfig: h,
			DialerFactory: a.getDialerFactory(),
			Policy:        policy.New(),
			IPResolver:    a.getIPResolver(),
			EchResolver:   a.echResolver,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to create outbound handler")
			continue
		}
		handlers = append(handlers, handler)
	}
	result, err := subscription.FetchSubscription(ctx, req.Link, downloader.NewDownloader0(handlers))
	if err != nil {
		return nil, err
	}
	return &FetchSubscriptionContentResponse{
		Handlers:    result.Configs,
		FailedNodes: result.FailedNodes,
		Description: result.Description,
	}, nil
}

func (a *Api) Decode(ctx context.Context, req *DecodeRequest) (*DecodeResponse, error) {
	result, err := util.Decode(req.Data)
	if err != nil {
		return nil, err
	}
	return &DecodeResponse{
		Handlers:    result.Configs,
		FailedNodes: result.FailedNodes,
	}, nil
}

func (a *Api) DefaultNICHasGlobalV6(ctx context.Context, req *DefaultNICHasGlobalV6Request) (*DefaultNICHasGlobalV6Response, error) {
	if a.mon == nil {
		return nil, errors.New("default nic monitor not set")
	}
	has, err := a.mon.HasGlobalIPv6()
	if err != nil {
		return nil, err
	}
	return &DefaultNICHasGlobalV6Response{
		HasGlobalV6: has,
	}, nil
}

func (a *Api) CloseDb(ctx context.Context, req *CloseDbRequest) (*Receipt, error) {
	d := a.GetDb()
	if d == nil {
		return &Receipt{}, nil
	}

	db, err := d.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to close database: %w", err)
	}
	err = db.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close database: %w", err)
	}
	a.dbLock.Lock()
	defer a.dbLock.Unlock()
	a.db = nil
	return &Receipt{}, nil
}

func (a *Api) OpenDb(ctx context.Context, req *OpenDbRequest) (*Receipt, error) {
	db, err := gorm.Open(sqlite.Open(req.Path),
		&gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	if runtime.GOOS != "android" {
		db.Exec("PRAGMA journal_mode = WAL")
	}
	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA busy_timeout = 5000")

	a.dbLock.Lock()
	defer a.dbLock.Unlock()
	a.db = db
	return &Receipt{}, nil
}
