package mybase

import (
	"database/sql"
	"github.com/bytedance/sonic"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"testing"
	"time"
)

func TestWrap(t *testing.T) {
	a := uint8(8)
	aType := reflect.TypeOf(a)
	t.Logf("%v\n", aType.Kind())

	t.Log(WrapSql(`select * from usr where uid=? and uuid=?`, 1, "2'2"))
}

type UsrBill struct {
	Uid       int64      `json:"uid"`
	Avatar    string     `json:"avatar"`
	Nickname  string     `json:"nickname"`
	Cost      int64      `json:"cost"`
	CreatedAt *time.Time `json:"created_at"`
}

func TestParseDbTime(t *testing.T) {
	var err error
	imp := new(DBMgrBase)
	if imp.DbInst, err = sql.Open("mysql", "test:111111@tcp(127.0.0.1:3306)/fish_game?charset=utf8mb4&loc=Local&parseTime=True"); err != nil {
		t.Error(err)
		return
	}

	if err = imp.DbInst.Ping(); err != nil {
		t.Error(err)
		return
	}

	bills := make([]*UsrBill, 0)
	my := make([]*UsrBill, 0)
	h := []any{&bills, &my}
	err = imp.CallQueryResultSets(h, `call proc_total_daily(?,?,?)`, time.Now().Format(TimeFmtDB), -1, 230600)
	if err != nil {
		t.Error(err)
	}

	s, _ := sonic.MarshalString(bills)
	t.Log(s)
	s, _ = sonic.MarshalString(my)
	t.Log(s)
}

func TestTimeToString(t *testing.T) {
	now := time.Now()
	input := H{
		"created_at": now,
	}
	output := struct {
		CreatedAt string `json:"created_at"`
	}{}

	if err := DecodeEx(input, &output, true, func(src reflect.Type, dest reflect.Type, in interface{}) (interface{}, error) {
		t.Logf("%v:%s %v:%s\n", src.Kind(), src.String(), dest.Kind(), dest.String())
		if src.Kind() == reflect.Struct && src.String() == "time.Time" && dest.Kind() == reflect.String {
			newIn := in.(time.Time)
			return newIn.Format(TimeFmtDB), nil
		} else if src.Kind() == reflect.Ptr && src.String() == "*time.Time" && dest.Kind() == reflect.String {
			newIn := in.(*time.Time)
			return newIn.Format(TimeFmtDB), nil
		}
		return in, nil
	}); err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", output)
}

func TestPatchCreate(t *testing.T) {
	//批量插入
}
