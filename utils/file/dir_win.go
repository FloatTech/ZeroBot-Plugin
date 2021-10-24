//go:build windows
// +build windows

package file

import (
	"os"
	"strings"
)

func Pwd() string {
	path, _ := os.Getwd()
	return strings.ReplaceAll(path, "\\", "/")
}
