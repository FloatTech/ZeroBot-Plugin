package math

import "strconv"

// Str2Int64 string to int64
func Str2Int64(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}
