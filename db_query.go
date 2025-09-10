package easydb

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
)

// GetOneData 根据where条件查询单条数据，支持结构体指针或map接收结果
// querySQL SQL查询语句 例：select field1, field2 from table1 where name = $1 and status = $2
// dest: 用于接收结果的结构体指针或map[string]any
// args: SQL参数
// 示例：
//
//	data := make(map[string]interface{}, 3)
//	d.GetOneData("SELECT id, name, age, wallet_balance FROM $1", &data, "users")
//	fmt.Printf("-----GetOneData--result(%+v)----\n", d.DecodeInterface(data))
func (d *EasyDb) GetOneData(querySQL string, dest interface{}, args ...interface{}) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("dest必须是有效的非空指针")
	}

	stmt, err := d.db.Prepare(querySQL)
	if err != nil {
		return fmt.Errorf("预处理SQL语句失败: %v", err)
	}
	defer stmt.Close()

	// 改用Query获取sql.Rows（即使只查一行）
	rows, err := stmt.Query(args...)
	if err != nil {
		return fmt.Errorf("查询失败: %v", err)
	}
	defer rows.Close()

	// 直接读取首行（模拟QueryRow行为）
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return fmt.Errorf("查询错误: %v", err)
		}
		return nil // 无数据
	}

	switch dd := dest.(type) {
	case *map[string]any:
		return d.scanRowToMap(rows, dd)
	default:
		if val.Elem().Kind() == reflect.Struct {
			return d.scanRowToStruct(rows, dest)
		}
		return fmt.Errorf("不支持的dest类型(%T)", dest)
	}
}

// DecodeInterface 用于解码GetOneData方法返回的结果
// data 参数类型实际是 map[string][]byte
// 示例：
//
//	data := make(map[string]interface{}, 3)
//	d.GetOneData("SELECT id, name, age, wallet_balance FROM $1", &data, "users")
//	fmt.Printf("-----GetOneData--result(%+v)----\n", d.DecodeInterface(data))
func (d EasyDb) DecodeInterface(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case []byte:
			strVal := string(v)
			// 尝试转为int64
			if i, err := strconv.ParseInt(strVal, 10, 64); err == nil {
				result[key] = i
				continue
			}
			// 尝试转为float64
			if f, err := strconv.ParseFloat(strVal, 64); err == nil {
				result[key] = f
				continue
			}
			// 尝试转为bool
			if b, err := strconv.ParseBool(strVal); err == nil {
				result[key] = b
				continue
			}
			// 其他情况保留为string
			result[key] = strVal
		default:
			result[key] = value
		}
	}
	return result
}

// GetOne 根据where条件查询单条数据
// querySQL SQL查询语句
// args: SQL参数
// dest: 用于接收结果的结构体指针
// 示例：
//
//	var qrid *int
//	var qrToUrl *string
//	GetOne("select id, to_url from qr_list where code = $1", []interface{}{qrid, qrToUrl}, "codexxx")
func (d *EasyDb) GetOne(querySQL string, dest []interface{}, args ...interface{}) error {
	// 使用预处理语句执行查询，防止SQL注入
	stmt, err := d.db.Prepare(querySQL)
	if err != nil {
		return fmt.Errorf("预处理SQL语句失败: %v", err)
	}
	defer stmt.Close()

	// 执行预处理查询
	row := stmt.QueryRow(args...)
	if err := row.Scan(dest...); err != nil {
		if err == sql.ErrNoRows {
			// return fmt.Errorf("未找到匹配的数据记录")
			return nil
		}
		return fmt.Errorf("查询数据失败: %v", err)
	}
	return nil
}

// GetMany 根据where条件查询多条数据
// querySQL SQL查询语句
// dest: 用于接收结果的切片指针
// args: SQL参数
// 示例：
//
//	var datalist []map[string]interface{}
//	d.GetMany("SELECT id, name, age, wallet_balance FROM users", &datalist)
func (d *EasyDb) GetMany(querySQL string, dest interface{}, args ...interface{}) error {
	// 使用预处理语句执行查询，防止SQL注入
	stmt, err := d.db.Prepare(querySQL)
	if err != nil {
		return fmt.Errorf("预处理SQL语句失败: %v", err)
	}
	defer stmt.Close()

	// 执行预处理查询
	rows, err := stmt.Query(args...)
	if err != nil {
		return fmt.Errorf("查询数据失败: %v", err)
	}
	defer rows.Close()

	// 使用sql.Rows.Scan将结果扫描到目标切片
	return d.scanRows(rows, dest)
}
