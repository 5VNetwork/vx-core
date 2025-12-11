package mocks

import "reflect"

type FakeContainer struct {
	Interfaces []interface{}
}

func (c *FakeContainer) RequireFeatures(callback interface{}, must bool) error {
	cb := reflect.ValueOf(callback)
	var input []reflect.Value
	callbackType := cb.Type()
	for i := 0; i < callbackType.NumIn(); i++ {
		pt := callbackType.In(i)
		for _, f := range c.Interfaces {
			if reflect.TypeOf(f).AssignableTo(pt) {
				input = append(input, reflect.ValueOf(f))
				break
			}
		}
	}
	var err error
	ret := cb.Call(input)
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
