package mybase

import (
	"fmt"
	"reflect"
	"strconv"
)

type H map[string]any

func (e H) GetInterface(key string) (any, bool) {
	dataI, ok := e[key]
	return dataI, ok
}

func (e H) GetH(key string) (H, bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return nil, false
	}
	data, ok := dataI.(map[string]any)
	return H(data), ok
}

func (e H) ForceInt64(key string) int64 {
	if r, ok := e.GetInt64(key); !ok {
		if dataI, ok1 := e.GetInterface(key); !ok1 {
			return 0
		} else {
			rV := reflect.ValueOf(dataI)
			rVKind := rV.Kind()
			panic(fmt.Sprintf("%s can't convert to int64 kind:%s value:`%v`", key, rVKind.String(), dataI))
		}
	} else {
		return r
	}
}

func (e H) ForceString(key string) string {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return ""
	}

	v := e[key]
	rV := reflect.ValueOf(dataI)
	rVKind := rV.Kind()
	switch {
	case rV.CanInt(), rV.CanUint():
		return fmt.Sprintf("%d", v)
	case rV.CanFloat():
		return fmt.Sprintf("%.0f", v)
	case rVKind == reflect.String:
		return rV.String()
	}
	panic(fmt.Sprintf("%s can't convert to string kind:%s", key, rVKind.String()))
}

func (e H) GetInt64(key string) (ret int64, ok bool) {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return 0, false
	}

	rV := reflect.ValueOf(dataI)
	rVKind := rV.Kind()
	switch {
	case rV.CanInt():
		return rV.Int(), true
	case rV.CanUint():
		return int64(rV.Uint()), true
	case rV.CanFloat():
		return int64(rV.Float()), true
	case rVKind == reflect.String:
		vS, _ := e.GetString(key)
		if ret, err := strconv.ParseInt(vS, 10, 64); err == nil {
			return ret, true
		}
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

/*
*
@output 必须是跟存的类型保持一致，output必须是指针类型
*/
func (e H) Get(key string, output any) bool {
	dataI, ok := e.GetInterface(key)
	if !ok {
		return false
	}
	SameTransfer(dataI, output)
	return true
}

func (e H) Set(key string, val any) {
	e[key] = val
}

func NewData() H {
	ret := make(H)
	return ret
}
