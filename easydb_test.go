package easydb

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestPostgresAdd(t *testing.T) {
	var sqldb *sql.DB
	var err error
	sqlCreateTable := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50),
        age INT,
		wallet_balance DECIMAL(10,2) DEFAULT 0.00
    )`

	sqldb, err = sql.Open("postgres", "user=postgres password=postgres dbname=postgres host=127.0.0.1 port=5432 sslmode=disable search_path=public")
	if err != nil {
		t.Error(err)
	}
	d := NewEasyDbBySqlDB(sqldb)
	// d := NewEasyDb("postgres", "127.0.0.1", "postgres", "postgres", "postgres", 5432)
	_, err = d.Exec(sqlCreateTable)
	if err != nil {
		t.Error(err)
	}
	// 插入数据
	for i := 1; i <= 5; i++ {
		_, err = d.Exec("INSERT INTO users (name, age, wallet_balance) VALUES ($1, $2, $3)", fmt.Sprintf("Hankin%d", i), i, float64(i*10))
		if err != nil {
			t.Error(err)
		}
	}

	d.CloseDb()
	// t.Logf("---TestParseShadowsocks---Result(%+v)---", nds)
}

func TestPostgresQuery(t *testing.T) {
	testDb(t, "postgres")
	testDb(t, "odoo")
}

func testDb(t *testing.T, dbname string) {
	var err error
	d := NewEasyDb("postgres", "127.0.0.1", "postgres", "postgres", dbname, 5432)
	data := make(map[string]any, 2)
	// FROM 后面不能放占位符，应接真实的数据表名。否则报错【预处理SQL语句失败: pq: 语法错误 在 "$1" 或附近的】
	// err = d.GetOneData("SELECT id, name, age, wallet_balance FROM users WHERE id = $1", &data, 10)
	err = d.GetOneData("SELECT id ID, name AS 姓名, age AS 年龄, wallet_balance AS 钱包余额 FROM users ORDER BY id DESC", data)
	if err != nil {
		t.Error(err)
	}
	user := User{}
	err = d.GetOneData("SELECT id, name, age, wallet_balance FROM users ORDER BY id DESC", &user)
	if err != nil {
		t.Error(err)
	}
	t.Logf("---GetOneData--dbname(%s)--resultMap(%+v)--user(%+v)----\n", dbname, data, user)
	var datalist []map[string]any
	err = d.GetMany("SELECT id, name, age, wallet_balance FROM users", &datalist)
	if err != nil {
		t.Error(err)
	}
	for i := range datalist {
		t.Logf("---GetMany--row(%d)---resultMap(%+v)----\n", i, datalist[i])
	}
	var users []User
	err = d.GetMany("SELECT id, name, age, wallet_balance FROM users", &users)
	if err != nil {
		t.Error(err)
	}
	for i, user := range users {
		t.Logf("---GetMany--row(%d)---result(%+v)----\n", i, user)
	}
}

type User struct {
	ID            int     `db:"id"`
	Name          string  `db:"name"`
	Age           int     `db:"age"`
	WalletBalance float64 `db:"wallet_balance"`
}
