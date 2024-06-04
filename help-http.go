package mybase

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crypto/md5"
	"encoding/json"
	"math"
	"sort"
	"strconv"
)

const (
	TimeFmt             = "2006/01/02 15:04:05.000" //毫秒保留3位有效数字
	TimeFmtDB           = "2006-01-02 15:04:05"     //写入数据库用的时间
	TimeFmtDB2          = "2006-01-02 15:04:05.000" //写入数据库用的时间 带上毫秒
	TimeFmtSeq          = "20060102150405"          //yyyyMMddHHmmss
	TimeFmtSeqHW        = "20060102150405000"
	DateFmtDB           = "2006-01-02" //写入数据库用的日期
	AesKey       string = "Aabc#123admin@12"
)

// Http 返回值中的code状态码
const (
	WGSuccess                  = iota //成功
	WGFail                            //请求失败
	WGErrorParam                      //参数非法
	WGErrorTime                       //时间戳非法
	WGErrorSign                       //签名错误
	WGErrorTip                        //提示msg信息
	WGErrorNeedLogin                  //需要重新登录
	WGErrorParse                      //json格式解析出错
	WGErrorNet                        //网络异常
	WGErrorDataBase                   //数据库操作失败
	WGErrorNoReg                      //尚未注册
	WGIPForbidden                     //该IP禁止访问
	WGErrorBusy                       //客户端提示请勿频繁操作 12
	WGErrorRegistered                 //已注册 13
	WGErrorNoImp                      //未实现 14
	WGErrorEmpty                      //当前库存为0
	WGNeedKefu                        //请求可能成功，但是需要客服进一步处理
	WGSuccessWithTipDeprecated        //成功处理，但是还需要异步回调通知。客户端可以先给提示。--该提示有被刷风险，尽量避免
	WGLimitToday                      //今日次数已达上限 18
	WGIdSuspend                       //账号封禁 19
	WGErrConf                         //配置异常 20
	WGErrStat                         //状态异常 21
	WGErrNoChance                     //次数不足 22
	WGErrServer                       //服务器异常 23
	WGErrData                         //数据解析失败 24
	WGErrBuyed                        //您已买过此类商品 25
	WGServerBusy                      //服务器繁忙 26
	WGErrorExtBegin            = 1000 //扩展错误码起始
)

var debugHttpReq = false
var timeout time.Duration = time.Second * 10

// 开启调试
func OnDebugHttpReq() {
	debugHttpReq = true
}

// 设置http请求超时时间，默认10s
func SetDefaultHttpTimeout(t time.Duration) {
	timeout = t
}

type HttpResult struct {
	Code int    `json:"code"` //状态码
	Msg  string `json:"msg"`  //信息
	Data any    `json:"data"` //数据结构
}

func UrlEncode(param string) string {
	return url.QueryEscape(param)

	//v := url.Values{}
	//v.Add("encode", param)
	//encoded := v.Encode()
	//return encoded[strings.Index(encoded, "=")+1:]
}

func BuildResultEx(w http.ResponseWriter, r *http.Request, result string) {
	//if r != nil && r.Method == http.MethodOptions { //支持跨域访问
	//解决egret跨域访问的问题
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//解决egret跨域访问head里面可以增加一些参数
	w.Header().Add("Access-Control-Allow-Headers", "curtime")
	w.Header().Add("Access-Control-Allow-Headers", "nonce")
	//}
	//fmt.Println("buildResult ", string(bs))

	//w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	//w.Header().Add("Access-Control-Allow-Credentials", "true")

	_, _ = fmt.Fprint(w, result)
}

// 构造http返回结果
func BuildResult(w http.ResponseWriter, r *http.Request, code int32, msg string, data any) {
	var result = map[string]any{}
	result["status"] = code //为了兼容老版本，后续会逐步去掉这个status，以code返回值为准
	result["code"] = code
	result["msg"] = msg
	if data != nil {
		result["data"] = data
	}

	bs, _ := json.Marshal(result)
	BuildResultEx(w, r, string(bs))
}

func BuildResult2(w http.ResponseWriter, r *http.Request, code int32, msg string) {
	BuildResult(w, r, code, msg, nil)
}

func BuildResult1(w http.ResponseWriter, r *http.Request, code int32) {
	BuildResult2(w, r, code, "")
}

func HttpGet(host, api string, param map[string]any, preFix, subFix string) (string, error) {
	return HttpGetEx(host, api, param, preFix, subFix, true)
}

func HttpGetEx(host, api string, param map[string]any, preFix, subFix string, needSign bool) (string, error) {
	return HttpGetUrl(fmt.Sprintf("%s/%s", host, api), param, preFix, subFix, needSign)
}

func HttpGetUrlNoSign(url string, param map[string]any) (string, error) {
	return HttpGetUrl(url, param, "", "", false)
}

func HttpGetUrl(httpUrl string, param map[string]any, preFix, subFix string, needSign bool) (string, error) {
	return HttpGetUrlEx(httpUrl, param, nil, preFix, subFix, needSign)
}

func HttpGetUrlEx(httpUrl string, param, customHead map[string]any, preFix, subFix string, needSign bool) (string, error) {
	values := url.Values{}
	if param != nil {
		for k, v := range param {
			values.Add(k, fmt.Sprintf("%v", v))
		}
	}
	return HttpGetUrlValues(httpUrl, values, customHead, preFix, subFix, needSign)
}

func HttpGetUrlValues(httpUrl string, query url.Values, customHead map[string]any, preFix, subFix string, needSign bool) (string, error) {
	var sign = ""
	var curTimeStr = ""
	var nonce = ""
	if needSign {
		var paramsMap = make(map[string]string)
		curTime := time.Now().Unix()
		curTimeStr = strconv.FormatInt(curTime, 10)
		paramsMap["curtime"] = curTimeStr

		nonce = GetRandomString(6)
		paramsMap["nonce"] = nonce
		var paramKeys []string
		paramKeys = append(paramKeys, "nonce")
		paramKeys = append(paramKeys, "curtime")
		for k, _ := range query {
			paramsMap[k] = query.Get(k)
			paramKeys = append(paramKeys, k)
		}
		sort.Strings(paramKeys)

		var paramSlice []string
		for i := range paramKeys {
			paramSlice = append(paramSlice, paramKeys[i]+"="+paramsMap[paramKeys[i]])
		}
		plainText := preFix + strings.Join(paramSlice, "&") + subFix
		sign = fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
		query.Add("sign", sign)
	}

	urlFull := httpUrl
	if len(query) > 0 {
		urlFull = urlFull + "?" + query.Encode()
	}

	//提交请求
	reqHttp, err := http.NewRequest("GET", urlFull, nil)
	if err != nil {
		W("HttpGetUrlValues new url=%s,err=%s", urlFull, err)
		return "", err
	}

	if needSign {
		//增加header选项
		reqHttp.Header.Set("curtime", curTimeStr)
		reqHttp.Header.Set("nonce", nonce)
	}

	if customHead != nil {
		for k := range customHead {
			reqHttp.Header.Set(k, fmt.Sprintf("%v", customHead[k]))
		}
	}

	//处理返回结果 10秒超时
	cli := http.Client{Timeout: timeout}
	response, err := cli.Do(reqHttp)
	//response, err := http.DefaultClient.Do(reqHttp)
	if err != nil {
		W("HttpGetUrlValues do url=%s,err=%s", urlFull, err)
		return "", err
	}
	bs, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		W("HttpGetUrlValues read url=%s,err=%s", urlFull, err)
		return "", err
	}

	if response.StatusCode != 200 {
		W("HttpGetUrlValues http(%d) url=%s", response.StatusCode, urlFull)
	}

	result := string(bs)
	if debugHttpReq {
		fmt.Printf("HttpGetUrlValues url=[%s],head=[%v] result=[%s]\n", urlFull, reqHttp.Header, result) //只在控制台打印一下。
	}
	return result, nil
}

func HttpPostJson(strURL string, params, heads map[string]any) (string, error) {
	theBody, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	return HttpPostJsonString(strURL, string(theBody), heads)
}

func HttpPostJsonString(strURL, bodyStr string, heads map[string]any) (string, error) {
	if heads == nil {
		heads = make(map[string]any)
	}

	if _, ok := heads["Content-Type"]; !ok {
		heads["Content-Type"] = "application/json"
	}
	return HttpPost(strURL, bodyStr, heads)
}

func HttpPostForm(strURL string, params, heads map[string]any) (string, error) {
	return HttpPostFormWithQuery(strURL, params, heads, nil)
}

func HttpPostFormWithQuery(strURL string, params, heads map[string]any, query url.Values) (string, error) {
	values := url.Values{}
	for k := range params {
		values.Add(k, fmt.Sprintf("%v", params[k]))
	}

	return HttpPostFormWithQuery2(strURL, values, heads, query)
}

func HttpPostFormWithQuery2(strURL string, params url.Values, heads map[string]any, query url.Values) (string, error) {
	if heads == nil {
		heads = make(map[string]any)
	}
	heads["Content-Type"] = "application/x-www-form-urlencoded"
	body := ""
	if params != nil {
		body = params.Encode()
	}
	return HttpPostWithQuery(strURL, body, heads, query)
}

// 直接调用这个HttpPost 需要指定heads["Content-Type"]
func HttpPost(strURL, body string, heads map[string]any) (string, error) {
	return HttpPostWithQuery(strURL, body, heads, nil)
}

// 直接调用这个HttpPost 需要指定heads["Content-Type"]
func HttpPostWithQuery(strURL, body string, heads map[string]any, query url.Values) (string, error) {
	urlFull := strURL
	if query != nil {
		urlFull = urlFull + "?" + query.Encode()
	}
	req, err := http.NewRequest("POST", urlFull, strings.NewReader(body))
	if err != nil {
		W("HttpPostWithQuery new url=%s,body=%s,heads=%+v,err=%v", strURL, body, heads, err)
		return "", err
	}

	if heads != nil {
		for k := range heads {
			req.Header.Add(k, fmt.Sprintf("%v", heads[k]))
		}
	}
	//10s超时
	cli := http.Client{Timeout: timeout}
	resp, err := cli.Do(req)
	if err != nil {
		W("HttpPostWithQuery do url=%s,body=%s,heads=%+v,err=%v", strURL, body, heads, err)
		return "", err
	}
	respBodyBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		W("HttpPostWithQuery read url=%s,body=%s,heads=%+v,err=%v", strURL, body, heads, err)
		return "", err
	}
	//控制台打印一下。
	if debugHttpReq {
		fmt.Printf("HttpPostWithQuery url=[%s],body=[%s],heads=[%v],result=[%s]\n", urlFull, body, heads, string(respBodyBytes))
	}
	return string(respBodyBytes), nil
}

func SortParam(param map[string]any) string {
	var paramKeys []string
	for k := range param {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)

	var paramSlice []string
	for i := range paramKeys { //+的效率最高，参见 help-http_test.go
		paramSlice = append(paramSlice, paramKeys[i]+"="+fmt.Sprintf("%v", param[paramKeys[i]]))
	}
	return strings.Join(paramSlice, "&")
}

func CheckHttpOptions(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		//如果POST请求，第一次是OPTIONS，第二次才是正常的POST，
		//修改"Content-Type: application/json" =》"Content-Type: application/x-www-form-urlencoded"
		fmt.Println("CheckHttpHeader http.MethodOptions")

		w.Header().Add("Allow", "GET,POST")

		BuildResult1(w, r, WGSuccess)
		return false
	}
	return true
}

func CheckHttpHeader(w http.ResponseWriter, r *http.Request, isProduct bool, preFix, subFix string) (bool, string) {
	var body, bodyMd5 string
	//@注意，如果不在ParseForm 之前读取数据的话，后面想在读取就没有了，ParseForm方法会去读取body数据。所以我们这里先把他读取掉。
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" { //如果是带有body的方法，我们先去读取一下body数据。
		bits, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		if len(bits) > 0 {
			body = string(bits)
			bodyMd5 = fmt.Sprintf("%x", md5.Sum(bits[:]))
		}
	}
	_ = r.ParseForm()

	debug := r.Form.Get("debug")
	if !isProduct && debug == "1" {
		return true, ""
	}

	if !CheckHttpOptions(w, r) {
		return false, ""
	}

	//获取传递的所有参数
	var paramsMap = make(map[string]string)
	var paramKeys []string
	for k, v := range r.Form {
		if k == "sign" { //跳过签名字段
			continue
		}
		paramsMap[k] = v[0]
		paramKeys = append(paramKeys, k)
	}

	curTime := r.Header.Get("curtime")
	paramKeys = append(paramKeys, "curtime")
	paramsMap["curtime"] = curTime
	nonce := r.Header.Get("nonce")
	paramKeys = append(paramKeys, "nonce")
	paramsMap["nonce"] = nonce

	if bodyMd5 != "" {
		paramKeys = append(paramKeys, "body")
		paramsMap["body"] = bodyMd5
	}
	//fmt.Println("CheckHttpHeader curTime=", curTime, "nonce=", nonce)

	sort.Strings(paramKeys)
	var plainText = ""
	for _, v := range paramKeys {
		plainText += v + "=" + paramsMap[v] + "&"
	}
	plainText = preFix + plainText[0:len(plainText)-1] + subFix
	var sign = fmt.Sprintf("%x", md5.Sum([]byte(plainText)))
	var signReq = strings.ToLower(r.Form.Get("sign"))

	//@注意：防止计时攻击 参见：https://coolshell.cn/articles/21003.html  sign!=signReq
	if subtle.ConstantTimeCompare([]byte(sign), []byte(signReq)) != 1 {
		//fmt.Printf("CheckHttpHeader md5[%s] %s==%s\n", plainText, sign, signReq)
		E("CheckHttpHeader md5[%s] %s==%s", plainText, sign, signReq)
		BuildResult1(w, r, WGErrorSign)
		return false, ""
	}

	i64CurTime, err := strconv.ParseInt(curTime, 10, 64)
	if err != nil { //时间格式错误
		E("CheckHttpHeader time[%s] err[%s] ", curTime, err.Error())
		BuildResult1(w, r, WGErrorParam)
		return false, ""
	}

	if math.Abs(float64(time.Now().Unix()-i64CurTime)) > float64(300) { //300秒 10分钟内，上下5分钟
		//超过有效期
		E("CheckHttpHeader time[%v] outdate", i64CurTime)
		BuildResult1(w, r, WGErrorTime)
		return false, ""
	}
	return true, body
}
