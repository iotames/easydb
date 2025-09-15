## 简介

简单的Go数据库接口程序。对 `database/sql` 标准库和各数据库驱动进行封装，使用通用接口操作数据库，无需关心底层差异。

1. 轻量极简，没有过多依赖包袱。数据驱动仅在使用时，显式声明，按需导入。
2. 兼容性极佳，支持大多数数据库，如MySQL、PostgreSQL、SQLite、SQL Server、Oracle等。
3. 仅在官方库基础上做轻量级封装，摆脱ORM等框架束缚，回归底层，重新找回对数据库的掌控。


## 各数据库驱动实现

- MySQL：github.com/go-sql-driver/mysql
- PostgreSQL：github.com/lib/pq
- SQLite：github.com/mattn/go-sqlite3
- SQL Server：github.com/denisenkom/go-mssqldb
- Oracle: github.com/godror/godror

1. 数据库连接:

```go
	import (
	"github.com/iotames/easydb"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	)
func main() {
	d1 := easydb.NewEasyDb("mysql", "127.0.0.1", "root", "password", "testdb", 3306)
	d2 := easydb.NewEasyDb("postgres", "127.0.0.1", "username", "password", "testdb", 5432)
	// 测试连接d1
	if err := d1.Ping(); err != nil {
		log.Fatal(err)
	}
	// 关闭整个d1连接池
	defer d1.CloseDb()
	// 测试连接d2
	if err := d2.Ping(); err != nil {
		log.Fatal(err)
	}
	// 关闭整个d2连接池
	defer d2.CloseDb()

	// 用原始的DSN数据源字符串，创建数据库操作对象
	var sqldb *sql.DB
	var err error
	// sqldb, err := sql.Open("sqlite3", "./mydb.sqlite")
	// sqldb, err = sql.Open("postgres", "user=postgres password=secret dbname=dbname host=127.0.0.1 port=5432 sslmode=disable search_path=myschema")
	sqldb, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/dbname")
	if err != nil {
		panic(err)
	}
	// 创建数据库操作对象
	d := easydb.NewEasyDbBySqlDB(sqldb)
	// 测试连接d
	if err := d.Ping(); err != nil {
		log.Fatal(err)
	}
	// 关闭整个d连接池
	defer d.CloseDb()
}
```

2. 执行SQL语句，写入数据

```go
import (
	"github.com/iotames/easydb"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
	d := easydb.NewEasyDb("mysql", "127.0.0.1", "root", "password", "testdb", 3306)
	// 创建数据表
	sqlCreateTable := `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50),
        age INT,
		wallet_balance DECIMAL(10,2) DEFAULT 0.00
    )`
	_, err = d.Exec(sqlCreateTable)
	if err != nil {
		fmt.Println("Error creating table:", err)
	}

	// 插入数据
	for i := 1; i <= 5; i++ {
		_, err = d.Exec("INSERT INTO users (name, age, wallet_balance) VALUES ($1, $2, $3)", fmt.Sprintf("Hankin%d", i), i, float64(i*10))
		if err != nil {
			fmt.Println("Error inserting data:", err)
		}
	}
}
```

3. 获取数据

```go

type User struct {
	ID            int     `db:"id"`
	Name          string  `db:"name"`
	Age           int     `db:"age"`
	WalletBalance float64 `db:"wallet_balance"`
}

func main() {
	d := easydb.NewEasyDb("postgres", "127.0.0.1", "username", "password", "testdb", 5432)
	// 获取一条数据，传入字典或字典的指针
	data := make(map[string]interface{})
	d.GetOneData("SELECT id, name, age, wallet_balance FROM $1", data, "users")
	// 传字典的指针亦可
	// d.GetOneData("SELECT id, name, age, wallet_balance FROM $1", &data, "users")

	// 获取一条数据，传入结构体指针
	user := User{}
	d.GetOneData("SELECT id, name, age, wallet_balance FROM users ORDER BY id DESC", &user)
	
	fmt.Printf("-----GetOneData--data(%+v)--user(%+v)--\n", data, user)

	// 获取多条数据，传入字典的切片的指针
	var datalist []map[string]interface{}
	d.GetMany("SELECT id, name, age, wallet_balance FROM users", &datalist)
	fmt.Printf("-----GetMany--result(%+v)----\n", datalist)

    // 获取多条数据，传入结构体的切片的指针
	var users []User
	err := d.GetMany("SELECT id, name, age, wallet_balance FROM users", &users)
	if err != nil {
		fmt.Println("Error getting data:", err)
		return
	}
	for i, user := range users {
		fmt.Printf("---GetMany--row(%d)---result(%+v)----\n", i, user)
	}

}

```