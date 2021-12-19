//go:build windows
// +build windows

package file

import (
	"os"
	"strings"
)

// Pwd 获取当前路径的正斜杠表示
func Pwd() string {
	path, _ := os.Getwd()
	return strings.ReplaceAll(path, "\\", "/")
}
