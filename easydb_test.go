package easydb

import (
	"testing"
	// "database/sql"
)

func TestPostgresAdd(t *testing.T) {
	// var sqldb *sql.DB
	var err error
	sqlCreateTable := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50),
        age INT
    )`

	// sqldb, _ = sql.Open("postgres", "user=postgres password=postgres dbname=postgres host=127.0.0.1 port=5432 sslmode=disable search_path=public")
	// d = NewEasyDbBySqlDB(sqldb)
	d := NewEasyDb("postgres", "127.0.0.1", "postgres", "postgres", "postgres", 5432)
	_, err = d.Exec(sqlCreateTable)
	if err != nil {
		t.Error(err)
	}
	// 插入数据
	_, err = d.Exec("INSERT INTO users (name, age) VALUES ($1, $2)", "Hankin3", 3)
	if err != nil {
		t.Error(err)
	}
	d.CloseDb()
	// t.Logf("---TestParseShadowsocks---Result(%+v)---", nds)
}

func TestPostgresQuery(t *testing.T) {
	var err error
	d := NewEasyDb("postgres", "127.0.0.1", "postgres", "postgres", "postgres", 5432)
	data := make(map[string]interface{}, 3)
	// FROM 后面不能放占位符，应接真实的数据表名。否则报错【预处理SQL语句失败: pq: 语法错误 在 "$1" 或附近的】
	// err = d.GetOneData("SELECT id, name, age FROM users WHERE id = $1", &data, 10)
	err = d.GetOneData("SELECT id ID, name AS 姓名, age AS 年龄 FROM users ORDER BY id DESC", &data)
	if err != nil {
		t.Error(err)
	}
	t.Logf("---GetOneData--resultRAW(%+v)---resultDecode(%+v)--\n", data, d.DecodeBase64(data))
	var datalist []map[string]interface{}
	err = d.GetMany("SELECT id, name, age FROM users", &datalist)
	if err != nil {
		t.Error(err)
	}
	t.Logf("---GetMany--result(%+v)----\n", datalist)
}
