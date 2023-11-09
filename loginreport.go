package mybase

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type GovReportData struct {
	No    int    `json:"no"`           //编号
	Si    string `json:"si"`           //服务器内部用户ID
	Bt    int    `json:"bt"`           //0下线，1上线
	Ot    int64  `json:"ot"`           //时间戳，单位秒
	Ct    int    `json:"ct"`           //0已认证用户，2游客
	Di    string `json:"di,omitempty"` //设备唯一标识  ct=2传这个
	Pi    string `json:"pi,omitempty"` //用户唯一标识 ct=0传这个
	Debug bool   `json:"-"`
}

type GovReportDatas struct {
	Collections []*GovReportData `json:"collections"`
}

var (
	SmAppId     = ""
	SmBizId     = ""
	SmAppSecret = ""
	SmlUrl      = ""
)
var chanSig = make(chan *GovReportData)
var datas GovReportDatas
var chanEnd = make(chan bool)

func PostReportData(data *GovReportData) {
	go func() {
		chanSig <- data
	}()
}

func SendReport() {
	chanEnd <- true
	<-chanEnd
}

func reportData() {
	//上报数据
	heads := make(map[string]any)
	heads["Content-Type"] = "application/json;charset=utf-8"
	heads["appId"] = SmAppId
	heads["bizId"] = SmBizId
	heads["timestamps"] = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	bytes, err := json.Marshal(datas)
	if err != nil {
		return
	}

	cipherBytes, err := AESGCMEncrypt(SmAppSecret, string(bytes))
	if err != nil {
		return
	}
	body := fmt.Sprintf(`{"data":"%s"}`, base64.StdEncoding.EncodeToString(cipherBytes))
	plainText := fmt.Sprintf("%sappId%sbizId%stimestamps%s%s", SmAppSecret, SmAppId, SmBizId, heads["timestamps"], body)
	signByte := sha256.Sum256([]byte(plainText))
	heads["sign"] = hex.EncodeToString(signByte[:])

	ret, err := HttpPost(SmlUrl, body, heads)
	if err != nil {
		return
	}
	I("gov report result=%s", ret)
}

func InitReport(appId, bizId, appSecret, url string) {
	SmAppId = appId
	SmBizId = bizId
	SmAppSecret = appSecret
	SmlUrl = url

	datas.Collections = make([]*GovReportData, 0)
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for {
			select {
			case <-ticker.C:
				if len(datas.Collections) > 0 {
					reportData() //上报数据后清理所有数据
					datas.Collections = make([]*GovReportData, 0)
				}
			case data := <-chanSig:
				data.No = len(datas.Collections) + 1
				datas.Collections = append(datas.Collections, data)

				if data.Debug || data.No > 50 {
					reportData() //上报数据
					datas.Collections = make([]*GovReportData, 0)
				}
			case _ = <-chanEnd:
				if len(datas.Collections) > 0 {
					reportData() //上报数据后清理所有数据
					datas.Collections = make([]*GovReportData, 0)
				}
				chanEnd <- true
			}
		}
	}()
}
