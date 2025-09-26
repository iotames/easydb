package easydb

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
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
	kidlist := []reflect.Kind{
		reflect.Map,
		reflect.Struct,
		reflect.String,
		reflect.Int,
		reflect.Float64,
		reflect.Interface,
	}
	for rows.Next() {
		var elem reflect.Value
		if !kindsContains(elemType.Kind(), kidlist) {
			return fmt.Errorf("不支持的切片元素类型: %v", elemType.Kind())
		}

		if elemType.Kind() == reflect.String {
			// 处理字符串
			cols, err := rows.Columns()
			if err != nil {
				return err
			}
			if len(cols) < 1 {
				return fmt.Errorf("没有可用的列")
			}
			var val any
			if err := rows.Scan(&val); err != nil {
				return err
			}
			var str string
			switch v := val.(type) {
			case []byte:
				str = string(v)
			case string:
				str = v
			default:
				str = fmt.Sprintf("%v", v)
			}
			elem = reflect.ValueOf(str)
		}

		if elemType.Kind() == reflect.Int {
			// 处理int
			var val any
			if err := rows.Scan(&val); err != nil {
				return err
			}
			var intval int64
			switch v := val.(type) {
			case int64:
				intval = v
			case []byte:
				intval, _ = strconv.ParseInt(string(v), 10, 64)
			case string:
				intval, _ = strconv.ParseInt(v, 10, 64)
			default:
				return fmt.Errorf("无法转换为int: %v", v)
			}
			elem = reflect.ValueOf(int(intval))
		}

		if elemType.Kind() == reflect.Float64 {
			// 处理float64
			var val any
			if err := rows.Scan(&val); err != nil {
				return err
			}
			var floatval float64
			switch v := val.(type) {
			case float64:
				floatval = v
			case []byte:
				floatval, _ = strconv.ParseFloat(string(v), 64)
			case string:
				floatval, _ = strconv.ParseFloat(v, 64)
			default:
				return fmt.Errorf("无法转换为float64: %v", v)
			}
			elem = reflect.ValueOf(floatval)
		}

		if elemType.Kind() == reflect.Interface {
			// 处理interface{}
			var val interface{}
			if err := rows.Scan(&val); err != nil {
				return err
			}
			elem = reflect.ValueOf(decodeAny(val))
		}

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
		}

		if elemType.Kind() == reflect.Struct {
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
		}

		sliceVal.Set(reflect.Append(sliceVal, elem))
	}
	return rows.Err()
}

// 判断 kind 是否在 kinds 切片中
func kindsContains(kind reflect.Kind, kinds []reflect.Kind) bool {
	// for _, k := range kinds {
	//     if k == kind {
	//         return true
	//     }
	// }
	// Go 1.21+
	return slices.Contains(kinds, kind)
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
