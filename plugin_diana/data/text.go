// Package data 加载位于 datapath 的小作文
package data

import (
	"crypto/md5"
	"errors"
	"math/rand"
	"os"
	"sync"

	"github.com/RomiChan/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

const (
	datapath = "data/Diana"
	pbfile   = datapath + "/text.pb"
)

var (
	compo Composition
	// m 小作文保存锁
	m sync.Mutex
	// md5s 验证重复
	md5s []*[16]byte
)

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		err1 := LoadText()
		if err1 == nil {
			arrl := len(compo.Array)
			log.Printf("[Diana]读取%d条小作文", arrl)
			md5s = make([]*[16]byte, arrl)
			for i, t := range compo.Array {
				m := md5.Sum(helper.StringToBytes(t))
				md5s[i] = &m
			}
		} else {
			log.Printf("[Diana]读取小作文错误：%v", err1)
		}
	}()
}

// LoadText 加载小作文
func LoadText() error {
	data, err := file.GetLazyData(pbfile, true, false)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, &compo)
}

// AddText 添加小作文
func AddText(txt string) error {
	sum := md5.Sum(helper.StringToBytes(txt))
	if txt != "" && !isin(&sum) {
		m.Lock()
		defer m.Unlock()
		compo.Array = append(compo.Array, txt)
		md5s = append(md5s, &sum)
		return savecompo()
	}
	return nil
}

// RandText 随机小作文
func RandText() string {
	return (compo.Array)[rand.Intn(len(compo.Array)-1)+1]
}

// HentaiText 发大病
func HentaiText() string {
	return (compo.Array)[0]
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
	data, err := proto.Marshal(&compo)
	if err == nil {
		if file.IsExist(datapath) {
			f, err1 := os.OpenFile(pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			if err1 == nil {
				_, err2 := f.Write(data)
				f.Close()
				return err2
			}
			return err1
		}
		return errors.New("datapath is not exist")
	}
	return err
}
