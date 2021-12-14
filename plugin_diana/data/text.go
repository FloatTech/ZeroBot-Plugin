// Package data 加载位于 datapath 的小作文
package data

import (
	"crypto/md5"
	"os"
	"unsafe"

	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

const (
	datapath = "data/Diana"
	dbfile   = datapath + "/text.db"
)

var db = sql.Sqlite{DBPath: dbfile}

type Text struct {
	Id   int64  `db:"id"`
	Data string `db:"data"`
}

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		err = LoadText()
		if err == nil {
			err = db.Create("text", &Text{})
			if err != nil {
				panic(err)
			}
			c, _ := db.Count("text")
			log.Printf("[Diana]读取%d条小作文", c)
		} else {
			log.Printf("[Diana]读取小作文错误：%v", err)
		}
	}()
}

// LoadText 加载小作文
func LoadText() error {
	_, err := file.GetLazyData(dbfile, false, false)
	return err
}

// AddText 添加小作文
func AddText(txt string) error {
	s := md5.Sum(helper.StringToBytes(txt))
	i := *(*int64)(unsafe.Pointer(&s))
	return db.Insert("text", &Text{Id: i, Data: txt})
}

// RandText 随机小作文
func RandText() string {
	var t Text
	err := db.Pick("text", &t)
	if err != nil {
		return err.Error()
	}
	return t.Data
}

// HentaiText 发大病
func HentaiText() string {
	var t Text
	err := db.Find("text", &t, "where id = -3802576048116006195")
	if err != nil {
		return err.Error()
	}
	return t.Data
}
