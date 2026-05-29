package dsn

import (
	"database/sql"
	"fmt"

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
	code := miniutils.Md5(dsn)
	ds := DataSource{Code: code, Name: name, DriverName: driverName, Dsn: dsn}
	if len(d.DsnList) == 0 {
		d.ActiveCode = code
	}
	d.DsnList = append(d.DsnList, ds)
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
