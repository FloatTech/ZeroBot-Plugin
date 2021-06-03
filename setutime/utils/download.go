package utils

import (
	"crypto/md5"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
)

// urlCache 缓存并返回缓存路径
func (this *Illust) PixivPicDown(path string) (savePath string, err error) {
	url := this.ImageUrls
	pid := this.Pid
	url = strings.ReplaceAll(url, "img-original", "img-master")
	url = strings.ReplaceAll(url, "_p0", "_p0_master1200")
	url = strings.ReplaceAll(url, ".png", ".jpg")
	// 文件名为url的hash值
	savePath = path + Int2Str(pid) + ".jpg"
	// 文件存在或文件大小大于10kb
	if PathExists(savePath) && FileSize(savePath) > 10240 {
		return savePath, nil
	}

	// 模拟QQ客户端请求
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			// 绕过sni审查
			TLSClientConfig: &tls.Config{
				ServerName:         "-",
				InsecureSkipVerify: true,
			},
			// 更改dns
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("tcp", "210.140.92.142:443")
			},
		},
	}
	reqest, _ := http.NewRequest("GET", url, nil)
	reqest.Header.Set("Referer", "https://www.pixiv.net/")
	reqest.Header.Set("Host", "i.pximg.net")
	reqest.Header.Set("User-Agent", "QQ/8.2.0.1296 CFNetwork/1126")

	resp, err := client.Do(reqest)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Download failed, code %d", resp.StatusCode))
	}
	defer resp.Body.Close()
	// 写入文件
	data, _ := ioutil.ReadAll(resp.Body)
	f, _ := os.OpenFile(savePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	f.Write(data)

	return savePath, err
}

// PicHash 返回图片的 md5 值
func PicHash(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum(data)))
}
