package easydb

import (
	"fmt"
	"strings"
)

// ExecInsert 执行单条插入语句
func (d *EasyDb) ExecInsert(tableName string, columns []string, values []interface{}) error {
	// 构建插入SQL语句
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	sqlText := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// 执行插入操作
	_, err := d.db.Exec(sqlText, values...)
	if err != nil {
		return fmt.Errorf("插入数据失败: %v", err)
	}

	return nil
}

// ExecUpdate 执行更新语句
// tableName: 表名
// columns: 要更新的列名
// values: 要更新的值
// whereClause: WHERE子句。例如："id = $1"
// whereValues: WHERE子句中的参数值。例如：[]interface{}{1}
func (d *EasyDb) ExecUpdateByValues(tableName string, columns []string, values []interface{}, whereClause string, whereValues []interface{}) error {
	// 构建SET子句
	setClause := make([]string, len(columns))
	paramCount := 1
	for i, col := range columns {
		setClause[i] = fmt.Sprintf("%s = $%d", col, paramCount)
		paramCount++
	}

	// 构建完整的更新SQL语句
	sqlText := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		tableName,
		strings.Join(setClause, ", "),
		whereClause,
	)

	// 合并values和whereValues
	allValues := append(values, whereValues...)

	// 执行更新操作
	result, err := d.db.Exec(sqlText, allValues...)
	if err != nil {
		return fmt.Errorf("更新数据失败: %v", err)
	}

	// 检查受影响的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取受影响行数失败: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("没有记录被更新")
	}

	return nil
}
