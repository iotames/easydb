package easydb

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	once     sync.Once
	instance *EasyDb
)

// GetEasyDb 获取EasyDb数据库连接单例
func GetEasyDb() *EasyDb {
	once.Do(func() {
		if instance == nil {
			panic("请先调用SetEasyDb方法初始化数据库连接")
		}
	})
	return instance
}

// SetEasyDb 设置EasyDb数据库连接单例
// 使用NewEasyDbBySqlDB 或 NewEasyDb方法，初始化EasyDb数据库连接实例。然后通过此函数设置单例
func SetEasyDb(edb *EasyDb) {
	if edb == nil {
		panic("数据库连接单例edb *EasyDb不能设置为nill")
	}
	instance = edb
}

type EasyDb struct {
	db       *sql.DB
	loglevel int
}

// SowLog 展示运行日志。默认0为不展示。数值越大越详细。
func (d *EasyDb) SowLog(level int) {
	d.loglevel = level
}

// NewEasyDb 初始化EasyDb数据库连接实例。
// driverName See https://golang.org/s/sqldrivers for a list of third-party drivers.
//
// 示例：
//
//	import (
//	"github.com/iotames/easydb"
//	_ "github.com/go-sql-driver/mysql"
//	_ "github.com/lib/pq"
//	)
//	d := easydb.NewEasyDb("mysql", "127.0.0.1", "root", "password", "testdb", 3306)
//	d1 := easydb.NewEasyDb("postgres", "127.0.0.1", "username", "password", "testdb", 5432)
func NewEasyDb(driverName, dbHost, dbUser, dbPassword, dbName string, dbPort int) *EasyDb {
	cf := NewDsnConf(driverName, dbHost, dbUser, dbPassword, dbName, dbPort)
	return NewEasyDbByConf(*cf)
}

// NewEasyDbByConf 使用DsnConf参数初始化EasyDb实例。
// 示例：
//
//	import (
//	"github.com/iotames/easydb"
//	_ "github.com/go-sql-driver/mysql"
//	_ "github.com/lib/pq"
//	)
//	cf := easydb.NewDsnConf("mysql", "127.0.0.1", "root", "password", "testdb", 3306)
//	// 可选：cf.UpdateDsnTpl("mysql", "DB_USER:DB_PASSWORD@tcp(DB_HOST:DB_PORT)/DB_NAME?charset=utf8mb4&parseTime=True&loc=Local")
//	dbmysql := easydb.NewEasyDbByConf(*cf)
//	cfpg := easydb.NewDsnConf("postgres", "127.0.0.1", "username", "password", "testdb", 5432)
//	// 可选：cfpg.UpdateDsnTpl("postgres", "user=DB_USER password=DB_PASSWORD dbname=DB_NAME host=DB_HOST port=DB_PORT sslmode=disable search_path=public")
//	dbpg := easydb.NewEasyDbByConf(*cfpg)
func NewEasyDbByConf(cf DsnConf) *EasyDb {
	var err error
	var dsn string
	if !cf.CheckAvailable() {
		var supportDrivers []string
		dmp := cf.GetAvailableDsnTplMap()
		for k := range dmp {
			supportDrivers = append(supportDrivers, k)
		}
		panic(fmt.Sprintf("driverName不支持%s，仅支持[%s]。请自选合适的数据库驱动，调用NewEasyDbBySqlDB或NewEasyDbByConf方法初始化。可用驱动：https://golang.org/s/sqldrivers", cf.DriverName, strings.Join(supportDrivers, ",")))
	}
	dsn, err = cf.GetDsn()
	if err != nil {
		panic(err)
	}
	var sqldb *sql.DB
	sqldb, err = sql.Open(cf.DriverName, dsn)
	if err != nil {
		panic(err)
	}
	return NewEasyDbBySqlDB(sqldb)
}

// NewEasyDbBySqlDB 使用sqldb *sql.DB参数初始化EasyDb实例。
//
// 各数据库驱动：https://golang.org/s/sqldrivers
//
//	import (
//	"github.com/iotames/easydb"
//	_ "github.com/go-sql-driver/mysql"
//	_ "github.com/lib/pq"
//	_ "github.com/mattn/go-sqlite3"
//	)
//
//	sqldb, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/debugdb")
//	//  sqldb, err = sql.Open("postgres", "user=postgres password=postgres dbname=postgres host=127.0.0.1 port=5432 sslmode=disable search_path=public")
//	//  sqldb, err := sql.Open("sqlite3", "./mydb.sqlite")
//	d := NewEasyDbBySqlDB(sqldb)
func NewEasyDbBySqlDB(sqldb *sql.DB) *EasyDb {
	return &EasyDb{db: sqldb}
}

// Query 重写Query方法以记录SQL查询
func (d *EasyDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	if d.loglevel > 1 {
		log.Printf("Query SQL: (%s) args: (%v)", query, args)
	}
	rows, err := d.db.Query(query, args...)
	if d.loglevel > 0 {
		log.Printf("Query Done. Cost: %v", time.Since(start))
	}
	return rows, err
}

// QueryRow 重写QueryRow方法以记录SQL查询
func (d *EasyDb) QueryRow(query string, args ...interface{}) *sql.Row {
	// log.Printf("查询单行SQL: %s 参数: %v", query, args)
	return d.db.QueryRow(query, args...)
}

func (d *EasyDb) Ping() error {
	return d.db.Ping()
}

// CloseDb 关闭整个数据库连接池
func (d *EasyDb) CloseDb() error {
	return d.db.Close()
}
