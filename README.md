## 简介

简单的Go数据库接口程序。对 `database/sql` 标准库和各数据库驱动进行封装，使用通用接口操作数据库，无需关心底层差异。


## 各数据库驱动实现

- MySQL：github.com/go-sql-driver/mysql
- PostgreSQL：github.com/lib/pq
- SQLite：github.com/mattn/go-sqlite3
- SQL Server：github.com/denisenkom/go-mssqldb
- Oracle: github.com/godror/godror

连接示例

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var sqldb *sql.DB
var err error

// 连接数据库
sqldb, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/dbname")
// sqldb, err = sql.Open("postgres", "user=postgres password=secret dbname=dbname host=127.0.0.1 port=5432 sslmode=disable search_path=myschema")
// sqldb, err := sql.Open("sqlite3", "./mydb.sqlite")

if err != nil {
	panic(err)
}

// 测试连接
if err := sqldb.Ping(); err != nil {
	log.Fatal(err)
}

// 这个很少用。是关闭整个连接池。
defer sqldb.Close()
```