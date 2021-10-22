// Package file 文件实用工具
package file

import (
	"io"
	"net/http"
	"os"
)

// DownloadTo 下载到路径
func DownloadTo(url, file string) error {
	resp, err := http.Get(url)
	if err == nil {
		var f *os.File
		f, err = os.Create(file)
		if err == nil {
			_, err = io.Copy(f, resp.Body)
			resp.Body.Close()
			f.Close()
		}
	}
	return err
}

// IsExist 文件/路径存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// IsExist 文件/路径不存在
func IsNotExist(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsNotExist(err)
}
