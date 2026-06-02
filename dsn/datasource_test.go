package dsn

import (
	"testing"
)

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
			err := checkMySQLDSNPassword(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkMySQLDSNPassword(%q) error = %v, wantErr = %v", tt.dsn, err, tt.wantErr)
			}
		})
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
