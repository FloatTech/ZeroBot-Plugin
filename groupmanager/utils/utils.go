package utils

import (
	"strconv"
)

func Int2Str(val int64) string {
	str := strconv.FormatInt(val, 10)
	return str
}

func Str2Int(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}
