package buildclient

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/5vnetwork/vx-core/app/client"
	"github.com/5vnetwork/vx-core/app/clientgrpc"
	"github.com/5vnetwork/vx-core/app/configs"
	"github.com/5vnetwork/vx-core/app/create"
	"github.com/5vnetwork/vx-core/app/dispatcher"
	"github.com/5vnetwork/vx-core/app/dns"
	"github.com/5vnetwork/vx-core/app/geo"
	"github.com/5vnetwork/vx-core/app/inbound/proxy"
	"github.com/5vnetwork/vx-core/app/logger"
	"github.com/5vnetwork/vx-core/app/memmon"
	"github.com/5vnetwork/vx-core/app/policy"
	"github.com/5vnetwork/vx-core/app/subscription"
	"github.com/5vnetwork/vx-core/app/tester"
	"github.com/5vnetwork/vx-core/app/userlogger"
	"github.com/5vnetwork/vx-core/app/util"
	"github.com/5vnetwork/vx-core/app/util/downloader"
	"github.com/5vnetwork/vx-core/app/xsqlite"
	"github.com/5vnetwork/vx-core/common"
	"github.com/5vnetwork/vx-core/common/signal/done"
	"github.com/5vnetwork/vx-core/i"
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

func NewX(config *configs.TmConfig, opts ...Option) (*client.Client, error) {
	builder := New()
	for _, opt := range opts {
		if err := opt(builder); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	x := &client.Client{
		Components:          &common.Components{},
		AllFakeDns:          &dns.AllFakeDns{},
		Hysteria2RejectQuic: config.Hysteria2RejectQuic,
	}
	builder.addComponent(x.AllFakeDns)
	x.FakeDnsEnabled.Store(config.Dns.GetEnableFakeDns())

	// logger
	if config.Log == nil {
		config.Log = &configs.LoggerConfig{LogLevel: configs.Level_ERROR, ConsoleWriter: true, ShowCaller: true}
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

	}

	// monitor
	if config.Log.LogLevel == configs.Level_DEBUG || runtime.GOOS == "ios" {
		interval := time.Second * 5
		if config.Log.LogLevel == configs.Level_DEBUG {
			interval = time.Second * 1
		}
		monitor := memmon.NewMonitor(interval)
		builder.requireFeature(func(d *dispatcher.Dispatcher) {
			monitor.Dispatcher = d
		})
		common.Must(builder.addComponent(monitor))
	}

	size := 1000
	if runtime.GOOS == "ios" {
		size = 100
	}
	ul := userlogger.NewUserLogger(config.Log.UserLog, config.Log.LogAppId, size)
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
		err := builder.requireFeature(func(ha *dispatcher.Dispatcher, policy *policy.Policy) error {
			h, err := proxy.NewInbound(handlerConfig, ha, policy)
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
	_, err = buildOutbound(config, builder, x)
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
	gw := &geo.GeoWrapper{}
	x.Geo = gw
	if err := gw.UpdateGeo(config.Geo); err != nil {
		return nil, fmt.Errorf("failed to UpdateGeo: %w", err)
	}
	common.Must(builder.addComponent(gw))
	common.Log()

	if config.SysProxy != nil {
		err = NewSysProxy(config.SysProxy, builder)
		if err != nil {
			return nil, fmt.Errorf("NewSysProxy failed")
		}
	}
	if config.DbPath != "" {
		db, err := gorm.Open(sqlite.Open(config.DbPath),
			&gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect database: %w", err)
		}
		db.Exec("PRAGMA foreign_keys = ON")
		x.DB = &xsqlite.Database{DB: db}
		err = builder.addComponent(x.DB)
		if err != nil {
			return nil, fmt.Errorf("failed to add database: %w", err)
		}
		// subscription
		if config.Subscription != nil {
			builder.requireFeature(func(r i.Router) {
				sm := subscription.NewSubscriptionManager(time.Duration(config.Subscription.Interval)*time.Minute,
					db, downloader.NewDownloader(r), subscription.WithPeriodicUpdate(config.Subscription.PeriodicUpdate))
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
	tester := &tester.Tester{
		SpeedTestFunc: func(ctx context.Context, h i.Outbound) (int64, error) {
			return util.Speedtest(ctx, util.SpeedtestURL1, h), nil
		},
		UsableTestFunc: func(ctx context.Context, h i.Outbound) (bool, error) {
			return util.ApiHandlerUsable1(ctx, h, util.TraceList[0])
		},
		PingTestFunc: func(ctx context.Context, h i.Outbound) (int, error) {
			return util.ApiHandlerPing(ctx, h, util.TraceList[0])
		},
	}
	x.Tetser = tester
	builder.addComponent(tester)

	if config.Grpc != nil {
		grpc := &clientgrpc.ClientGrpc{
			Done:           done.New(),
			GrpcConfig:     config.Grpc,
			Client:         x,
			UpdateLantency: config.UseRealLatency,
		}
		builder.addComponent(grpc)
		if x.Subscription != nil {
			subscription.WithOnUpdatedCallback(grpc.OnSubscriptionUpdated)(x.Subscription)
		}
		x.Tetser.ResultReporter = grpc
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
