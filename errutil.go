package mybase

import "errors"

var (
	ErrNoDB      = errors.New("db not init")
	ErrNoImp     = errors.New("no implementation")
	ErrNoRedis   = errors.New("redis not init")
	ErrParam     = errors.New("param error")
	ErrParse     = errors.New("parse error")
	ErrNoData    = errors.New("database/redis no data")
	ErrData      = errors.New("redis data error")
	ErrDataOld   = errors.New("data not new,need reget")    //你的数据不是最新的请重新获取
	ErrIsWriting = errors.New("data is writing by someone") //这个数据正在被某个服务器改写
	ErrAbort     = errors.New("abort")
	ErrTryMax    = errors.New("try max limit") //重试次数已达上限
	ErrOccur     = errors.New("err occur")     //发生了一次错误
	ErrBuffer    = errors.New("buffer error")
	ErrInner     = errors.New("inner error")
	ErrOOM       = errors.New("oom")
	ErrTimeout   = errors.New("time out")
	ErrClose     = errors.New("closed by the peer")
	ErrThird     = errors.New("third platform error")
)
