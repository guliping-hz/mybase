package mybase

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// maxValue > 0
func GetRandom(maxValue int) int {
	if maxValue <= 0 {
		return 0
	}
	return RandInt(0, maxValue)
}

func GetRandomI32(maxValue int) int32 {
	return int32(GetRandom(maxValue))
}

func GetFullPath(filename string) (string, error) {
	filePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return filename, err
	}
	strings.Replace(filename, "\\", "/", -1)
	fullPath := filePath + "/" + filename
	return fullPath, nil
}

// 区间：[minValue,maxValue)
func RandInt(minValue, maxValue int) int {
	diff := maxValue - minValue
	ret, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 0
	}
	return int(ret.Int64()) + minValue
}

func ConvertVersion(version string) int64 {
	var versum int64
	s := strings.Split(version, ".")
	var l = len(s)
	if l == 2 {
		v0, _ := strconv.Atoi(s[0])
		v1, _ := strconv.Atoi(s[1])
		versum = (int64(v0) << 48) | (int64(v1) << 32)
	} else if l == 3 {
		v0, _ := strconv.Atoi(s[0])
		v1, _ := strconv.Atoi(s[1])
		v2, _ := strconv.Atoi(s[2])
		versum = (int64(v0) << 48) | (int64(v1) << 32) | (int64(v2) << 16)
	} else if l == 4 {
		v0, _ := strconv.Atoi(s[0])
		v1, _ := strconv.Atoi(s[1])
		v2, _ := strconv.Atoi(s[2])
		v3, _ := strconv.Atoi(s[3])
		versum = (int64(v0) << 48) | (int64(v1) << 32) | (int64(v2) << 16) | (int64(v3))
	}

	return versum
}

func CompileVer(ver1, ver2 string) int {
	s1 := strings.Split(ver1, ".")
	s2 := strings.Split(ver2, ".")

	for i := 0; i < len(s1) && i < len(s2); i++ {
		n1, _ := strconv.Atoi(s1[i])
		n2, _ := strconv.Atoi(s2[i])

		if n1 > n2 { //ver1 大
			return 1
		} else if n1 < n2 { //ver2 大
			return -1
		}
	}

	if len(s1) == len(s2) {
		return 0
	} else if len(s1) > len(s2) {
		return 1
	} else {
		return -1
	}
}

func EarthDistance(lat1, lng1, lat2, lng2 float64) float64 {
	radius := 6371000.0 // 6378137
	rad := math.Pi / 180.0

	lat1 = lat1 * rad
	lng1 = lng1 * rad
	lat2 = lat2 * rad
	lng2 = lng2 * rad

	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))

	return dist * radius
}

func GetRandomString(l int) string {
	bs := []byte("0123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ") //去掉l,I,O
	res := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		n := GetRandom(len(bs))
		res = append(res, bs[n])
	}
	return string(res)
}

func LoadCfg(filename string, cfg any) error {
	filePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		E("path err=%v", err)
		return err
	}
	fullPathFile := filePath + "/" + filename
	buf, err := ioutil.ReadFile(fullPathFile)
	if err != nil {
		E("LoadCfg ReadFile[%s]: %s", fullPathFile, err.Error())
		return err
	}

	if err := json.Unmarshal(buf, cfg); err != nil {
		E("LoadCfg Unmarshal error[%s]: %s", fullPathFile, err.Error())
		return err
	}

	return nil
}

// 获取到指定时间的0点的time.Time
func GetTodayMidnightEx(theTime *time.Time) time.Time {
	if theTime == nil {
		now := time.Now()
		theTime = &now
	}
	strTime := fmt.Sprintf("%04d-%02d-%02d 00:00:00", theTime.Year(), theTime.Month(), theTime.Day())
	midnight, err := time.ParseInLocation(TimeFmtDB, strTime, theTime.Location())
	if err != nil {
		E("err=%v", err)
		return *theTime
	}
	//fmt.Printf("midnight = %d\n", midnight.Unix())
	return midnight
}

// 获取到今天0点的time.Time
func GetTodayMidnight() time.Time {
	return GetTodayMidnightEx(nil)
}

// 明天0点
func GetTomorrowMidnight() time.Time {
	return GetTodayMidnightEx(nil).Add(time.Hour * 24)
}

/*
*
map[string]any ->数据结构
数据结构 -> map[string]any

@param input []map[string]any 或者 map[string]any 或者 结构
@param output 结构指针 或者 map指针
@param weakly 是否支持弱转换  比如 string=>int int=>string

@return nil无错误
*/
func Decode(input, outputPtr any, weakly bool) error {
	return DecodeEx(input, outputPtr, weakly, nil)
}

/*
*
map[string]any ->结构指针
相比于Decode；DecodeRedis会自动转换需要的数据类型；

	比如string转换成int。当然前提是该数据类型支持转换
*/
func DecodeRedis(input, outputPtr any) error {
	return DecodeEx(input, outputPtr, true, nil)
}

func DecodeDb(input, output any) error {
	return DecodeEx(input, output, true, func(src reflect.Type, dest reflect.Type, in interface{}) (interface{}, error) {
		//支持解析time.Time 转字符串
		if src.Kind() == reflect.Struct && src.String() == "time.Time" && dest.Kind() == reflect.String {
			if newIn, ok := in.(time.Time); ok {
				return newIn.Format(TimeFmtDB), nil
			}
		} else if src.Kind() == reflect.Ptr && src.String() == "*time.Time" && dest.Kind() == reflect.String {
			if newIn, ok := in.(*time.Time); ok {
				if newIn == nil {
					return "1970-07-01 00:00:00", nil
				}
				return newIn.Format(TimeFmtDB), nil
			}
		}
		return in, nil
	})
}

/*
*
@outputPtr 需要指针类型
*/
func DecodeEx(input, outputPtr any, weakly bool, hook mapstructure.DecodeHookFuncType) error {
	//dataType := reflect.TypeOf(outputPtr) //获取数据类型
	//if dataType.Kind() != reflect.Ptr {
	//	return fmt.Errorf("need Ptr")
	//}
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           outputPtr,
		TagName:          "json",
		WeaklyTypedInput: weakly,
		Squash:           true,
	}
	if hook != nil {
		config.DecodeHook = hook
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

/*
*
outputPtr 如果是 切片int数组，建议传int64
*/
func SameTransfer(input, outputPtr any) {
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
		vIEKind := vIE.Kind()
		vOEKind := vOE.Kind()
		if (vIEKind == reflect.Slice || vIEKind == reflect.Array) && (vOEKind == reflect.Slice || vOEKind == reflect.Array) {
			if vOE.Len() < vIE.Len() {
				vOE.Grow(vIE.Len() - vOE.Len())
				vOE.SetLen(vIE.Len())
			}

			for i := 0; i < vIE.Len(); i++ {
				vIEi := vIE.Index(i)
				vOEi := vOE.Index(i)

				//log.Println("vIEi type", vIEi.Kind().String())
				//log.Println("vOEi type", vOEi.Kind().String())

				if vIEi.Kind() == reflect.Interface {
					vIEi = vIEi.Elem()
				}

				//log.Println("vIEi type", vIEi.Kind().String())
				//log.Println("vOEi type", vOEi.Kind().String())

				if vIEi.CanConvert(vOEi.Type()) {
					vOEi.Set(vIEi.Convert(vOEi.Type()))
				} else {
					goto FAIL
				}
			}
			return
		}
	FAIL:
		panic(fmt.Sprintf("the input and output not the same type %s != %s", vIE.Type(), vOE.Type()))
	}
	vOE.Set(vIE.Convert(vOE.Type()))
}

func EasyGetMap(dict *sync.Map, key any, output any) bool {
	dataI, ok := dict.Load(key)
	if !ok {
		return false
	}
	SameTransfer(dataI, output)
	return true
}

/*
*
同类型指针 简单数值相加。目前仅支持整数 int64 及以内。
*/
func SameTypeAdd(dest, src any) {
	valPtr := reflect.ValueOf(dest)
	addPtr := reflect.ValueOf(src)

	val := valPtr.Elem()
	add := addPtr.Elem()
	typ := val.Type()

	if typ != add.Type() {
		panic("not the same type")
	}

	for i := 0; i < val.NumField(); i++ {
		fieldTyp := typ.Field(i)
		if !fieldTyp.IsExported() {
			continue
		}

		field := val.Field(i)
		addField := add.Field(i)
		field.SetInt(field.Int() + addField.Int())
	}
}

// 过滤 单引号 ，双引号，斜杠
func GetSafeUserInput(input string) string {
	output := strings.Replace(input, "'", "*", -1)  //替换单引号
	output = strings.Replace(output, "\"", "*", -1) //替换双引号
	output = strings.Replace(output, "\\", "*", -1) //替换斜杠
	return output
}

func GetRandSeed() int64 {
	var a = 0 //变量地址当做随机数
	var b = 0 //变量地址当做随机数
	aPtr, _ := strconv.ParseInt(fmt.Sprintf("%p", &a), 0, 64)
	bPtr, _ := strconv.ParseInt(fmt.Sprintf("%p", &b), 0, 64)

	return time.Now().Unix() * aPtr * bPtr
}
