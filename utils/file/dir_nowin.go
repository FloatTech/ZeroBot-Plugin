//go:build !windows
// +build !windows

package file

import "os"

func Pwd() (path string) {
	path, _ = os.Getwd()
	return
}
