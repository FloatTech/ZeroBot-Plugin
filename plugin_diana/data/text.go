// Package data 加载位于 datapath 的小作文
package data

import (
	"crypto/md5"
	"encoding/binary"
	"os"

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

type text struct {
	ID   int64  `db:"id"`
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
			err = db.Create("text", &text{})
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
	i := binary.LittleEndian.Uint64(s[:8])
	return db.Insert("text", &text{ID: int64(i), Data: txt})
}

// RandText 随机小作文
func RandText() string {
	var t text
	err := db.Pick("text", &t)
	if err != nil {
		return err.Error()
	}
	return t.Data
}

// HentaiText 发大病
func HentaiText() string {
	var t text
	err := db.Find("text", &t, "where id = -3802576048116006195")
	if err != nil {
		return err.Error()
	}
	return t.Data
}
