package dsn

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/iotames/miniutils"
)

type DataSource struct {
	Code       string
	DriverName string
	Dsn        string
	Name       string // 人工可读的连接名，用于 ETL 任务按名引用
}

type DsnGroup struct {
	ActiveCode string
	DsnList    []DataSource
}

func (d DsnGroup) GetDefaultDSN() DataSource {
	if len(d.DsnList) == 0 {
		return DataSource{}
	}
	dsn := d.GetActiveDSN()
	if d.ActiveCode == "" || dsn.Code == "" {
		return d.DsnList[0]
	}
	return dsn
}

func (d DsnGroup) GetActiveDSN() DataSource {
	mp := d.getDsnMap()
	var ok bool
	var dsn DataSource
	if dsn, ok = mp[d.ActiveCode]; !ok {
		return DataSource{}
	}
	return dsn
}

func (d DsnGroup) getDsnMap() map[string]DataSource {
	mp := make(map[string]DataSource, len(d.DsnList))
	for _, dd := range d.DsnList {
		mp[dd.Code] = dd
	}
	return mp
}

func (d DsnGroup) HasDsn(dsn string) bool {
	hasd := false
	for _, dd := range d.DsnList {
		if dd.Dsn == dsn {
			hasd = true
			break
		}
	}
	return hasd
}

func (d DsnGroup) HasActive(dsnCode string) bool {
	mp := d.getDsnMap()
	_, ok := mp[dsnCode]
	if !ok {
		return false
	}
	return d.ActiveCode == dsnCode
}

func (d *DsnGroup) Active(dsnCode string) error {
	found := false
	for _, dd := range d.DsnList {
		if dd.Code == dsnCode {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("not found dsnCode: %s", dsnCode)
	}
	d.ActiveCode = dsnCode
	return nil
}

// AppendDsn 追加数据源（向下兼容，Name 为空）
func (d *DsnGroup) AppendDsn(driverName, dsn string) error {
	return d.appendDsn("", driverName, dsn)
}

// AppendNamedDsn 追加带连接名的数据源
func (d *DsnGroup) AppendNamedDsn(name, driverName, dsn string) error {
	return d.appendDsn(name, driverName, dsn)
}

func (d *DsnGroup) appendDsn(name, driverName, dsn string) error {
	drivers := sql.Drivers()
	if miniutils.GetIndexOf(driverName, drivers) == -1 {
		return fmt.Errorf("数据库驱动%s未注册。已注册的数据库驱动有：%v", driverName, drivers)
	}
	if driverName == "mysql" {
		if err := checkMySQLDSNPassword(dsn); err != nil {
			return err
		}
	}
	code := miniutils.Md5(dsn)
	ds := DataSource{Code: code, Name: name, DriverName: driverName, Dsn: dsn}
	if len(d.DsnList) == 0 {
		d.ActiveCode = code
	}
	d.DsnList = append(d.DsnList, ds)
	return nil
}

// mysqlNetworks MySQL DSN 支持的 network 类型列表
// DSN 格式: user:password@network(addr)/dbname
var mysqlNetworks = []string{"tcp", "tcp4", "tcp6", "unix"}

// findLastDSNAt 从右往左找最后一个 @network( 分隔符的位置。
// 密码可能包含 @，所以不能简单用 strings.LastIndex(dsn, "@")。
// 必须找到 @ 后面紧跟着 network 类型 + ( 的才是真正的分隔符。
// 返回 @ 的索引，未找到返回 -1。
func findLastDSNAt(dsn string) int {
	// 收集所有 network 类型匹配的 @ 位置，取最靠右的
	rightmost := -1
	for _, netName := range mysqlNetworks {
		tag := "@" + netName + "("
		if idx := strings.LastIndex(dsn, tag); idx > rightmost {
			rightmost = idx
		}
	}
	if rightmost >= 0 {
		return rightmost
	}
	// 兜底：从右往左逐字符扫描，找 @ + 字母序列 + (
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] != '@' {
			continue
		}
		for _, netName := range mysqlNetworks {
			end := i + 1 + len(netName)
			if end < len(dsn) && dsn[i+1:end] == netName && dsn[end] == '(' {
				return i
			}
		}
	}
	return -1
}

// checkMySQLDSNPassword 检测 MySQL DSN 密码中是否含裸 @。
// 密码中的 @ 会与 @network( 分隔符混淆，导致 DSN 解析失败。
// 由于 DSN 是完整字符串传入，不确定密码是否已编码，故不自动编码，而是报错提示。
func checkMySQLDSNPassword(dsn string) error {
	atIdx := findLastDSNAt(dsn)
	if atIdx < 0 {
		// 无法确定 DSN 结构，跳过校验
		return nil
	}
	userPass := dsn[:atIdx]
	colonIdx := strings.Index(userPass, ":")
	if colonIdx < 0 {
		// 无密码部分，跳过校验
		return nil
	}
	password := userPass[colonIdx+1:]
	if strings.Contains(password, "@") {
		return fmt.Errorf("MySQL DSN 密码中包含 @ 符号，请先使用 url.QueryEscape() 对密码中的特殊字符编码后再传入")
	}
	return nil
}

// GetDSNByName 按连接名查找数据源
func (d DsnGroup) GetDSNByName(name string) (DataSource, bool) {
	for _, dd := range d.DsnList {
		if dd.Name == name {
			return dd, true
		}
	}
	return DataSource{}, false
}

// HasName 检查连接名是否存在
func (d DsnGroup) HasName(name string) bool {
	_, ok := d.GetDSNByName(name)
	return ok
}
