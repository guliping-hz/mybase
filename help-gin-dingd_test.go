package mybase

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestDingWarn(t *testing.T) {
	r := gin.Default()
	r.POST("/dingd", GetDingWarnMidW("XXXXXX"))
	r.Run(":7000")
}
