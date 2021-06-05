package utils

import (
	"os"
	"strconv"
	"strings"
)

// Str2Int string --> int64
func Str2Int(str string) int64 {
	val, _ := strconv.Atoi(str)
	return int64(val)
}

// Int2Str int64 --> string
func Int2Str(val int64) string {
	str := strconv.FormatInt(val, 10)
	return str
}

// PathExecute 返回当前运行目录
func PathExecute() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir + "/"
}

// CreatePath 生成路径或文件所对应的目录
func CreatePath(path string) {
	length := len(path)
	switch {
	case path[length:] != "/":
		path = path[:strings.LastIndex(path, "/")]
	case path[length:] != "\\":
		path = path[:strings.LastIndex(path, "\\")]
	default:
		//
	}
	if !PathExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}
}

// PathExists 判断路径或文件是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// FileSize 获取文件大小
func FileSize(file string) int64 {
	if fi, err := os.Stat(file); err == nil {
		return fi.Size()
	}
	return 0
}

// Min 返回两数最小值
func Min(a, b int) int {
	switch {
	default:
		return a
	case a > b:
		return b
	case a < b:
		return a
	}
}
