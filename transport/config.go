package transport

import (
	"reflect"
)

var TransportNameToConfigType = make(map[string]reflect.Type)
var SecurityNameToConfigType = make(map[string]reflect.Type)
