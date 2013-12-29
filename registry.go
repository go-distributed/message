package message

import (
	"reflect"
)

var registry map[uint8]reflect.Type

func init() {
	registry = make(map[uint8]reflect.Type, 256)
}

func register(msgType uint8, t reflect.Type) {
	registry[msgType] = t
}
