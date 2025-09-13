package easydb

import (
	"database/sql"
	"fmt"
	"strings"
)

const DSN_TPL_MYSQL = "DB_USER:DB_PASSWORD@tcp(DB_HOST:DB_PORT)/DB_NAME"
const DSN_TPL_POSTGRES = "user=DB_USER password=DB_PASSWORD dbname=DB_NAME host=DB_HOST port=DB_PORT sslmode=disable"

type DsnConf struct {
	DriverName                         string
	DbHost, DbUser, DbPassword, DbName string
	DbPort                             int
	dsnTplMap                          map[string]string
}

// NewDsnConf 创建DsnConf实例，包含常用数据库的dsn模板。
// If you need to add more database drivers, please use AddDsnTpl method to add dsn template after creating DsnConf instance.
// 示例：
//
//	import (
//	"github.com/iotames/easydb"
//	_ "github.com/go-sql-driver/mysql"
//	_ "github.com/lib/pq"
//	)
//	cf1 := easydb.NewDsnConf("mysql", "127.0.0.1", "root", "password", "testdb", 3306)
//	// 可选：cf1.UpdateDsnTpl("mysql", "DB_USER:DB_PASSWORD@tcp(DB_HOST:DB_PORT)/DB_NAME?charset=utf8mb4&parseTime=True&loc=Local")
//	dbmysql := easydb.NewEasyDbByConf(*cf1)
//	cf2 := easydb.NewDsnConf("postgres", "127.0.0.1", "username", "password", "testdb", 5432)
//	// 可选：cf2.UpdateDsnTpl("postgres", "user=DB_USER password=DB_PASSWORD dbname=DB_NAME host=DB_HOST port=DB_PORT sslmode=disable search_path=public")
//	dbpg := easydb.NewEasyDbByConf(*cf2)
func NewDsnConf(driverName, dbHost, dbUser, dbPassword, dbName string, dbPort int) *DsnConf {
	mp := map[string]string{
		"mysql":    DSN_TPL_MYSQL,
		"postgres": DSN_TPL_POSTGRES,
	}
	return &DsnConf{DriverName: strings.ToLower(driverName), DbHost: dbHost, DbUser: dbUser, DbPassword: dbPassword, DbName: dbName, DbPort: dbPort, dsnTplMap: mp}
}

// GetDsn 生成dsn字符串
func (cf DsnConf) GetDsn() (string, error) {
	tpl, err := cf.GetDsnTpl()
	if err != nil {
		return "", err
	}
	dsn := strings.ReplaceAll(tpl, "DB_USER", cf.DbUser)
	dsn = strings.ReplaceAll(dsn, "DB_PASSWORD", cf.DbPassword)
	dsn = strings.ReplaceAll(dsn, "DB_NAME", cf.DbName)
	dsn = strings.ReplaceAll(dsn, "DB_HOST", cf.DbHost)
	dsn = strings.ReplaceAll(dsn, "DB_PORT", fmt.Sprintf("%d", cf.DbPort))
	return dsn, nil
}

// GetDsn 获取dsn字符串模板
func (cf DsnConf) GetDsnTpl() (string, error) {
	tpl, ok := cf.dsnTplMap[cf.DriverName]
	if !ok {
		return "", fmt.Errorf("DriverName %s not found in dsnTplMap", cf.DriverName)
	}
	return tpl, nil
}

// CheckAvailable 检查当前系统是否支持该数据库驱动
func (cf DsnConf) CheckAvailable() bool {
	mp := cf.GetAvailableDsnTplMap()
	if len(mp) == 0 {
		return false
	}
	_, ok := mp[cf.DriverName]
	return ok
}

// GetAvailableDsnTplMap 获取当前系统支持的数据库驱动及其对应的dsn模板
func (cf DsnConf) GetAvailableDsnTplMap() map[string]string {
	drivers := sql.Drivers()
	mp := make(map[string]string)
	for _, d := range drivers {
		if tpl, ok := cf.dsnTplMap[d]; ok {
			mp[d] = tpl
		}
	}
	return mp
}

// AddDsnTpl 添加新的dsn模板
// If the driverName already exists, it returns an error.
// 示例：
//
//	// test.db是dbName参数，填sqlite3的文件名，可以换成绝对路径
//	cf := easydb.NewDsnConf("sqlite3", "", "", "", "test.db", 0)
//	err := cf.AddDsnTpl("sqlite3", "DB_NAME")
func (cf *DsnConf) AddDsnTpl(driverName, dsnTpl string) error {
	if cf.dsnTplMap == nil {
		cf.dsnTplMap = make(map[string]string)
	}
	_, ok := cf.dsnTplMap[driverName]
	if ok {
		return fmt.Errorf("driverName %s already exist in dsnTplMap, please use UpdateDsnTpl method to update it", driverName)
	}
	cf.dsnTplMap[driverName] = dsnTpl
	return nil
}

// UpdateDsnTpl 更新已存在的dsn模板
// If the driverName does not exist, it returns an error.
// 示例：
//
//	cf := easydb.NewDsnConf("postgres", "127.0.0.1", "username", "password", "testdb", 5432)
//	cf.UpdateDsnTpl("postgres", "user=DB_USER password=DB_PASSWORD dbname=DB_NAME host=DB_HOST port=DB_PORT sslmode=disable search_path=public")
func (cf *DsnConf) UpdateDsnTpl(driverName, dsnTpl string) error {
	if cf.dsnTplMap == nil {
		return fmt.Errorf("dsnTplMap is nil, please use AddDsnTpl method to add dsn template first")
	}
	_, ok := cf.dsnTplMap[driverName]
	if !ok {
		return fmt.Errorf("driverName %s not exist in dsnTplMap, please use AddDsnTpl method to add dsn template first", driverName)
	}
	cf.dsnTplMap[driverName] = dsnTpl
	return nil
}
