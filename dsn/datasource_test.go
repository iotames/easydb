package dsn

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

// stubDriver 仅满足 sql.Register 的 driver.Driver 接口，用于测试驱动注册校验，
// 不提供真实连接（Open 永远失败，测试不触达连接）。
type stubDriver struct{}

func (stubDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("stub driver: 仅供注册测试，不可连接")
}

func init() { sql.Register("dsntest-stub", stubDriver{}) }

func TestCheckMySQLDSNPassword(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{"normal password", "root:pass@tcp(127.0.0.1:3306)/testdb", false},
		{"no password", "root@tcp(127.0.0.1:3306)/testdb", false},
		{"empty password", "root:@tcp(127.0.0.1:3306)/testdb", false},
		{"bare @ in password", "root:pass@word@tcp(127.0.0.1:3306)/testdb", true},
		{"already encoded @", "root:pass%40word@tcp(127.0.0.1:3306)/testdb", false},
		{"multiple bare @", "root:p@ss@word@tcp(127.0.0.1:3306)/testdb", true},
		{"tcp4 network", "user:p@ss@tcp4(host:3306)/db", true},
		{"unix socket", "user:pass@unix(/tmp/mysql.sock)/db", false},
		{"no @network( pattern", "root:pass@unknown@host/db", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMySQLDSNPassword(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckMySQLDSNPassword(%q) error = %v, wantErr = %v", tt.dsn, err, tt.wantErr)
			}
		})
	}
}

func TestAppendDsnDedup(t *testing.T) {
	g := &DsnGroup{}
	dsn1 := "stub://point?x=1"
	if err := g.AppendNamedDsn("a", "dsntest-stub", dsn1); err != nil {
		t.Fatalf("首次添加应成功: %v", err)
	}
	// 同 DSN 不同 Name：同 DSN 即同 Code，必须拦截（否则同 Code 双条目，按 code 删除语义未定义）
	if err := g.AppendNamedDsn("b", "dsntest-stub", dsn1); err == nil {
		t.Fatal("重复 DSN 未被拦截")
	}
	// 不同 DSN 正常追加
	if err := g.AppendNamedDsn("c", "dsntest-stub", "stub://point?x=2"); err != nil {
		t.Fatalf("不同 DSN 添加应成功: %v", err)
	}
	if len(g.DsnList) != 2 {
		t.Fatalf("添加去重后应恰好 2 条，实际 %d 条", len(g.DsnList))
	}
	// 未注册驱动照旧拦截（原有行为不回归）
	if err := g.AppendNamedDsn("d", "no-such-driver", "x://y"); err == nil {
		t.Fatal("未注册驱动未被拦截")
	}
}

func TestFindLastDSNAt(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want int
	}{
		{"simple tcp", "user:pass@tcp(host:3306)/db", 9},
		{"password with @", "user:pass@word@tcp(host:3306)/db", 14},
		{"tcp4", "user:pass@tcp4(host:3306)/db", 9},
		{"tcp6", "user:pass@tcp6(host:3306)/db", 9},
		{"unix", "user:pass@unix(/tmp/mysql.sock)/db", 9},
		{"no match", "user:pass:host:port/db", -1},
		{"empty", "", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findLastDSNAt(tt.dsn)
			if got != tt.want {
				t.Errorf("findLastDSNAt(%q) = %v, want %v", tt.dsn, got, tt.want)
			}
		})
	}
}
