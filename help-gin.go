package mybase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	GinKeyRedisLockKey = "gin_key_redis_lock_key" //string
	GinKeyRedisLockCnt = "gin_key_redis_lock_cnt" //int64
	GinKeyRedisLockTtl = "gin_key_redis_lock_ttl" //time.Duration
)

type HelpRedisFirewall interface {
	RedisCheckFireWall(key string, limit int64, duration time.Duration) bool
	RedisDel(key string) bool
}

var (
	helpWall HelpRedisFirewall
)

func InitGinMidW(wall HelpRedisFirewall) {
	helpWall = wall
}

func Abort(ctx *gin.Context, code int32) {
	AbortEx(ctx, code, nil, "")
}

func AbortWithMsg(ctx *gin.Context, code int32, msg string) {
	AbortEx(ctx, code, nil, msg)
}

func AbortWithData(ctx *gin.Context, code int32, data interface{}) {
	AbortEx(ctx, code, data, "")
}

func AbortEx(ctx *gin.Context, code int32, data interface{}, msg string) {
	if ctx == nil {
		return
	}
	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{"code": code, "data": data, "msg": msg})
}

func EasyGet(ctx *gin.Context, key string, output interface{}) bool {
	dataI, ok := ctx.Get(key)
	if !ok {
		return false
	}
	SameTransfer(dataI, output)
	return true
}

func CrossMidW(ctx *gin.Context) {
	w := ctx.Writer
	method := ctx.Request.Method
	origin := ctx.Request.Header.Get("Origin") //请求头部
	if origin != "" {
		//接收客户端发送的origin （重要！）
		ctx.Header("Access-Control-Allow-Origin", origin)
		//服务器支持的所有跨域请求的方法
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		//允许跨域设置可以返回其他字段，可以自定义字段
		ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
		w.Header().Add("Access-Control-Allow-Headers", "utc")
		w.Header().Add("Access-Control-Allow-Headers", "nonce")
		w.Header().Add("Access-Control-Allow-Headers", "sign")
		w.Header().Add("Access-Control-Allow-Headers", "token")
		// 允许浏览器（客户端）可以解析的头部 （重要）
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
		//设置缓存时间
		ctx.Header("Access-Control-Max-Age", "172800")
		//允许客户端传递校验信息比如 cookie (重要)
		ctx.Header("Access-Control-Allow-Credentials", "true")
	}

	//允许类型校验
	if method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
	}

	ctx.Next()
}

// 互斥请求。防止多线程请求。请求完毕后会解锁。
// @注意：如果是防止暴力破解的话，需要在那个具体的业务上加锁。
func RedisLockMidW(ctx *gin.Context) {
	key := ctx.GetString(GinKeyRedisLockKey)
	if key == "" {
		key = "lock-" + ctx.Request.Method + "-" + ctx.Request.RequestURI + "-" + ctx.ClientIP()
	}
	limit := ctx.GetInt64(GinKeyRedisLockCnt)
	if limit == 0 {
		limit = 10
	}
	duration := ctx.GetDuration(GinKeyRedisLockTtl)
	if duration == 0 {
		duration = time.Minute
	}

	if helpWall == nil {
		Abort(ctx, WGErrorDataBase)
		return
	}

	//保护业务的原子性，一致性，完整性
	if !helpWall.RedisCheckFireWall(key, limit, duration) {
		Abort(ctx, WGErrorBusy)
		return
	}

	ctx.Next()

	//保护期结束
	helpWall.RedisDel(key)
}

// 互斥请求。防止多线程请求
func GetRedisLockMidW(key string, limit int64, ttl time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(GinKeyRedisLockKey, key)
		ctx.Set(GinKeyRedisLockCnt, limit)
		ctx.Set(GinKeyRedisLockTtl, ttl)
		RedisLockMidW(ctx)
	}
}

type CustomKey func(ctx *gin.Context) string

// 互斥请求。防止多线程请求
func GetRedisLockCustomMidW(f CustomKey, limit int64, ttl time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		newKey := f(ctx)
		if newKey == "" {
			Abort(ctx, WGFail)
			return
		}
		ctx.Set(GinKeyRedisLockKey, newKey)
		ctx.Set(GinKeyRedisLockCnt, limit)
		ctx.Set(GinKeyRedisLockTtl, ttl)
		RedisLockMidW(ctx)
	}
}
