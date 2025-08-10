package easydb

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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
	db *sql.DB
}

// NewEasyDb 初始化EasyDb数据库连接实例
// driverName See https://golang.org/s/sqldrivers for a list of third-party drivers.
func NewEasyDb(driverName, dbHost, dbUser, dbPassword, dbName string, dbPort int) *EasyDb {
	var driverMap = map[string]string{
		// "user=postgres password=postgres dbname=postgres host=127.0.0.1 port=5432 sslmode=disable search_path=public" or "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full" See: https://pkg.go.dev/github.com/lib/pq
		"postgres": fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", dbUser, dbPassword, dbName, dbHost, dbPort),
		// username:password@protocol(address)/dbname?param=value See: https://github.com/go-sql-driver/mysql/
		"mysql": fmt.Sprintf(`%s:%s@tcp(%s:%d)/%s`, dbUser, dbPassword, dbHost, dbPort, dbName),
	}

	driv, ok := driverMap[driverName]
	if !ok {
		var supportDrivs []string
		for k := range driverMap {
			supportDrivs = append(supportDrivs, k)
		}
		panic(fmt.Sprintf("driverName 不支持%s 仅支持: %s 要使用其他数据库，请自行选择驱动，然后调用NewEasyDbBySqlDB方法。可用驱动请移步链接：https://golang.org/s/sqldrivers", driverName, strings.Join(supportDrivs, ",")))
	}
	var err error
	var sqldb *sql.DB
	sqldb, err = sql.Open(driverName, driv)
	if err != nil {
		panic(err)
	}
	return NewEasyDbBySqlDB(sqldb)
}

// NewEasyDbBySqlDB 使用sqldb *sql.DB参数初始化EasyDb实例。
// 各数据库驱动如下所示：
//
// import (
//
//	_ "github.com/go-sql-driver/mysql"
//	_ "github.com/lib/pq"
//	_ "github.com/mattn/go-sqlite3"
//
// )
//
//			  sqldb, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/debugdb")
//	     //  sqldb, err = sql.Open("postgres", "user=postgres password=postgres dbname=postgres host=127.0.0.1 port=5432 sslmode=disable search_path=public")
//	     //  sqldb, err := sql.Open("sqlite3", "./mydb.sqlite")
//		      d := NewEasyDbBySqlDB(sqldb)
func NewEasyDbBySqlDB(sqldb *sql.DB) *EasyDb {
	return &EasyDb{db: sqldb}
}

// Query 重写Query方法以记录SQL查询
func (d *EasyDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	// log.Printf("查询SQL: %s 参数: %v", query, args)
	rows, err := d.db.Query(query, args...)
	log.Printf("SQL查询完成，耗时: %v", time.Since(start))
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

func (d *EasyDb) CloseDb() error {
	return d.db.Close()
}
