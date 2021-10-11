package data

// min 返回两数最大值，该函数将被内联
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min 返回两数最小值，该函数将被内联
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
