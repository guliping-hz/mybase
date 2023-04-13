package mybase

import (
	"errors"
	"github.com/guliping-hz/mybase/net2"
)

var (
	ErrNoDB    = errors.New("db not init")
	ErrNoImp   = errors.New("no implementation")
	ErrNoRedis = errors.New("redis not init")

	ErrParse     = errors.New("parse error")
	ErrNoData    = errors.New("database/redis no data")
	ErrData      = errors.New("redis data error")
	ErrDataOld   = errors.New("data not new,need reget")    //你的数据不是最新的请重新获取
	ErrIsWriting = errors.New("data is writing by someone") //这个数据正在被某个服务器改写
	ErrAbort     = errors.New("abort")
	ErrTryMax    = errors.New("try max limit") //重试次数已达上限
	ErrOccur     = errors.New("err occur")     //发生了一次错误
	ErrInner     = errors.New("inner error")
	ErrThird     = errors.New("third platform error")

	//net2
	ErrTimeout = net2.ErrTimeout
	ErrParam   = net2.ErrParam
	ErrBuffer  = net2.ErrBuffer
	ErrClose   = net2.ErrClose
	ErrOOM     = net2.ErrOOM
)
