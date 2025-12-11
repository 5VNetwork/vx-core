package api

import (
	context "context"
	gotls "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	sync "sync"
	"time"

	idns "github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/sysproxy"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/transport/dlhelper"

	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/app/outbound"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/subscription/decode"
	"github.com/5vnetwork/vx-core/app/util/downloader"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/net"
	"github.com/5vnetwork/vx-core/common/protocol/tls/cert"
	"github.com/5vnetwork/vx-core/common/signal"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
	"github.com/5vnetwork/vx-core/tun"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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
	ipResolver i.IPResolver
	dnsServer  i.ECHResolver

	dbLock sync.Mutex
	db     *gorm.DB

	vpnOn bool

	sshClientCacheLock sync.Mutex
	sshClientCache     map[string]*SshClientCacheItem

	// mac only
	sysProxy *sysproxy.SysProxy
}

type ApiOption func(*Api)

func StartApiServer(config *ApiServerConfig, options ...ApiOption) (*Api, error) {
	// set log
	logger.SetLog(&configs.LoggerConfig{
		LogLevel:      configs.Level(config.LogLevel),
		ConsoleWriter: true,
		ShowCaller:    true,
	})

	api := &Api{
		ApiServerConfig: config,
		sshClientCache:  make(map[string]*SshClientCacheItem),
	}
	for _, option := range options {
		option(api)
	}

	var lis net.Listener
	var err error
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
	db, err := gorm.Open(sqlite.Open(config.DbPath),
		&gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	db.Exec("PRAGMA foreign_keys = ON")

	server := grpc.NewServer(opts...)
	api.server = server
	api.db = db

	// in android, if app is added to vpn black list, all dns queries does not go through tun.

	if config.BindToDefaultNic {
		if api.mon == nil {
			mon, err := tun.NewInterfaceMonitor(config.TunName)
			if err != nil {
				return nil, err
			}
			err = mon.Start()
			if err != nil {
				return nil, err
			}
			api.mon = mon
		}

		dialer := transport.NewBindToDefaultNICDialer(api.mon, &dlhelper.SocketSetting{})
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
		dnsServer1 := idns.NewDnsServerSerial(nameServersDirect, dialer, nil)
		api.dnsServer = idns.NewDnsServerToResolver(dnsServer1)
		api.ipResolver = idns.NewDnsServerToResolver(dnsServer1)
		api.dialFactory = transport.NewDialerFactoryImp(transport.DialerFactoryOption{
			BindToDefaultNIC:        true,
			IpResolver:              api.ipResolver,
			DefaultInterfaceMonitor: api.mon,
		})
		dnsServer1.Start()
	} else {
		api.ipResolver = &idns.DnsResolver{}
		api.dialFactory = transport.DefaultDialerFactory()
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
	if a.dnsServer != nil {
		common.Close(a.dnsServer)
	}
	if a.sysProxy != nil {
		a.sysProxy.Close()
	}
	a.server.Stop()
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
	log.Debug().Msgf("HandlerUsable for: %v", req.Handler.GetTag())
	for i := 0; i < 3; i++ {
		rsp := a.HandlerTest(ctx, req, util.TraceList[0])
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
	// if req.GetSize() == 1 {
	// 	url = SpeedtestURL1
	// }
	for _, t := range req.GetHandlers() {
		wg.Go(func() error {
			log.Debug().Msgf("SpeedTest for: %v", t.GetTag())

			rsp := &SpeedTestResponse{
				Tag: t.GetTag(),
			}
			h, err := outbound.NewHandler(&outbound.HandlerConfig{
				HandlerConfig: t,
				DialerFactory: a.getDialerFactory(),
				Policy:        policy.New(),
				IPResolver:    a.getIPResolver(),
				DnsServer:     a.dnsServer,
			})
			if err != nil {
				log.Debug().Err(err).Str("tag", t.GetTag()).Msg("failed to create outbound handler")
				rsp.Down = -1
			} else {
				rst := util.Speedtest(in.Context(), url, h)
				rsp.Down = int32(rst)
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
	handlers := make([]i.Outbound, 0, len(req.Handlers))
	// for android, if there are other vpn apps running, bind to default nic will not work, resulting in failed download.
	// therefore, add a plain handler first
	// if runtime.GOOS == "android" {
	// 	// add freedom handler
	// 	freedomHandler := freedom.New(transport.DefaultDialer, transport.DefaultPacketListener, "direct", &dns.DnsResolver{})
	// 	handlers = append(handlers, freedomHandler)
	// }
	for _, h := range req.Handlers {
		handler, err := outbound.NewHandler(&outbound.HandlerConfig{
			HandlerConfig: h,
			DialerFactory: a.getDialerFactory(),
			Policy:        policy.New(),
			IPResolver:    a.getIPResolver(),
			DnsServer:     a.dnsServer,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to create outbound handler")
			continue
		}
		handlers = append(handlers, handler)
	}

	db := a.GetDb()
	if db == nil {
		return nil, errors.New("database not open")
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

func (a *Api) Decode(ctx context.Context, req *DecodeRequest) (*DecodeResponse, error) {
	result, err := decode.Decode(req.Data)
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
	db.Exec("PRAGMA foreign_keys = ON")

	a.dbLock.Lock()
	defer a.dbLock.Unlock()
	a.db = db
	return &Receipt{}, nil
}
