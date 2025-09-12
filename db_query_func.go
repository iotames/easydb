package easydb

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
)

// scanRows 扫描*sql.Rows数据到切片指针
// dest 切片的指针
func (d *EasyDb) scanRows(rows *sql.Rows, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("目标参数必须是切片指针")
	}
	sliceVal := v.Elem()
	elemType := sliceVal.Type().Elem()
	for rows.Next() {
		var elem reflect.Value
		if elemType.Kind() == reflect.Map {
			cols, err := rows.Columns()
			if err != nil {
				return err
			}
			values := make([]interface{}, len(cols))
			for i := range values {
				values[i] = new(interface{})
			}
			if err := rows.Scan(values...); err != nil {
				return err
			}
			m := reflect.MakeMap(elemType)
			for i, col := range cols {
				vv := decodeAny(*(values[i].(*interface{})))
				m.SetMapIndex(reflect.ValueOf(col), reflect.ValueOf(vv))
			}
			elem = m
		} else if elemType.Kind() == reflect.Struct {
			// 处理结构体
			newElem := reflect.New(elemType).Interface()
			cols, err := rows.Columns()
			if err != nil {
				return err
			}
			fields := make([]interface{}, len(cols))
			destVal := reflect.ValueOf(newElem).Elem()
			for i, col := range cols {
				fieldFound := false
				for j := 0; j < destVal.NumField(); j++ {
					tag := destVal.Type().Field(j).Tag.Get("db")
					if tag == col {
						fields[i] = destVal.Field(j).Addr().Interface()
						fieldFound = true
						break
					}
				}
				if !fieldFound {
					return fmt.Errorf("列 %s 无对应的结构体字段", col)
				}
			}
			if err := rows.Scan(fields...); err != nil {
				return err
			}
			elem = destVal
		} else {
			return fmt.Errorf("不支持的切片元素类型: %v", elemType.Kind())
		}
		sliceVal.Set(reflect.Append(sliceVal, elem))
	}
	return rows.Err()
}

// scanRowToMap 扫描*sql.Rows数据到*map[string]any
// func (d *EasyDb) scanRowToMap(rows *sql.Rows, dest *map[string]any) error {
func (d *EasyDb) scanRowToMap(rows *sql.Rows, dest map[string]any) error {
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("获取列失败: %v", err)
	}

	values := make([]interface{}, len(cols))
	for i := range values {
		values[i] = new(interface{})
	}

	if err := rows.Scan(values...); err != nil {
		return fmt.Errorf("扫描失败: %v", err)
	}

	// result := make(map[string]any)
	for i, col := range cols {
		dtval := *(values[i].(*interface{}))
		// result[col] = decodeAny(dtval)
		dest[col] = decodeAny(dtval)
	}
	// *dest = result
	return nil
}

// scanRowToStruct 扫描*sql.Rows数据到结构体（通用逻辑）
func (d *EasyDb) scanRowToStruct(rows *sql.Rows, dest interface{}) error {
	// 使用rows.Columns()验证列与结构体标签匹配
	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("获取列失败: %v", err)
	}

	destVal := reflect.ValueOf(dest).Elem()
	fields := make([]interface{}, len(cols))

	// 按列名映射结构体字段
	for i, col := range cols {
		fieldFound := false
		for j := 0; j < destVal.NumField(); j++ {
			tag := destVal.Type().Field(j).Tag.Get("db")
			if tag == col {
				fields[i] = destVal.Field(j).Addr().Interface()
				fieldFound = true
				break
			}
		}
		if !fieldFound {
			return fmt.Errorf("列 %s 无对应的结构体字段", col)
		}
	}

	return rows.Scan(fields...)
}

func decodeMapAny(data map[string]any) map[string]any {
	result := make(map[string]any, len(data))
	for key, value := range data {
		result[key] = decodeAny(value)
	}
	return result

}

func decodeAny(value any) any {
	switch v := value.(type) {
	case []byte:
		strVal := string(v)
		// 尝试转为int64
		if i, err := strconv.ParseInt(strVal, 10, 64); err == nil {
			return i
		}
		// 尝试转为float64
		if f, err := strconv.ParseFloat(strVal, 64); err == nil {
			return f
		}
		// 尝试转为bool
		if b, err := strconv.ParseBool(strVal); err == nil {
			return b
		}
		return strVal

	default:
		// result[key] = value
		return value
	}
}
