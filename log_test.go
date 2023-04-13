package mybase

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	if err := InitLogModule("./bin", "test", 3, false, logrus.TraceLevel, context.Background()); err != nil {
		t.Error(err)
		return
	}

	D("test log %s", "日志1")
	I("test log %s", "日志2")
	W("test log %s", "日志3")
	E("test log %s", "日志4")

	T("test log %s", "日志")

	t.Log("wait...")

	//由于使用chan处理多个goroutine过来的日志，没有sleep的话就没有日志写入文件了。。
	time.Sleep(time.Millisecond)
	t.Log("finish...")
}
