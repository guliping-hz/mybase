package mybase

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	GinKeyDingDSecret = "gin_key_dingd_secret"
	DingDingUrl       = "https://oapi.dingtalk.com/robot/send?access_token="
)

type ReqDingMsg struct {
	Token    string `json:"token" form:"token"`
	MsgTitle string `json:"title" form:"title"`
	MsgDing  string `json:"msg" form:"msg"`
	Phones   string `json:"phones" form:"phones"`
	All      int32  `json:"all" form:"all"`
}

type MsgDing struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
type At struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

func DingWarn(dingMsg *ReqDingMsg, secret string) (bool, string) {
	builder := &strings.Builder{}
	builder.WriteString(dingMsg.MsgDing)

	phones := strings.Split(dingMsg.Phones, ",")
	for i := range phones {
		_, _ = fmt.Fprintf(builder, "\n@%s", phones[i])
	}

	var param = struct {
		MsgType  string  `json:"msgtype"`
		Markdown MsgDing `json:"markdown"`
		At       At      `json:"at"`
	}{
		"markdown", MsgDing{
			dingMsg.MsgTitle,
			builder.String(),
		}, At{phones, dingMsg.All == 1},
	}
	buf, err := json.Marshal(param)
	if err != nil {
		return false, err.Error()
	}

	var sign, timeStamp string
	if secret != "" {
		timeStamp = strconv.FormatInt(time.Now().Unix()*1000, 10)
		plain := timeStamp + "\n" + secret
		buf := HMACSHA256Buf([]byte(plain), []byte(secret))
		sign = url.QueryEscape(base64.StdEncoding.EncodeToString(buf))
	}

	var urlDD string
	if sign == "" {
		urlDD = fmt.Sprintf("%s%s", DingDingUrl, dingMsg.Token)
	} else {
		urlDD = fmt.Sprintf("%s%s&timestamp=%s&sign=%s", DingDingUrl, dingMsg.Token, timeStamp, sign)
	}

	heads := make(map[string]any)
	heads["Content-Type"] = "application/json"
	body, err := HttpPost(urlDD, string(buf), heads)
	if err != nil {
		return false, err.Error()
	}
	return true, body
}

func DingWarnMidW(ctx *gin.Context) {
	dingMsg := ReqDingMsg{}
	_ = ctx.ShouldBind(&dingMsg)

	ok, result := DingWarn(&dingMsg, ctx.GetString(GinKeyDingDSecret))
	if !ok {
		AbortWithMsg(ctx, WGFail, result)
		return
	}

	AbortWithData(ctx, WGSuccess, result)
}

func GetDingWarnMidW(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(GinKeyDingDSecret, secret)
		DingWarnMidW(ctx)
	}
}
