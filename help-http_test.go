package mybase

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var paramKeys []string
var paramsMap = make(map[string]string)

func init() {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		paramKeys = append(paramKeys, key)
		paramsMap[key] = fmt.Sprintf("value%d", i)
	}
}

func BenchmarkFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		paramSlice := make([]string, 0)
		for i := range paramKeys {
			paramSlice = append(paramSlice, fmt.Sprintf("%s=%s", paramKeys[i], paramsMap[paramKeys[i]]))
		}
		//plainText := strings.Join(paramSlice, "&") //对于数组字符拼接，最好用高效的方法。
	}
}

func BenchmarkBufferString(b *testing.B) {
	builder := strings.Builder{}
	for i := 0; i < b.N; i++ {
		paramSlice := make([]string, 0)
		for i := range paramKeys {
			builder.Reset()
			builder.Grow(len(paramKeys[i]) + 1 + len(paramsMap[paramKeys[i]]))
			builder.WriteString(paramKeys[i])
			builder.WriteString("=")
			builder.WriteString(paramsMap[paramKeys[i]])
			paramSlice = append(paramSlice, builder.String())
		}
		//plainText := strings.Join(paramSlice, "&") //对于数组字符拼接，最好用高效的方法。
	}
}

func BenchmarkBuffer(b *testing.B) {
	builder := bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		paramSlice := make([]string, 0)
		for i := range paramKeys {
			builder.Reset()
			builder.Grow(len(paramKeys[i]) + 1 + len(paramsMap[paramKeys[i]]))
			builder.WriteString(paramKeys[i])
			builder.WriteString("=")
			builder.WriteString(paramsMap[paramKeys[i]])
			paramSlice = append(paramSlice, builder.String())
		}
		//plainText := strings.Join(paramSlice, "&") //对于数组字符拼接，最好用高效的方法。
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		paramSlice := make([]string, 0)
		for i := range paramKeys {
			paramSlice = append(paramSlice, paramKeys[i]+"="+paramsMap[paramKeys[i]])
		}
		//plainText := strings.Join(paramSlice, "&") //对于数组字符拼接，最好用高效的方法。
	}
}

/*
基准性能分析
函数名称-内核     循环次数     运行纳秒/单次   分配内存/单次     分配内存次数/单次

goos: windows
goarch: amd64
pkg: util
BenchmarkFmt
BenchmarkFmt-4                     30537             39126 ns/op            8881
 B/op        308 allocs/op
BenchmarkBufferString
BenchmarkBufferString-4            47556             21727 ns/op            5680
 B/op        108 allocs/op
BenchmarkBuffer
BenchmarkBuffer-4                  45922             22740 ns/op            5680
 B/op        108 allocs/op
BenchmarkAdd
BenchmarkAdd-4                     88471             13404 ns/op            5680
 B/op        108 allocs/op
PASS
*/

//上面的性能测试表面，字符串 + 的效率最高。
