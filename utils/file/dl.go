// Package file 文件实用工具
package file

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	nochkcrtcli = &http.Client{Transport: tr}
)

// DownloadTo 下载到路径
//nolint: bodyclose
func DownloadTo(url, file string, chkcrt bool) error {
	var resp *http.Response
	var err error
	if chkcrt {
		resp, err = http.Get(url)
	} else {
		resp, err = nochkcrtcli.Get(url)
	}
	if err == nil {
		var f *os.File
		f, err = os.Create(file)
		if err == nil {
			_, err = io.Copy(f, resp.Body)
			f.Close()
		}
		resp.Body.Close()
	}
	return err
}
