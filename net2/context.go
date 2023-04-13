package net2

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Context struct {
	Con      Conn     //连接对象
	socket   iSocket  //socket对象 仅net2内部使用
	OnSocket OnSocket //对socket的监听

	keys map[string]interface{}
	mu   sync.RWMutex

	readDB *bytes.Buffer

	once     sync.Once
	chanStop chan struct{}

	dataDecoder DataDecodeBase

	ttl  time.Duration //写超时
	rTtl time.Duration //读超时

	sessionId uint64
}

func (c *Context) String() string {
	return fmt.Sprintf("net2.Context sessionId=%d", c.SessionId())
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return c.chanStop
}

func (c *Context) Err() error {
	return nil
}

func (c *Context) Value(key interface{}) interface{} {
	keyStr, ok := key.(string)
	if !ok {
		return nil
	}
	v, _ := c.Get(keyStr)
	return v
}

func (c *Context) SessionId() uint64 {
	if c.sessionId == 0 {
		c.sessionId, _ = strconv.ParseUint(fmt.Sprintf("%p", c), 0, 64)
	}
	return c.sessionId
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}

	c.keys[key] = value
	c.mu.Unlock()
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.keys[key]
	c.mu.RUnlock()
	return
}

func (c *Context) GetEx(key string, output interface{}) bool {
	dataI, ok := c.Get(key)
	if !ok {
		return false
	}
	SameTransfer(dataI, output)
	return true
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint(key string) (ui uint) {
	if val, ok := c.Get(key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (c *Context) GetUint64(key string) (ui64 uint64) {
	if val, ok := c.Get(key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

func SameTransfer(input, outputPtr interface{}) {
	var vOE reflect.Value
	vO := reflect.ValueOf(outputPtr)
	if vO.Kind() != reflect.Ptr {
		panic("outputPtr must be ptr")
	}
	vOE = vO.Elem()

	var vIE reflect.Value
	vI := reflect.ValueOf(input)
	if vI.Kind() == reflect.Ptr {
		vIE = vI.Elem()
	} else {
		vIE = vI
	}

	if !vIE.CanConvert(vOE.Type()) {
		panic(fmt.Sprintf("the input and output not the same type %s != %s", vIE.Type().Name(), vOE.Type()))
	}
	vOE.Set(vIE)
}
