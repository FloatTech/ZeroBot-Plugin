// Package math 计算实用工具
package math

// Max 返回两数最大值，该函数将被内联
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 返回两数最小值，该函数将被内联
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Abs 返回绝对值，该函数将被内联
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
