package mybase

import (
	"reflect"
)

type H map[string]interface{}

func (e H) GetInterface(key string) (interface{}, bool) {
	dataI, ok := e[key]
	return dataI, ok
}

func (e H) GetH(key string) (H, bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return nil, false
	}
	data, ok := dataI.(map[string]interface{})
	return H(data), ok
}

func (e H) GetInt64(key string) (ret int64, ok bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return 0, false
	}

	rV := reflect.ValueOf(dataI)
	if rV.Kind() == reflect.Int || rV.Kind() == reflect.Int8 || rV.Kind() == reflect.Int16 ||
		rV.Kind() == reflect.Int32 || rV.Kind() == reflect.Int64 {
		return rV.Int(), true
	}
	retF, ok := e.GetFloat64(key)
	if ok {
		return int64(retF), true
	}
	return 0, false
}

func (e H) GetInt32(key string) (ret int32, ok bool) {
	temp, ok := e.GetInt64(key)
	return int32(temp), ok
}

func (e H) GetInt(key string) (int, bool) {
	ret, ok := e.GetInt64(key)
	return int(ret), ok
}

func (e H) GetUInt64(key string) (uint64, bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return 0, false
	}

	rV := reflect.ValueOf(dataI)
	if rV.Kind() == reflect.Uint || rV.Kind() == reflect.Uint8 || rV.Kind() == reflect.Uint16 ||
		rV.Kind() == reflect.Uint32 || rV.Kind() == reflect.Uint64 || rV.Kind() == reflect.Uintptr {
		return rV.Uint(), true
	}
	v, ok := e.GetInt64(key)
	if ok {
		return uint64(v), true
	}
	return 0, false
}

func (e H) GetFloat64(key string) (float64, bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return 0, false
	}

	rV := reflect.ValueOf(dataI)
	if rV.Kind() == reflect.Float32 || rV.Kind() == reflect.Float64 {
		return rV.Float(), true
	}
	return 0, false
}

func (e H) GetString(key string) (string, bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return "", false
	}

	rV := reflect.ValueOf(dataI)
	if rV.Kind() == reflect.String {
		return rV.String(), true
	}
	return "", false
}

func (e H) GetBool(key string) (bool, bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return false, false
	}

	rV := reflect.ValueOf(dataI)
	if rV.Kind() == reflect.Bool {
		return rV.Bool(), true
	}
	return false, false
}

/**
@output 必须是跟存的类型保持一致，output必须是指针类型
*/
func (e H) Get(key string, output interface{}) bool {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return false
	}
	SameTransfer(dataI, output)
	return true
}

func (e H) Set(key string, val interface{}) {
	e[key] = val
}

func NewData() H {
	ret := make(H)
	return ret
}
