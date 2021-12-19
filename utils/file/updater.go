package file

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"unsafe"

	reg "github.com/fumiama/go-registry"
	"github.com/sirupsen/logrus"
)

const (
	dataurl = "https://codechina.csdn.net/u011570312/ZeroBot-Plugin/-/raw/master/"
)

var (
	registry = reg.NewRegReader("reilia.eastasia.azurecontainer.io:32664", "fumiama")
	lzmu     sync.Mutex
)

func GetLazyData(path string, isReturnDataBytes, isDataMustEqual bool) ([]byte, error) {
	var data []byte
	var resp *http.Response
	var filemd5 *[16]byte
	var ms string

	u := dataurl + path
	lzmu.Lock()
	logrus.Infoln("[file]检查懒加载文件:", path)
	err := registry.Connect()
	if err != nil {
		logrus.Errorln("[file]无法连接到md5验证服务器，请自行确保下载文件的正确性:", err)
	} else {
		ms, err = registry.Get(path)
		if err != nil || len(ms) != 16 {
			logrus.Errorln("[file]获取md5失败，请自行确保下载文件 %s 的正确性:", path, err)
		} else {
			filemd5 = (*[16]byte)(*(*unsafe.Pointer)(unsafe.Pointer(&ms)))
			logrus.Infoln("[file]从验证服务器获得文件md5:", hex.EncodeToString(filemd5[:]))
		}
	}
	_ = registry.Close()
	lzmu.Unlock()

	if IsExist(path) {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if filemd5 != nil {
			if md5.Sum(data) == *filemd5 {
				logrus.Infoln("[file]文件md5匹配，文件已存在且为最新")
				goto ret
			} else if !isDataMustEqual {
				logrus.Warnln("[file]文件", path, "md5不匹配，但不主动更新")
				goto ret
			}
			logrus.Infoln("[file]文件md5不匹配，开始更新文件")
		} else {
			logrus.Warnln("[file]文件", path, "存在，已跳过md5检查")
			goto ret
		}
	}

	// 下载
	resp, err = http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.ContentLength <= 0 {
		return nil, errors.New("resp body len <= 0")
	}
	logrus.Printf("[file]从镜像下载数据%d字节...", resp.ContentLength)
	// 读取数据
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("read body len <= 0")
	}
	if filemd5 != nil {
		if md5.Sum(data) == *filemd5 {
			logrus.Infoln("[file]文件下载完成，md5匹配，开始保存")
		} else {
			logrus.Errorln("[file]文件", path, "md5不匹配，下载失败")
			return nil, errors.New("file md5 mismatch")
		}
	} else {
		logrus.Warnln("[file]文件", path, "下载完成，已跳过md5检查，开始保存")
	}
	// 写入数据
	err = os.WriteFile(path, data, 0644)
ret:
	if isReturnDataBytes {
		return data, err
	}
	return nil, err
}
