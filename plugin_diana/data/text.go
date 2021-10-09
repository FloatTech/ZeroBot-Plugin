// Package data 加载位于 datapath 的小作文
package data

import (
	"crypto/md5"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/data"
)

const (
	datapath = "data/Diana"
	pbfile   = datapath + "/text.pb"
	pburl    = "https://codechina.csdn.net/u011570312/ZeroBot-Plugin/-/raw/master/" + pbfile
)

var (
	compo Composition
	// Array 小作文数组指针
	Array = &compo.Array
	// m 小作文保存锁
	m sync.Mutex
	// md5s 验证重复
	md5s []*[16]byte
)

func init() {
	go func() {
		time.Sleep(time.Second)
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		err1 := LoadText()
		if err1 == nil {
			arrl := len(*Array)
			log.Printf("[Diana]读取%d条小作文", arrl)
			md5s = make([]*[16]byte, arrl)
			for i, t := range *Array {
				m := md5.Sum(data.Str2bytes(t))
				md5s[i] = &m
			}
		} else {
			log.Printf("[Diana]读取小作文错误：%v", err1)
		}
	}()
}

// LoadText 加载小作文
func LoadText() error {
	if _, err := os.Stat(pbfile); err == nil || os.IsExist(err) {
		f, err := os.Open(pbfile)
		if err == nil {
			defer f.Close()
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					return compo.Unmarshal(data)
				}
			}
			return err1
		}
	} else { // 如果没有小作文，则从 url 下载
		f, err := os.Create(pbfile)
		if err != nil {
			return err
		}
		defer f.Close()
		resp, err := http.Get(pburl)
		if err == nil {
			defer resp.Body.Close()
			if resp.ContentLength > 0 {
				log.Printf("[Diana]从镜像下载小作文%d字节...", resp.ContentLength)
				data, err := io.ReadAll(resp.Body)
				if err == nil && len(data) > 0 {
					_, _ = f.Write(data)
					return compo.Unmarshal(data)
				}
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

// AddText 添加小作文
func AddText(txt string) error {
	sum := md5.Sum(data.Str2bytes(txt))
	if txt != "" && !isin(&sum) {
		m.Lock()
		defer m.Unlock()
		compo.Array = append(compo.Array, txt)
		md5s = append(md5s, &sum)
		return savecompo()
	}
	return nil
}

func isin(sum *[16]byte) bool {
	for _, t := range md5s {
		if *t == *sum {
			return true
		}
	}
	return false
}

// savecompo 同步保存作文
func savecompo() error {
	data, err := compo.Marshal()
	if err == nil {
		if _, err := os.Stat(datapath); err == nil || os.IsExist(err) {
			f, err1 := os.OpenFile(pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			if err1 == nil {
				_, err2 := f.Write(data)
				f.Close()
				return err2
			}
			return err1
		}
	}
	return err
}
