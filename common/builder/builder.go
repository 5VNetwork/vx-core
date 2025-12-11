package builder

import (
	"reflect"
	"sync"

	"github.com/5vnetwork/vx-core/common"
)

type Builder struct {
	fLock       sync.RWMutex
	rLock       sync.Mutex
	resolotions []resolution
	components  *common.Components
}

func NewBuilder(c *common.Components) *Builder {
	return &Builder{
		components: c,
	}
}

func (s *Builder) Resolved() bool {
	for _, r := range s.resolotions {
		if r.must {
			return false
		}
	}
	s.resolotions = nil
	return true
}

func (s *Builder) GetFeature(t reflect.Type) interface{} {
	s.fLock.RLock()
	defer s.fLock.RUnlock()

	for _, i := range s.components.AllComponents() {
		if reflect.TypeOf(i) == t {
			return i
		}
	}
	for _, i := range s.components.AllComponents() {
		if reflect.TypeOf(i).AssignableTo(t) {
			return i
		}
	}
	return nil
}

func (s *Builder) RequireOptionalFeatures(callback interface{}) error {
	return s.requireFeatureCommon(callback, false)
}

func (s *Builder) RequireFeature(callback interface{}) error {
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

func (s *Builder) AddComponent(component interface{}) error {
	s.components.AddComponent(component)

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
		if i.GetFeature(d) == nil {
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
		fs = append(fs, i.GetFeature(d))
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
