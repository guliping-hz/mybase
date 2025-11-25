package mybase

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mohae/deepcopy"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//redis 缓存信息的版本号，从数据库第一次取出来放进去是0
	VerInRedis = "ver_in_redis"

	//redis写入标志位 请求写入的时候 incrBy 1,如果返回1表示可以写入，大于1表示已有其他服务器正在写入，要么等待要么直接返回
	WritingInRedis = "writing_in_redis"
)

var SqlRegExp *regexp.Regexp

// 此函数仅限服务器内部的高频率输入数据使用
func WrapSql(query string, args ...any) string {
	if len(args) == 0 {
		return query
	}

	i := 0
	return SqlRegExp.ReplaceAllStringFunc(query, func(s string) (arg string) {
		argV := args[i]

		argvType := reflect.TypeOf(argV)
		argvValue := reflect.ValueOf(argV)
		switch {
		case argvType.Kind() == reflect.String:
			//防止 argvValue 字符串中有单引号
			arg = fmt.Sprintf(`'%s'`, strings.Replace(argvValue.String(), `'`, `\'`, -1))
		case argvValue.CanInt():
			arg = fmt.Sprintf(`%v`, argvValue.Int())
		case argvValue.CanUint():
			arg = fmt.Sprintf(`%v`, argvValue.Uint())
		default:
			arg = fmt.Sprintf(`%v`, argV)
		}

		i++
		return
	})
}

const (
	TimeSplitDay   = "20060102"
	TimeSplitMonth = "200601"
	TimeSplitYear  = "2006"
)

type TableBatch interface {
	schema.Tabler
	IsBatch() bool
}

type PatchInsert struct {
	Logs  []map[string]any
	Model any
}

type GormWriter struct {
}

func (w *GormWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "Error") {
		C2(Custom2, p)
	} else {
		C(p)
	}
	return len(p), nil
}

type DBMgrBase struct {
	DbInst    *sql.DB
	GormDb    *gorm.DB
	RedisInst *redis.Client

	insertStmtCache sync.Map

	patchSqlMutex   sync.Mutex
	patchSqlCtx     context.Context
	patchSqlDict    map[string]*PatchInsert
	c               *cron.Cron
	createBatchSize int
	createBatchTag  string

	chanSignal chan bool
}

func (d *DBMgrBase) Init(ctx context.Context, maxDBCon int, config *gorm.Config, batchTag string, reloadF func(), dst ...any) error {
	var err error
	dsn := os.Getenv("db_dsn")

	redisHost := os.Getenv("redis_host")
	redisPwd := os.Getenv("redis_pwd")
	redisDb := 0
	if redisHost != "" {
		redisDb, err = strconv.Atoi(os.Getenv("redis_db"))
		if err != nil {
			return err
		}
	}

	fmt.Printf("Init Host=%s,Redis=%s[%s][%d]\n", dsn, redisHost, redisPwd, redisDb)

	if config == nil {
		config = &gorm.Config{
			Logger: logger.New(log.New(&GormWriter{}, "", 0), logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  false,
			}),
			CreateBatchSize: 1000,
		}
	}
	d.createBatchSize = config.CreateBatchSize
	if d.createBatchSize < 100 {
		d.createBatchSize = 100
	}
	d.createBatchTag = batchTag
	if d.createBatchTag == "" {
		d.createBatchTag = "json"
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return err
	}
	//只能迁移表。字段消失的话不会删掉旧的字段。很棒
	//&ActivityLog{}, &ServerCfg{}, &NlLog{}, &OnlineCountLog{}, &FlavorCfg{}, &ActivityCfg{},
	err = db.AutoMigrate(dst...)
	if err != nil {
		return err
	}

	d.GormDb = db

	//d.DbInst, err = sql.Open("mysql", dbFullPath)
	d.DbInst, err = db.DB()
	if err != nil {
		return err
	}

	if redisHost != "" {
		d.RedisInst = redis.NewClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPwd,
			DB:       redisDb, // use default DB
		})
	}

	//这里成功，并不能代表真的成功。。。，可能这个数据库服务器压根访问不到
	//所以我们这里尝试ping一下
	err = d.CheckDBConnectEx(redisHost != "")
	if err != nil {
		//fmt.Println("Init DB module error=", err)
		return err
	}

	//设置当前同时打开的最大连接数。
	d.DbInst.SetMaxOpenConns(maxDBCon)

	if reloadF != nil {
		reloadF()
		d.chanSignal = make(chan bool, 3) //同一个时间最多只有三个重新加载的通知
		go func() {
			for {
				select {
				case <-d.chanSignal:
					reloadF()
					time.Sleep(time.Second * 10) //10秒内响应最多1个请求
					I("reloadCfg")
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	var ctxCancelF context.CancelFunc
	d.patchSqlCtx, ctxCancelF = context.WithCancel(context.Background())
	d.patchSqlDict = make(map[string]*PatchInsert)
	d.c = cron.New(cron.WithSeconds())
	d.c.AddFunc("0 */2 * * * *", func() {
		d.patchInsertAll("")
	})
	d.c.Start()
	go func() {
		for {
			select {
			case <-ctx.Done():
				d.c.Stop()
				d.patchInsertAll("")
				ctxCancelF()
				return
			}
		}
	}()

	return err
}

func (d *DBMgrBase) UpdateCfg() bool {
	if d.chanSignal == nil {
		return false
	}

	I("UpdateCfg len(chanSignal)=%d", len(d.chanSignal))
	if len(d.chanSignal) > 1 { //同一时刻的通知太多了。。
		return false
	}
	d.chanSignal <- true
	return true
}

func (d *DBMgrBase) Wait() {
	<-d.patchSqlCtx.Done()
}

func (d *DBMgrBase) patchInsertAll(key string) {
	//fmt.Println("patchInsertAll")
	d.patchSqlMutex.Lock()
	defer d.patchSqlMutex.Unlock()

	insertF := func(k string, v *PatchInsert) {
		if strings.Contains(k, ".") {
			schemeAndTableName := strings.Split(k, ".")
			scheme := schemeAndTableName[0]
			tableName := schemeAndTableName[1]
			var cnt int64
			if d.GormDb.Raw(`SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`, scheme, tableName).Scan(&cnt); cnt == 0 {
				d.GormDb.Table(k).Migrator().CreateTable(v.Model)
			}
		} else {
			if !d.GormDb.Migrator().HasTable(k) {
				d.GormDb.Table(k).Migrator().CreateTable(v.Model)
			}
		}

		//panic报错  目前的Gorm不支持 []any  那么只能暂时先把log以 []map[string]any的形式存起来
		if tx := d.GormDb.Table(k).Model(v.Model).Create(v.Logs); tx.Error != nil {
			E("patchInsertAll key:%s err:%s", key, tx.Error.Error())
		}
		delete(d.patchSqlDict, k)
	}

	if key == "" {
		for k, v := range d.patchSqlDict {
			insertF(k, v)
		}
	} else {
		if v, ok := d.patchSqlDict[key]; ok {
			insertF(key, v)
		}
	}
}

func (d *DBMgrBase) Create(log any) (tx *gorm.DB) {
	return d.CreateLimit(log, d.createBatchSize)
}

func (d *DBMgrBase) CreateLimit(log any, limit int) (tx *gorm.DB) {
	if t, ok := log.(TableBatch); ok && t.IsBatch() {

		rVFirst := reflect.ValueOf(log)
		for rVFirst.Kind() == reflect.Ptr {
			rVFirst = rVFirst.Elem()
		}
		if rVFirst.Kind() != reflect.Struct {
			panic(fmt.Sprintf("need struct give %s", rVFirst.Kind().String()))
		}

		newV := make(map[string]any)
		var searchF func(reflect.Value)
		searchF = func(rV reflect.Value) {
			rT := rV.Type()
			for i := 0; i < rV.NumField(); i++ {
				rValField := rV.Field(i)
				rValType := rValField.Type()

				for rValField.Kind() == reflect.Ptr {
					if rValField.IsNil() {
						rValType = nil
						break
					} else {
						rValField = rValField.Elem()
						rValType = rValField.Type()
					}
				}

				rTypeField := rT.Field(i)
				name := rTypeField.Tag.Get("json")
				if name == "" {
					if rValType != nil && rValType.Kind() == reflect.Struct {
						searchF(rValField)
					}
					continue
				} else if name == "-" {
					continue
				}

				name = strings.TrimSpace(strings.Split(name, ",")[0])
				if name == "created_at" && rValField.IsZero() {
					newV[name] = time.Now()
				} else if name == "updated_at" {
					newV[name] = time.Now()
				} else {
					newV[name] = rValField.Interface()
				}
			}
		}
		searchF(rVFirst)

		key := t.TableName()

		d.patchSqlMutex.Lock()
		defer d.patchSqlMutex.Unlock()

		if _, ok := d.patchSqlDict[key]; ok {
			d.patchSqlDict[key].Logs = append(d.patchSqlDict[key].Logs, newV)
		} else {
			d.patchSqlDict[key] = &PatchInsert{
				Logs:  []map[string]any{newV},
				Model: log,
			}
		}

		if len(d.patchSqlDict[key].Logs) >= limit {
			//短时间累计很多数据，需要立刻插入
			go d.patchInsertAll(key)
			return nil
		}

		return nil
	} else {
		return d.GormDb.Create(log)
	}
}

func (d *DBMgrBase) CheckDBConnect() error {
	return d.CheckDBConnectEx(false)
}

func (d *DBMgrBase) CheckDBConnectEx(withRedis bool) error {
	//这里成功，并不能代表真的成功。。。，可能这个数据库服务器压根访问不到
	//所以我们这里尝试ping一下
	fmt.Println("Check DB...", d.DbInst)
	err := d.DbInst.Ping()
	if err != nil {
		fmt.Println("Check DB error=", err)
		return err
	}
	fmt.Println("Check DB OK")

	if withRedis && d.RedisInst != nil {
		_, err = d.RedisInst.Ping(d.RedisInst.Context()).Result()
		if err != nil {
			fmt.Println("Check Redis error=", err)
			return err
		}
		fmt.Println("Check Redis OK")
	}
	return nil
}

// 读取多个数据库表到 [][]map[string]any 中
func (d *DBMgrBase) LoadTableEx(query string, args ...any) ([][]map[string]any, error) {
	// 将数据填入mapUsrLv中
	stmt, err := d.DbInst.Prepare(query)
	if err != nil {
		E("query=%s, Prepare error=%s", WrapSql(query, args...), err.Error())
		return nil, err
	}
	defer stmt.Close()

	rawSql := WrapSql(query, args...)
	D("query=%s", rawSql)
	rows, err := stmt.Query(args...)
	if err != nil {
		E("query=%s, Query error=%s", rawSql, err.Error())
		return nil, err
	}
	defer rows.Close()

	result := make([][]map[string]any, 0)
	dealResult := func() error {
		cols, err := rows.Columns()
		if err != nil {
			E("query=%s, Columns error=%s", rawSql, err.Error())
			return err
		}

		colCnt := len(cols)
		tableData := make([]map[string]any, 0)
		values := make([]any, colCnt)
		valuesAddr := make([]any, colCnt)
		for i := range values {
			valuesAddr[i] = &values[i]
		}
		for rows.Next() {
			err = rows.Scan(valuesAddr...)
			if err != nil {
				E("query=%s, Scan error=%s", rawSql, err.Error())
				break
			}

			entry := make(map[string]any)
			for i, col := range cols {
				var v any
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					v = string(b)
				} else {
					v = val
				}
				entry[col] = v
			}
			tableData = append(tableData, entry)
		}
		result = append(result, tableData)
		return nil
	}

	err = dealResult()
	if err != nil {
		return nil, err
	}
	for rows.NextResultSet() {
		err = dealResult()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// 读取单个数据库表到 []map[string]any 中
func (d *DBMgrBase) LoadTable(query string, args ...any) ([]map[string]any, error) {
	// 将数据填入mapUsrLv中
	stmt, err := d.DbInst.Prepare(query)
	if err != nil {
		E("query=%s, Prepare error=%s", WrapSql(query, args...), err.Error())
		return nil, err
	}
	defer stmt.Close()

	rawSql := WrapSql(query, args...)
	//D("query=%s", rawSql)
	rows, err := stmt.Query(args...)
	if err != nil {
		E("query=%s, Query error=%s", rawSql, err.Error())
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		E("query=%s, Columns error=%s", rawSql, err.Error())
		return nil, err
	}

	col_cnt := len(cols)
	tableData := make([]map[string]any, 0)
	values := make([]any, col_cnt)
	valuePtrs := make([]any, col_cnt)
	for rows.Next() {
		for i := 0; i < col_cnt; i++ {
			valuePtrs[i] = &values[i]
		}
		err = rows.Scan(valuePtrs...)
		if err != nil {
			E("query=%s, Scan error=%s", rawSql, err.Error())
			break
		}

		entry := make(map[string]any)
		for i, col := range cols {
			var v any
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	return tableData, nil
}

func (d *DBMgrBase) GetCnt(query string, args ...any) int64 {
	cnt, _ := d.GetCntEx(query, args...)
	return cnt
}

func (d *DBMgrBase) GetCntEx(query string, args ...any) (int64, bool) {
	tableData, err := d.LoadTable(query, args...)
	if err != nil {
		return 0, false
	}

	if len(tableData) == 0 {
		return 0, true
	}

	row := tableData[0]
	data := H(row)
	cnt, _ := data.GetInt64("cnt")
	return cnt, true
}

func (d *DBMgrBase) GetSum(query string, args ...any) int64 {
	var sum int64
	tableData, err := d.LoadTable(query, args...)
	if err == nil {
		if len(tableData) > 0 {
			row := tableData[0]
			data := H(row)
			sum, _ = data.GetInt64("s")
		}
	}
	return sum
}

func (d *DBMgrBase) GetSumFloat64(query string, args ...any) float64 {
	var sum float64
	tableData, err := d.LoadTable(query, args...)
	if err == nil {
		if len(tableData) > 0 {
			row := tableData[0]
			data := H(row)
			sum, _ = data.GetFloat64("s")
		}
	}
	return sum
}

// the v must be a pointer to a map or struct.
func (d *DBMgrBase) SelectObject(v any, query string, args ...any) error {
	return d.selectObject(true, v, query, args...)
}
func (d *DBMgrBase) SelectObjectNoWarn(v any, query string, args ...any) error {
	return d.selectObject(false, v, query, args...)
}

/*
*
@param v 必须是数据结构指针.
*/
func (d *DBMgrBase) selectObject(logWarn bool, v any, query string, args ...any) error {
	dataType := reflect.TypeOf(v) //获取数据类型
	if dataType.Kind() != reflect.Ptr {
		E("query=%s, Kind error=need Ptr", WrapSql(query, args...))
		return ErrParam
	}

	tableData, err := d.LoadTable(query, args...)
	if err != nil {
		return err
	}

	if len(tableData) > 0 {
		row := tableData[0]
		if err = DecodeDb(row, v); err != nil {
			E("query=%s, DecodePath error=%s", WrapSql(query, args...), err.Error())
			return err
		} else {
			return nil
		}
	} else {
		if logWarn {
			W("query=%s, error=No Data", WrapSql(query, args...))
		}
		return ErrNoData
	}
}

/*
*
跟SelectObject比，这个SelectObjectsEx是返回一个数组的。

@param v 必须是存放map或是struct的数组的指针.
*/
func (d *DBMgrBase) SelectObjectsEx(v any, query string, args ...any) error {
	return d.SelectObjectsEx2(true, v, query, args...)
}
func (d *DBMgrBase) SelectObjectsExNoWarn(v any, query string, args ...any) error {
	return d.SelectObjectsEx2(false, v, query, args...)
}
func (d *DBMgrBase) SelectObjectsEx2(logWarn bool, v any, query string, args ...interface{}) error {
	tableData, err := d.LoadTable(query, args...)
	if err != nil {
		return err
	}

	if len(tableData) > 0 {
		if err = DecodeDb(tableData, v); err != nil {
			E("query=%s, DecodePath error=%s", WrapSql(query, args...), err.Error())
			return err
		} else {
			//fmt.Printf("SelectObjectsEx DecodePath obj=%v\n", v)
			return nil
		}
	} else {
		if logWarn {
			W("query=%s, error=No Data", WrapSql(query, args...))
		}
		return ErrNoData
	}
}

func (d *DBMgrBase) Update(query string, args ...interface{}) bool {
	return d.Update2(true, query, args...) > 0
}
func (d *DBMgrBase) UpdateNoWarn(query string, args ...interface{}) bool {
	return d.Update2(false, query, args...) >= 0
}

// 通用的update方法
func (d *DBMgrBase) Update2(logWarn bool, query string, args ...interface{}) int64 {
	//I("Update,query=[%s] args=%v", query, args)
	stmt, err := d.DbInst.Prepare(query)
	if err != nil {
		E("Update=%s, Prepare error=%s", WrapSql(query, args...), err.Error())
		return -1
	}
	defer stmt.Close()

	r, err := stmt.Exec(args...)
	if err != nil {
		E("Update=%s, Exec error=%s", WrapSql(query, args...), err.Error())
		return -1
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		E("Update=%s, RowsAffected error=%s", WrapSql(query, args...), err.Error())
		return -1
	} else if rowsAffected == 0 && logWarn {
		W("Update=%s, RowsAffected is 0", WrapSql(query, args...))
	}
	return rowsAffected
}

// Insert 通用方法
func (d *DBMgrBase) Insert(query string, args ...interface{}) (int64, error) {
	insertId, err, _ := d.InsertEx(query, args...)
	return insertId, err
}

func (d *DBMgrBase) InsertEx(query string, args ...interface{}) (int64, error, bool) {
	return d.InsertExWithLastId(true, query, args...)
}

func (d *DBMgrBase) InsertExWithLastId(lastId bool, query string, args ...interface{}) (int64, error, bool) {
	return d.InsertExWithLastIdEx(true, lastId, query, args...)
}

func (d *DBMgrBase) InsertExWithLastIdEx(log, lastId bool, query string, args ...interface{}) (int64, error, bool) {
	//缓存stmt模式
	stmtV, ok := d.insertStmtCache.Load(query)
	var stmt *sql.Stmt
	if !ok {
		stmtNew, err := d.DbInst.Prepare(query)
		if err != nil {
			E("InsertEx=%s,Prepare error=%s", WrapSql(query, args...), err.Error())
			return -1, err, false
		}
		stmt = stmtNew
		d.insertStmtCache.Store(query, stmt) //缓存起来
	} else {
		stmt = stmtV.(*sql.Stmt)
	}

	//非缓存模式
	//stmt, err := d.DbInst.Prepare(query)
	//if err != nil {
	//	E("InsertEx=%s,Prepare error=%s", WrapSql(query, args...), err.Error())
	//	return -1, err, false
	//}
	//defer stmt.Close()

	r, err := stmt.Exec(args...)
	if err != nil {
		if log {
			E("InsertEx=%s,Exec error=%s", WrapSql(query, args...), err.Error())
		}
		return -1, err, false
	}
	n, err := r.RowsAffected()
	if err != nil {
		E("InsertEx=%s,RowsAffected error=%s", WrapSql(query, args...), err.Error())
		return -1, err, false
	}
	if lastId {
		id, err := r.LastInsertId()
		if err != nil {
			E("InsertEx=%s,LastInsertId error=%s", WrapSql(query, args...), err.Error())
			return -1, err, false
		}
		return id, nil, n > 0
	}
	return -1, nil, n > 0
}

// 执行存储过程或者多个sql语句
func (d *DBMgrBase) CallExec(query string, args ...interface{}) (int64, error) {
	stmt, err := d.DbInst.Prepare(query)
	if err != nil {
		E("CallExec=%s,Prepare error=%s", WrapSql(query, args...), err.Error())
		return 0, err
	}
	defer stmt.Close()

	rawSql := WrapSql(query, args...)
	D("CallExec=%s", rawSql)
	result, err := stmt.Exec(args...)
	if err != nil {
		E("CallExec=%s,Exec error=%s", rawSql, err.Error())
		return 0, err
	}
	return result.LastInsertId()
}

// 执行存储过程或者多个sql语句
// @insert 是否需要返回最后的插入自增ID
func (d *DBMgrBase) CallExecNoStmt(rawSql string, insert bool) (int64, error) {
	return d.CallExecNoStmtWithLog(true, rawSql, insert)
}

// @insert 是否需要返回最后的插入自增ID
func (d *DBMgrBase) CallExecNoStmtWithLog(log bool, rawSql string, insert bool) (int64, error) {
	//D("CallExecNoStmt=%s", rawSql)
	result, err := d.DbInst.Exec(rawSql)
	if err != nil {
		if log {
			E("CallExec=%s,Exec error=%s", rawSql, err.Error())
		}
		return 0, err
	}
	if insert {
		return result.LastInsertId()
	} else {
		return result.RowsAffected()
	}
}

/*
*
@param v 必须是指针类型 *Struct1{}
*/
func (d *DBMgrBase) CallQuery(v interface{}, query string, args ...interface{}) error {
	return d.SelectObject(v, query, args...)
}

// 打印一下日志，适用于重要的过程调用。
func (d *DBMgrBase) CallQueryWithLog(v interface{}, query string, args ...interface{}) error {
	I("CallQuery query=%s", WrapSql(query, args...))
	return d.CallQuery(v, query, args...)
}

/*
*
@param v 必须是指针类型 //*[]Struct1{}
*/
func (d *DBMgrBase) CallQuerys(v interface{}, query string, args ...interface{}) error {
	return d.SelectObjectsEx(v, query, args...)
}

/*
*
存储过程，返回多个数据表的信息,每个表信息需要单独存放。

多张表返回，每张表含有多行数据
@v 存储数据结构指针，每个位置对应了存储过程返回的指定位置的表结构信息 是一个二维数组结构指针 类似  *[*[]Struct1,*[]Struct2,*[]Struct3...]

如果只需要取第一张表的数据，可以直接调用CallQuery或CallQuerys
*/
func (d *DBMgrBase) CallQueryResultSets(v []interface{}, query string, args ...interface{}) error {
	resultSets, err := d.LoadTableEx(query, args...)
	if err != nil {
		return err
	}

	for i := range v {
		if i < len(resultSets) {
			if err = DecodeDb(resultSets[i], v[i]); err != nil {
				E("query=%s, DecodeDb error=%s", WrapSql(query, args...), err.Error())
				return err
			}
		}
	}
	return nil
}

// 多张表返回，每张表只取一行数据
// @v 是一个数组结构指针 类似  *[*Struct1,*Struct2,*Struct3...]
func (d *DBMgrBase) CallQueryResultSetsOnlyFirst(v []interface{}, query string, args ...interface{}) error {
	resultSets, err := d.LoadTableEx(query, args...)
	if err != nil {
		return err
	}

	for i := range v {
		if i < len(resultSets) {
			if len(resultSets[i]) == 0 {
				return ErrNoData
			}
			if err = DecodeDb(resultSets[i][0], v[i]); err != nil {
				E("query=%s, DecodeDb error=%s", WrapSql(query, args...), err.Error())
				return err
			}
		}
	}
	return nil
}

/**
****************************************************************************************
redis opt
*/

func (d *DBMgrBase) RedisExists(key string) bool {
	if d.RedisInst == nil {
		panic("Redis Instance is nil")
	}

	sc := d.RedisInst.Exists(d.RedisInst.Context(), key)
	//D("RedisExists sc=%v", sc)
	exits, err := sc.Result()
	if err != nil { //报错默认存在。
		return true
	}
	return exits == 1
}

func (d *DBMgrBase) RedisDel(key string) bool {
	ok, _ := d.RedisInst.Del(d.RedisInst.Context(), key).Result()
	return ok == 1
}

/*
*
@expiration 0 永久有效 ,-1也是永久有效 但是需要redis>6.0
*/
func (d *DBMgrBase) RedisSetEx(key string, value interface{}, expiration time.Duration) (bool, error) {
	ok, err := d.RedisInst.Set(d.RedisInst.Context(), key, value, expiration).Result()
	return ok == "OK", err
}

func (d *DBMgrBase) RedisSet(key string, value interface{}) (bool, error) {
	return d.RedisSetEx(key, value, 0)
}

func (d *DBMgrBase) RedisGet(key string) (string, error) {
	if !d.RedisExists(key) {
		return "", ErrNoData
	}

	return d.RedisInst.Get(d.RedisInst.Context(), key).Result()
}

func (d *DBMgrBase) RedisIncrBy(key string, incr int64) (int64, error) {
	return d.RedisInst.IncrBy(d.RedisInst.Context(), key, incr).Result()
}

/*
*
多个服务器分布式，读取redis中的标志位
*/
func (d *DBMgrBase) RedisIncrByFlagAdd(key string) (int64, error) {
	return d.RedisIncrBy(key, 1)
}
func (d *DBMgrBase) RedisIncrByFlagCheck(key string) (int64, error) {
	return d.RedisIncrBy(key, 0)
}
func (d *DBMgrBase) RedisIncrByFlagRelease(key string) (int64, error) {
	return d.RedisIncrBy(key, -1)
}

/*
*
不存在的key或者field，redis中默认存0
*/
func (d *DBMgrBase) RedisHIncrBy(key, field string, incr int64) (int64, error) {
	return d.RedisInst.HIncrBy(d.RedisInst.Context(), key, field, incr).Result()
}

/*
*
多个服务器分布式，读取redis中的标志位
*/
func (d *DBMgrBase) RedisHIncrByFlagAdd(key string) (int64, error) {
	return d.RedisHIncrBy(key, WritingInRedis, 1)
}
func (d *DBMgrBase) RedisHIncrByFlagCheck(key string) (int64, error) {
	return d.RedisHIncrBy(key, WritingInRedis, 0)
}
func (d *DBMgrBase) RedisHIncrByFlagRelease(key string) (int64, error) {
	if !d.RedisExists(key) { //减的时候我们需要检查这个key还是否有效。
		return 0, ErrNoData
	}

	return d.RedisHIncrBy(key, WritingInRedis, -1)
}
func (d *DBMgrBase) RedisHIncrByGetVer(key string) (int64, error) {
	return d.RedisHIncrBy(key, VerInRedis, 0)
}

/*
*
防刷，防止暴力破解
*/
func (d *DBMgrBase) RedisCheckFireWall(key string, limit int64, ttl time.Duration) bool {
	times, err := d.RedisInst.Incr(d.RedisInst.Context(), key).Result()
	if err != nil {
		return false
	}
	d.RedisExpire(key, ttl)
	return times <= limit
}

/*
*
正常读取redis的数据，不管是否有其他服务器正在写入
*/
func (d *DBMgrBase) RedisHGetAll(key string, dataPtr interface{}) error {
	sc := d.RedisInst.HGetAll(d.RedisInst.Context(), key)
	//D("RedisHGetAll sc=%v", sc)
	if sc.Err() != nil {
		if sc.Err() == redis.Nil {
			return ErrNoData
		}
		E("GetHM err=%v", sc.Err())
		return sc.Err()
	}
	return DecodeRedis(sc.Val(), dataPtr)
}

/*
*
正常读取redis的数据，如果有其他服务器正在写入，那么直接返回错误
*/
func (d *DBMgrBase) RedisHGetAllEx(key string, dataPtr interface{}) error {
	//这里，即使redis中没有该key，也会返回一个空的map,所以要用exists判断一下
	if !d.RedisExists(key) { //在检查写入标志之前优先检查该值是否存在于redis中。
		return ErrNoData
	}

	wir, _ := d.RedisHIncrByFlagCheck(key)
	if wir != 0 {
		return ErrIsWriting
	}

	return d.RedisHGetAll(key, dataPtr)
}

/*
*
正常读取redis的数据，如果有其他服务器正在写入，那么等待写完毕，获取最新的数据
@tryCnt 重试次数，如果为-1表示永远等待
*/
func (d *DBMgrBase) RedisHGetAllExLoop(key string, dataPtr interface{}, tryCnt int, force bool) error {
	for {
		err := d.RedisHGetAllEx(key, dataPtr)
		if err == ErrIsWriting {
			if tryCnt >= 0 {
				if tryCnt == 0 {
					if force {
						d.RedisHGetAll(key, dataPtr)
						return nil
					}
					return ErrTryMax
				}
				tryCnt--
			}
			time.Sleep(time.Microsecond * 50)
			continue
		}
		return err
	}
}

/*
*
正常写入redis的数据，不管是否有其他服务器正在写入

@注意
谨慎使用该函数，该函数会导致之前的缓存数据被改写！！！
@expire 单位秒，传入负数表示永不过期或者不想改变原本的有效时间
*/
func (d *DBMgrBase) RedisHMSet(key string, data interface{}, expire time.Duration) error {
	var dataMap map[string]interface{}
	err := Decode(data, &dataMap, true)
	if err != nil {
		return err
	}
	sc := d.RedisInst.HMSet(d.RedisInst.Context(), key, dataMap)
	//D("RedisHMSet sc=%v", sc)

	if sc.Err() == nil && expire >= 0 {
		d.RedisExpire(key, expire)
	}

	return sc.Err()
}

/*
*
正常写入redis的数据，如果其他服务器正在写入，那么直接返回错误,如果缓存的信息比本地版本高也返回错误
*/
func (d *DBMgrBase) RedisHMSetEx(key string, data interface{}, ver int64, ignoreVer bool, expire time.Duration) error {
	wflag, _ := d.RedisHIncrByFlagAdd(key)
	defer d.RedisHIncrByFlagRelease(key)

	if wflag > 1 {
		return ErrIsWriting
	}

	if !ignoreVer {
		cur, _ := d.RedisHIncrByGetVer(key)
		if cur > ver {
			return ErrDataOld
		}
	}

	return d.RedisHMSet(key, data, expire)
}

type DealWithConflict func(oldPtr, newPtr interface{}) (interface{}, int64)

/*
*
正常写入redis的数据，如果其他服务器正在写入或者缓存中的信息版本比较高，那么我们等待写完毕把他取出来

@dataPtr 需要数据结构指针
@tryCnt 重试次数 -1表示永远尝试直到成功为止
*/
func (d *DBMgrBase) RedisHMSetExLoop(key string, dataPtr interface{}, ver int64, expire time.Duration,
	hook DealWithConflict, tryCnt int) error {

	if tryCnt >= 0 {
		if tryCnt == 0 {
			return ErrTryMax
		}
		tryCnt--
	}

	for {
		err := d.RedisHMSetEx(key, dataPtr, ver, false, expire)
		if err == ErrIsWriting || err == ErrDataOld {
			if hook == nil {
				return ErrAbort
			}

			newPtr := deepcopy.Copy(dataPtr)                  //深拷贝一个备份
			err = d.RedisHGetAllExLoop(key, newPtr, 3, false) //默认尝试3次获取最新的数据
			if err != nil {
				return err
			}

			result, newVer := hook(dataPtr, newPtr)
			if result == nil {
				return ErrAbort
			}

			return d.RedisHMSetExLoop(key, result, newVer, expire, hook, tryCnt)
		}
		return err
	}
}

func (d *DBMgrBase) RedisSAdd(key string, members ...interface{}) error {
	return d.RedisInst.SAdd(d.RedisInst.Context(), key, members...).Err()
}

func (d *DBMgrBase) RedisSRem(key string, members ...interface{}) error {
	return d.RedisInst.SRem(d.RedisInst.Context(), key, members...).Err()
}

func (d *DBMgrBase) RedisSMembers(key string) ([]string, error) {
	if !d.RedisExists(key) {
		return nil, ErrNoData
	}

	return d.RedisInst.SMembers(d.RedisInst.Context(), key).Result()
}

func (d *DBMgrBase) RedisSCard(key string) int64 {
	lenHere, _ := d.RedisInst.SCard(d.RedisInst.Context(), key).Result()
	return lenHere
}

/*
*
设置过期时间
*/
func (d *DBMgrBase) RedisExpire(key string, duration time.Duration) {
	d.RedisInst.Expire(d.RedisInst.Context(), key, duration)
}

func (d *DBMgrBase) RedisExpireAt(key string, tm time.Time) {
	d.RedisInst.ExpireAt(d.RedisInst.Context(), key, tm)
}

/*
*
@time.Duration -1表示永不过期，-2表示已经过期，大于0表示还有X秒过期。
*/
func (d *DBMgrBase) RedisTTL(key string) (time.Duration, error) {
	return d.RedisInst.TTL(d.RedisInst.Context(), key).Result()
}

func (d *DBMgrBase) RedisPersist(key string) {
	d.RedisInst.Persist(d.RedisInst.Context(), key)
}

// Deprecated: 当redis中key较多的时候，会导致redis阻塞 建议使用 RedisScan
func (d *DBMgrBase) RedisKeys(pattern string) ([]string, error) {
	return d.RedisInst.Keys(d.RedisInst.Context(), pattern).Result()
}

// 遍历key，每次遍历count个key
func (d *DBMgrBase) RedisScan(pattern string, count int64) (keys []string, err error) {
	return d.RedisScanLimit(pattern, count, 0)
}

func (d *DBMgrBase) RedisScanLimit(pattern string, count int64, limit int) (keys []string, err error) {
	cursor := uint64(0)
	for {
		var curKeys []string
		if curKeys, cursor, err = d.RedisScanWithCount(pattern, cursor, count); err != nil {
			break
		}

		keys = append(keys, curKeys...)

		if cursor == 0 { //已经遍历完成。
			break
		}

		//达到限制了，退出循环
		if limit > 0 && len(keys) >= limit {
			break
		}
	}
	return
}

func (d *DBMgrBase) RedisScanWithCount(pattern string, cursorStart uint64, count int64) (keys []string, cursor uint64, err error) {
	return d.RedisInst.Scan(d.RedisInst.Context(), cursorStart, pattern, count).Result()
}

func init() {
	SqlRegExp, _ = regexp.Compile(`\?`)
}
