// Copyright 2025 5V Network LLC
// SPDX-License-Identifier: AGPL-3.0

package buildclient

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"

	"buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx"
	vxlog "buf.build/gen/go/vvvvv/vx/protocolbuffers/go/vx/log"
	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/create/inbound"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/fallbackmon"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/grpcserver"
	"github.com/5vnetwork/vx-core/app/grpcservice"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/app/memmon"
	outboundstats "github.com/5vnetwork/vx-core/app/outbound/stats"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/tester"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/app/util/downloader"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/i"
	"github.com/5vnetwork/vx-core/transport"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rs/zerolog/log"
)

type Option func(*Builder) error

func WithFeatures(features ...interface{}) Option {
	return func(i *Builder) error {
		for _, feature := range features {
			common.Must(i.addFeature(feature))
		}
		return nil
	}
}

func WithComponents(components ...interface{}) Option {
	return func(i *Builder) error {
		for _, component := range components {
			common.Must(i.addComponent(component))
		}
		return nil
	}
}

func NewX(config *vx.TmConfig, opts ...Option) (*client.Client, error) {
	builder := New()
	for _, opt := range opts {
		if err := opt(builder); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	x := &client.Client{
		Components: &common.Components{},
		OutStats:   outboundstats.NewOutStats(),
	}
	builder.addComponent(x.OutStats)

	// logger
	if config.GetLog() == nil {
		config.Log = &vxlog.LoggerConfig{
			LogLevel: vxlog.Level_DISABLED,
		}
	}
	l, err := logger.SetLog(config.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to set log: %w", err)
	}
	x.Logger = l

	if config.DefaultNicMonitor {
		err := Netmon(config, builder, x)
		if err != nil {
			return nil, fmt.Errorf("Netmon failed")
		}
		log.Debug().Msg("Netmon created")
	}

	// monitor
	if config.GetLog().GetLogLevel() == vxlog.Level_DEBUG {
		interval := time.Second * 1
		dir := path.Dir(config.RedirectStdErr)
		monitorConfig := &memmon.MonitorConfig{
			Interval:      interval,
			Path:          dir,
			ListenAddress: "0.0.0.0:6060",
		}
		if runtime.GOOS == "ios" {
			monitorConfig.Threshold = 25 * 1024 * 1024
		}
		monitor := memmon.NewMonitor(monitorConfig)
		builder.requireFeature(func(d *dispatcher.Dispatcher) {
			monitor.Dispatcher = d
		})
		common.Must(builder.addComponent(monitor))
	}

	size := 1000
	if runtime.GOOS == "ios" {
		size = 100
	}
	ul := userlogger.NewUserLogger(config.GetUserLog().GetEnable(),
		config.GetUserLog().GetLogAppId(), size)
	ul.SetLogSessionEnd(config.GetUserLog().GetLogSessionEnd())
	ul.SetLogRealtimeUsage(config.GetUserLog().GetLogRealtimeUsage())
	x.UserLogger = ul
	common.Must(builder.addComponent(ul))
	builder.requireOptionalFeatures(func(ipToDomain *dns.IPToDomain) {
		ul.SetDns(ipToDomain)
	})

	// dialer factory
	log.Print("NewDialerFactory")
	err = DialerFactory(config, builder, x)
	if err != nil {
		return nil, fmt.Errorf("failed to create dialer factory: %w", err)
	}

	// tun
	log.Print("Tun")
	err = Tun(config, builder, x)
	if err != nil {
		return nil, fmt.Errorf("failed to create tun : %w", err)
	}

	// wfp
	if config.Wfp != nil {
		log.Print("Wfp")
		err = Wfp(config.Wfp, builder)
		if err != nil {
			return nil, fmt.Errorf("failed to create wfp: %w", err)
		}
	}

	// inbound manager
	log.Print("NewInboundManager")
	im := proxy.NewManager()
	x.InboundManager = im
	for _, handlerConfig := range config.GetInboundManager().GetHandlers() {
		err := builder.requireFeature(func(ha *dispatcher.Dispatcher, df transport.DialerFactory, policy *policy.Policy, _ *dns.AllDnsServers) error {
			h, err := inbound.NewInbound(handlerConfig, ha, policy, x.IPResolver, df)
			if err != nil {
				return fmt.Errorf("failed to create inbound proxy handler: %w", err)
			}
			im.AddInbound(h)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	x.Inbounds = append(x.Inbounds, im)

	// outbound
	log.Debug().Msg("outbound")
	_, err = buildOutbound(config.Outbound, builder, x)
	if err != nil {
		return nil, err
	}
	common.Log()

	// dns
	log.Print("NewDNS")
	err = NewDNS(config, builder, x)
	if err != nil {
		return nil, fmt.Errorf("failed to create dns: %w", err)
	}
	common.Log()

	// policy
	p := create.NewPolicy(config.Policy)
	x.Policy = p
	if err := builder.addComponent(p); err != nil {
		return nil, fmt.Errorf("failed to add policy: %w", err)
	}

	// dispatcher
	log.Print("NewDispatcher")
	err = Handler(config, builder, x)
	if err != nil {
		return nil, fmt.Errorf("failed to create dispatcher: %w", err)
	}
	common.Log()

	// geo
	log.Print("NewGeo")
	gw := &geo.Geo{}
	x.Geo = gw
	if err := gw.UpdateGeo(config.Geo); err != nil {
		return nil, fmt.Errorf("failed to UpdateGeo: %w", err)
	}
	if err := builder.addComponent(gw); err != nil {
		return nil, fmt.Errorf("failed to add geo: %w", err)
	}
	common.Log()

	if config.SysProxy != nil && runtime.GOOS == "darwin" {
		err = NewSysProxy(config.SysProxy, builder)
		if err != nil {
			return nil, fmt.Errorf("NewSysProxy failed")
		}
	}
	if config.DbPath != "" {
		// PRAGMA foreign_keys and busy_timeout are per-connection in SQLite. Put
		// them in the DSN so every connection GORM opens in its pool inherits them
		// (a plain db.Exec("PRAGMA ...") after Open only configures one connection).
		db, err := gorm.Open(sqlite.Open(config.DbPath+"?_foreign_keys=on&_busy_timeout=5000"), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect database: %w", err)
		}
		if runtime.GOOS != "android" {
			db.Exec("PRAGMA journal_mode = WAL")
		}
		x.DB = &xsqlite.Database{DB: db}
		err = builder.addComponent(x.DB)
		if err != nil {
			return nil, fmt.Errorf("failed to add database: %w", err)
		}
		// subscription
		if config.Subscription != nil {
			builder.requireFeature(func(r i.Router) {
				sm := subscription.NewSubscriptionManager(
					&subscription.SubscriptionManagerConfig{
						Interval:   time.Duration(config.Subscription.Interval) * time.Minute,
						Db:         db,
						Downloader: downloader.NewDownloader(r),
						AutoUpdate: config.Subscription.PeriodicUpdate,
					})
				x.Subscription = sm
				builder.addComponent(sm)
			})
		}
	} else if config.ServicePort != 0 {
		db, err := xsqlite.NewDb(config.ServiceSecret, uint16(config.ServicePort))
		if err != nil {
			return nil, fmt.Errorf("failed to connect database: %w", err)
		}
		x.DB = db
		if err := builder.addComponent(db); err != nil {
			return nil, fmt.Errorf("failed to add database: %w", err)
		}
	}

	// tester
	t := &tester.Tester{
		SpeedTestFunc: func(ctx context.Context, h i.Outbound, size uint32) (int64, error) {
			return util.Speedtest(ctx, util.SpeedTestUrlBasic+strconv.Itoa(int(size)), h)
		},
		UsableTestFunc: func(ctx context.Context, h i.Outbound) (bool, error) {
			return util.ApiHandlerUsable1(ctx, h, util.UsableTestUrlCf)
		},
		PingTestFunc: func(ctx context.Context, h i.Outbound) (int, error) {
			return util.ApiHandlerPing(ctx, h, util.UsableTestUrlCf)
		},
	}
	x.Tester = t
	if err := builder.addComponent(t); err != nil {
		return nil, fmt.Errorf("failed to add tester: %w", err)
	}
	if err := builder.requireOptionalFeatures(func(reporter tester.ResultReporter) {
		t.ResultReporter = reporter
	}); err != nil {
		return nil, fmt.Errorf("failed to set result reporter: %w", err)
	}

	// fallback mon
	if config.FallbackMon != nil {
		err := builder.requireFeature(func(dispatcher *dispatcher.Dispatcher, geo *geo.Geo) error {
			fallbackMon := fallbackmon.NewFallbackMon(&fallbackmon.FallbackMonSetting{
				DomainSetName: config.FallbackMon.DomainSetName,
				Geo:           geo,
				LocalFile:     config.FallbackMon.LocalFile,
			})
			if err := builder.addComponent(fallbackMon); err != nil {
				return fmt.Errorf("failed to add fallback mon: %w", err)
			}
			builder.requireOptionalFeatures(func(db fallbackmon.Db) {
				fallbackMon.Db = db
			})
			dispatcher.AddOnFallback(fallbackMon)
			dispatcher.AddSessionEndHook(fallbackMon)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add fallback mon: %w", err)
		}
	}

	if config.Grpc != nil {
		grpc, err := grpcserver.NewGrpcServer(config.Grpc)
		if err != nil {
			return nil, fmt.Errorf("failed to create grpc server: %w", err)
		}
		if err := builder.addComponent(grpc); err != nil {
			return nil, fmt.Errorf("failed to add grpc server: %w", err)
		}
		x.GrpcServer = grpc
	}

	if config.GrpcService != nil {
		err := builder.requireFeature(func(grpcServer *grpcserver.GrpcServer) error {
			clientGrpc, _ := grpcservice.NewGrpcService(&grpcservice.GrpcServiceConfig{
				Client:         x,
				GrpcServer:     grpcServer.Server,
				UpdateLantency: config.GrpcService.UpdateLatency,
			})
			err := builder.addComponent(clientGrpc)
			if err != nil {
				return fmt.Errorf("failed to add grpc service: %w", err)
			}
			if x.Subscription != nil {
				x.Subscription.OnUpdatedCallback = clientGrpc.OnSubscriptionUpdated
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to add grpc service: %w", err)
		}
	}

	if !builder.resolved() {
		var missing []string
		for _, r := range builder.resolotions {
			if r.must {
				for _, d := range r.deps {
					if builder.getFeature(d) == nil {
						missing = append(missing, d.String())
						log.Error().Any("cb", r.callback).Msgf("missing %s", d.String())
					}
				}
			}
		}
		return nil, fmt.Errorf("not all features resolved: %v", missing)
	}

	for _, component := range builder.components {
		x.Components.AddComponent(component)
	}
	log.Info().Msg("NewX done")
	return x, nil
}

type Builder struct {
	rLock       sync.Mutex
	resolotions []resolution

	fLock    sync.RWMutex
	features []interface{}
	// components are those that needs to be started or closed
	components []interface{}
}

func New() *Builder {
	return &Builder{}
}

func (s *Builder) resolved() bool {
	s.fLock.RLock()
	defer s.fLock.RUnlock()

	for _, r := range s.resolotions {
		if r.must {
			return false
		}
	}
	s.resolotions = nil
	return true
}

func (s *Builder) getFeature(t reflect.Type) interface{} {
	s.fLock.RLock()
	defer s.fLock.RUnlock()

	for _, i := range s.features {
		if reflect.TypeOf(i) == t {
			return i
		}
	}
	for _, i := range s.features {
		if reflect.TypeOf(i).AssignableTo(t) {
			return i
		}
	}
	return nil
}

func (s *Builder) requireOptionalFeatures(callback interface{}) error {
	return s.requireFeatureCommon(callback, false)
}

func (s *Builder) requireFeature(callback interface{}) error {
	return s.requireFeatureCommon(callback, true)
}

func (s *Builder) requireFeatureCommon(callback interface{}, must bool) error {
	callbackType := reflect.TypeOf(callback)
	if callbackType.Kind() != reflect.Func {
		panic("not a function")
	}

	var featureTypes []reflect.Type
	for i := 0; i < callbackType.NumIn(); i++ {
		featureTypes = append(featureTypes, callbackType.In(i))
	}

	r := resolution{
		deps:     featureTypes,
		callback: callback,
		must:     must,
	}
	if r.canResolve(s) {
		return r.resolve(s)
	}
	s.rLock.Lock()
	s.resolotions = append(s.resolotions, r)
	s.rLock.Unlock()
	return nil
}

func (s *Builder) addComponent(component interface{}) error {
	s.components = append(s.components, component)
	return s.addFeature(component)
}

// addFeature registers a feature into current Instance.
func (s *Builder) addFeature(feature interface{}) error {
	s.fLock.Lock()
	s.features = append(s.features, feature)
	s.fLock.Unlock()

	s.rLock.Lock()
	if s.resolotions == nil {
		s.rLock.Unlock()
		return nil
	}
	var unResolvableResolutions []resolution
	var resolvableResolutions []resolution
	for _, r := range s.resolotions {
		if r.canResolve(s) {
			resolvableResolutions = append(resolvableResolutions, r)
		} else {
			unResolvableResolutions = append(unResolvableResolutions, r)
		}
	}
	s.resolotions = unResolvableResolutions
	s.rLock.Unlock()

	for _, r := range resolvableResolutions {
		err := r.resolve(s)
		if err != nil {
			return err
		}
	}
	return nil
}

type resolution struct {
	deps     []reflect.Type
	callback interface{}
	must     bool
}

func (r *resolution) canResolve(i *Builder) bool {
	for _, d := range r.deps {
		if i.getFeature(d) == nil {
			return false
		}
	}
	return true
}

// if all needed features are available, callback will be called, and return true and
// the err return by the callback
func (r *resolution) resolve(i *Builder) error {
	// check if all needed features are available
	var fs []interface{}
	for _, d := range r.deps {
		fs = append(fs, i.getFeature(d))
	}

	// rearrange the input parameters
	callback := reflect.ValueOf(r.callback)
	var input []reflect.Value
	callbackType := callback.Type()
	for i := 0; i < callbackType.NumIn(); i++ {
		pt := callbackType.In(i)
		for _, f := range fs {
			if reflect.TypeOf(f).AssignableTo(pt) {
				input = append(input, reflect.ValueOf(f))
				break
			}
		}
	}

	if len(input) != callbackType.NumIn() {
		panic("Can't get all input parameters")
	}

	var err error
	ret := callback.Call(input)
	errInterface := reflect.TypeOf((*error)(nil)).Elem()
	for i := len(ret) - 1; i >= 0; i-- {
		if ret[i].Type() == errInterface {
			v := ret[i].Interface()
			if v != nil {
				err = v.(error)
			}
			break
		}
	}
	return err
}
