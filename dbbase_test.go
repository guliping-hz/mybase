package mybase

import (
	"context"
	"database/sql"
	"github.com/bytedance/sonic"
	"gorm.io/gorm"
	"os"
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

// 游戏服使用，后台用另一个
type IdCreateS4 struct {
	ID        uint           `gorm:"primarykey" json:"id"`    // 主键ID
	CreatedAt time.Time      `gorm:"index" json:"created_at"` // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`              // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 删除时间
}

type CoinLog struct {
	IdCreateS4
	Uid  int64      `gorm:"type:bigint;comment:UID" json:"uid,omitempty"`
	Coin int64      `gorm:"type:bigint;comment:金币" json:"coin"`
	Tm   *time.Time `gorm:"type:datetime;comment:时间" json:"tm"`
}

func (c *CoinLog) TableName() string {
	name := "coin_log_"
	if c.CreatedAt.IsZero() {
		return name + time.Now().Format(TimeSplitDay)
	}
	return name + c.CreatedAt.Format(TimeSplitDay)
}
func (c *CoinLog) IsBatch() bool {
	return true
}

func TestPatchCreate(t *testing.T) {
	//批量插入
	os.Setenv("db_dsn", "root:111000@tcp(127.0.0.1:3306)/test?charset=utf8mb4&loc=Local&parseTime=True")
	os.Setenv("redis_host", "127.0.0.1:6379")
	os.Setenv("redis_pwd", "111000")
	os.Setenv("redis_db", "0")

	imp := new(DBMgrBase)
	if err := imp.Init(context.Background(), 100, nil, "json", nil, &CoinLog{}); err != nil {
		t.Error(err)
		return
	}

	//t.Log(time.Now().Format(time.RFC3339))
	getTm := func(tm string) time.Time {
		tm2, err := time.Parse(time.RFC3339, tm)
		if err != nil {
			t.Error(err)
		}
		return tm2
	}

	t1, t2, t3, t4 := &CoinLog{
		IdCreateS4: IdCreateS4{CreatedAt: getTm("2024-09-04T23:50:05+08:00")},
		Uid:        1,
		Coin:       20000,
	}, &CoinLog{
		IdCreateS4: IdCreateS4{CreatedAt: getTm("2024-09-04T23:52:05+08:00")},
		Uid:        1,
		Coin:       19900,
	}, &CoinLog{
		IdCreateS4: IdCreateS4{CreatedAt: getTm("2024-09-05T00:52:05+08:00")},
		Uid:        1,
		Coin:       19000,
	}, &CoinLog{
		IdCreateS4: IdCreateS4{},
		Uid:        1,
		Coin:       18000,
	}

	//imp.GormDb.Create([]*CoinLog{t1, t2})

	//err := imp.GormDb.Table(t1.TableName()).Migrator().CreateTable(t1)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

	imp.Create(t1)
	imp.Create(t2)
	imp.Create(t3)
	imp.Create(t4)
	//imp.patchInsertAll("")

	time.Sleep(time.Second * 600)

	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-04T23:54:05+08:00")},
	//	Uid:       1,
	//	Coin:      30000,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-04T23:55:05+08:00")},
	//	Uid:       1,
	//	Coin:      29900,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-04T23:58:05+08:00")},
	//	Uid:       1,
	//	Coin:      29800,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-04T23:59:05+08:00")},
	//	Uid:       1,
	//	Coin:      29700,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-05T00:01:05+08:00")},
	//	Uid:       1,
	//	Coin:      29600,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-05T00:02:05+08:00")},
	//	Uid:       1,
	//	Coin:      39600,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-05T00:05:05+08:00")},
	//	Uid:       1,
	//	Coin:      39500,
	//})
	//imp.Create(&CoinLog{
	//	IdCreate4: IdCreate4{CreatedAt: getTm("2024-09-05T00:10:05+08:00")},
	//	Uid:       1,
	//	Coin:      39400,
	//})

	//time.Sleep(time.Second * 600)
}
