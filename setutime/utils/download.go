package utils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// urlCache 缓存并返回缓存路径
func (this *Illust) PixivPicDown(path string) (savePath string, err error) {
	url := this.ImageUrls
	pid := this.Pid
	url = strings.ReplaceAll(url, "i.pximg.net", "i.pixiv.cat")
	url = strings.ReplaceAll(url, "img-original", "img-master")
	url = strings.ReplaceAll(url, "_p0", "_p0_master1200")
	url = strings.ReplaceAll(url, ".png", ".jpg")
	// TODO 文件名为url的hash值
	savePath = path + Int2Str(pid) + ".jpg"
	// TODO 文件存在或文件大小大于10kb
	if PathExists(savePath) && FileSize(savePath) > 10240 {
		return savePath, nil
	}
	zero.SendGroupMessage(CACHE_GROUP, "正在下载"+url)
	// TODO 模拟QQ客户端请求
	client := &http.Client{}
	reqest, _ := http.NewRequest("GET", url, nil)
	reqest.Header.Add("User-Agent", "QQ/8.2.0.1296 CFNetwork/1126")
	reqest.Header.Add("Net-Type", "Wifi")

	resp, err := client.Do(reqest)
	if err != nil {
		return "", err
	}
	fmt.Println(resp.StatusCode)
	if code := resp.StatusCode; code != 200 {
		return "", errors.New(fmt.Sprintf("Download failed, code %d", code))
	}
	defer resp.Body.Close()
	// TODO 写入文件
	data, _ := ioutil.ReadAll(resp.Body)
	f, _ := os.OpenFile(savePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	f.Write(data)

	return savePath, err
}

func PicHash(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%x", md5.Sum(data)))
}
