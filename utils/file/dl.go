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
