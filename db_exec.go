package easydb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

// Exec 重写Exec方法以记录SQL查询
func (d *EasyDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	if d.loglevel > 1 {
		log.Printf("Exec SQL: (%s) args: (%v)", query, args)
	}
	result, err := d.db.Exec(query, args...)
	if d.loglevel > 0 {
		// 中文字符在SSH控制台可能输出UTF-8 编码的字节序列，每个字节表示成十六进制的 <XX> 形式
		log.Printf("SQL Exec Done. Cost: %v", time.Since(start))
	}
	return result, err
}

// ExecByFile 从文件执行SQL脚本
func (d *EasyDb) ExecByFile(filepath string, args ...interface{}) (sql.Result, error) {
	// 读取SQL文件内容
	sqlBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return d.Exec(string(sqlBytes), args...)
}

// ExecSqlWithTransaction 在事务中执行多条SQL语句
func (d *EasyDb) ExecSqlWithTransaction(sqlStatements []string) error {
	// 开始事务
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 确保函数结束时要么提交要么回滚事务
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // 重新抛出panic
		}
	}()

	// 执行每条SQL语句
	for _, sql := range sqlStatements {
		_, err := tx.Exec(sql)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("执行SQL失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
