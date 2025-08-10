package sql

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iotames/miniutils"
)

//go:embed *.sql
var sqlFS embed.FS

// getSqlText 获取sql文本
// 优先从custom/sql自定义目录读取sql。如找不到SQL文件，则从默认的sql目录中读取。如再找不到文件，则从内嵌文件中读取。
func getSqlText(fpath string) (sqlTxt string, err error) {
	defaultFilePath := filepath.Join("sql", fpath)
	// 优先从custom/sql自定义目录读取sql。如找不到SQL文件，则从默认的sql目录中读取。
	sqlTxt, err = GetTextByFilePath(defaultFilePath, "")
	if sqlTxt == "" {
		// 找不到文件，从内嵌的文件中读取sql文件
		var sqlBytes []byte
		sqlBytes, err = sqlFS.ReadFile(fpath)
		if err != nil {
			return
		}
		sqlTxt = string(sqlBytes)
	}
	return
}

// GetSQL 获取sql文本
// replaceList 字符串列表，依次替换SQL文本中的?占位符
// TODO 需要强调占位符与通配符的区别，比如%和_在LIKE子句中不是占位符，而是通配符，需要和参数化查询中的占位符区分开。
func GetSQL(fpath string, replaceList ...string) (string, error) {
	sqlTxt, err := getSqlText(fpath)
	if err != nil {
		return "", err
	}
	for _, rerplaceStr := range replaceList {
		sqlTxt = strings.Replace(sqlTxt, "?", rerplaceStr, 1)
	}
	return sqlTxt, nil
}

func LsDir() []string {
	entries, err := sqlFS.ReadDir(".")
	if err != nil {
		panic(err)
	}
	var filenames []string
	for _, entry := range entries {
		filenames = append(filenames, entry.Name())
		if entry.IsDir() {
			// 读取子目录中的文件
			subEntries, err := sqlFS.ReadDir(entry.Name())
			if err != nil {
				panic(err)
			}
			// 将子目录中的文件名添加到列表中
			for _, subEntry := range subEntries {
				filenames = append(filenames, entry.Name()+"/"+subEntry.Name())
			}
		}
	}
	return filenames
}

func getExistFile(defaultFilePath string, customFilePath string) string {
	if miniutils.IsPathExists(customFilePath) {
		return customFilePath
	}
	if miniutils.IsPathExists(defaultFilePath) {
		return defaultFilePath
	}
	return ""
}

func GetTextByFilePath(defaultFilePath string, customFilePath string) (content string, err error) {
	fpath := getExistFile(defaultFilePath, customFilePath)
	if fpath == "" {
		return "", fmt.Errorf("file not found: %s", defaultFilePath)
	}
	return miniutils.ReadFileToString(fpath)
}
