package mybase

import (
	"reflect"
	"testing"
)

func TestWrap(t *testing.T) {
	a := uint8(8)
	aType := reflect.TypeOf(a)
	t.Logf("%v\n", aType.Kind())

	t.Log(WrapSql(`select * from usr where uid=? and uuid=?`, 1, "22"))
}
