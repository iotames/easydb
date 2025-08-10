package easydb

import (
	"fmt"
)

// 定义占位符映射
var placeholder = map[string]string{
	"postgres": "$%d",
	"mysql":    "?",
	"sqlite":   "?",
}

// getPlaceholder 生成参数占位符
// mysql, sqlite 的参数占位符是?. postgres则是 $1, $2, $3 ...
func GetPlaceholder(dbType string, index int) string {
	return fmt.Sprintf(placeholder[dbType], index+1)
}
