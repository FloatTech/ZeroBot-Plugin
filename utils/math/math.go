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

// intSize is either 32 or 64.
const intSize = 32 << (^uint(0) >> 63)

// Abs 返回绝对值，该函数将被内联
func Abs(x int) int {
	// m := -1 if x < 0. m := 0 otherwise.
	m := x >> (intSize - 1)

	// In two's complement representation, the negative number
	// of any number (except the smallest one) can be computed
	// by flipping all the bits and add 1. This is faster than
	// code with a branch.
	// See Hacker's Delight, section 2-4.
	return (x ^ m) - m
}

// Abs64 返回绝对值，该函数将被内联
func Abs64(x int64) int64 {
	// m := -1 if x < 0. m := 0 otherwise.
	m := x >> (64 - 1)

	// In two's complement representation, the negative number
	// of any number (except the smallest one) can be computed
	// by flipping all the bits and add 1. This is faster than
	// code with a branch.
	// See Hacker's Delight, section 2-4.
	return (x ^ m) - m
}
