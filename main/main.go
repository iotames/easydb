package main

import (
	"flag"
	"fmt"

	"github.com/iotames/easydb"
)

var DbDriver, DbHost, DbUser, DbPassword, DbName string
var DbPort int
var DbTable string

// driverName, dbHost, dbUser, dbPassword, dbName string, dbPort

func main() {
	var err error
	d := easydb.NewEasyDb(DbDriver, DbHost, DbUser, DbPassword, DbName, DbPort)
	err = d.Ping()
	if err != nil {
		panic(err)
	}
	// _, err = d.Exec(fmt.Sprintf(`INSERT INTO %s (name, age) VALUES ('EasyDB', 20)`, DbTable))
	// if err != nil {
	// 	panic(err)
	// }
	// // mysql使用？而Postgres使用$1 $2 $3 ...
	// holdplace := "?"
	// if DbDriver == "postgres" {
	// 	holdplace = "$1"
	// }
	// sqltxt := fmt.Sprintf(`SELECT id, name, age FROM %s`, holdplace)
	sqltxt := fmt.Sprintf(`SELECT id, name, age FROM %s`, DbTable)

	rows, err := d.Query(sqltxt)
	if err != nil {
		panic(err)
	}
	defer rows.Close() // 必须关闭释放资源
	for rows.Next() {
		var id, age int
		var name string
		rows.Scan(&id, &name, &age) // 按列顺序扫描到变量
		fmt.Sprintln(id, name, age)
	}
	if err = rows.Err(); err != nil { // 检查遍历错误
		panic(err)
	}

}

func init() {
	flag.StringVar(&DbDriver, "dbdriver", "mysql", "")
	flag.StringVar(&DbHost, "dbhost", "127.0.0.1", "")
	flag.StringVar(&DbUser, "dbuser", "root", "")
	flag.StringVar(&DbPassword, "dbpassword", "root", "")
	flag.StringVar(&DbName, "dbname", "debugdb", "")
	flag.StringVar(&DbTable, "dbtable", "users", "")
	flag.IntVar(&DbPort, "dbport", 3306, "")
	flag.Parse()
}
