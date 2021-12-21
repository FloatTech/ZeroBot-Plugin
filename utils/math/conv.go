package math

import "strconv"

func Str2Int64(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}
