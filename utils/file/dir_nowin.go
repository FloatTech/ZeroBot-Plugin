//go:build !windows
// +build !windows

package file

import "os"

// Pwd 获取当前路径
func Pwd() (path string) {
	path, _ = os.Getwd()
	return
}
