package easydb

import "fmt"

// 定义占位符映射
var placeholder = map[string]string{
	"postgres": "$%d",
	"mysql":    "?",
	"sqlite":   "?",
}

// getPlaceholder 生成参数占位符
// mysql, sqlite 的参数占位符是?（无序号语义，index 不参与格式化）；postgres则是 $1, $2, $3 ...
// 未知方言回落 "?"（与 mysql/sqlite 同形态）。
// 修复（实测踩坑）：此前对 ? 方言也走 fmt.Sprintf("?", index+1)，格式动词少于实参时
// 产出 "?%!(EXTRA int=n)" 垃圾尾巴，拼进 IN (...) 即 SQL 语法错误。
func GetPlaceholder(dbType string, index int) string {
	if dbType == "postgres" {
		return fmt.Sprintf(placeholder[dbType], index+1)
	}
	return "?"
}
