package easydb

import "testing"

// TestGetPlaceholder 占位符生成回归：postgres 按序号 $n；mysql/sqlite 恒为 "?"。
// 此前 ? 方言误走 fmt.Sprintf("?", index+1)，产出 "?%!(EXTRA int=n)" 垃圾尾巴
// （拼进 IN (...) 即 SQL 语法错误，GeoIP 同步在 Postgres 上实测踩坑）。
func TestGetPlaceholder(t *testing.T) {
	cases := []struct {
		dbType string
		index  int
		want   string
	}{
		{"postgres", 0, "$1"},
		{"postgres", 9, "$10"},
		{"mysql", 0, "?"},
		{"mysql", 4, "?"},
		{"sqlite", 0, "?"},
		{"sqlite", 4, "?"},
		{"未知方言回落", 0, "?"}, // 非法 dbType 不产生垃圾尾巴
	}
	for _, c := range cases {
		if got := GetPlaceholder(c.dbType, c.index); got != c.want {
			t.Errorf("GetPlaceholder(%q, %d) = %q, want %q", c.dbType, c.index, got, c.want)
		}
	}
}
