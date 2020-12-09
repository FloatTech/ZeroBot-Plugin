package utils

import (
	"io/ioutil"
	"os"
	"strconv"
)

func PathExecute() string {
	dir, _ := os.Getwd()
	return dir + "\\"
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func ReadAllText(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func WriteAllText(path, text string) {
	_ = ioutil.WriteFile(path, []byte(text), 0644)
}

func CreatePath(path string) error {
	err := os.MkdirAll(path, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Int2Str(val int64) string {
	str := strconv.FormatInt(val, 10)
	return str
}

func Str2Int(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}
