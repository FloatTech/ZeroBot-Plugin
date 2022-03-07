// Package data 加载位于 datapath 的小作文
package data

import (
	"crypto/md5"
	"encoding/binary"

	sql "github.com/FloatTech/sqlite"
	binutils "github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/file"
	"github.com/sirupsen/logrus"
)

var db = sql.Sqlite{}

type text struct {
	ID   int64  `db:"id"`
	Data string `db:"data"`
}

// LoadText 加载小作文
func LoadText(dbfile string) {
	_, err := file.GetLazyData(dbfile, false, false)
	db.DBPath = dbfile
	if err != nil {
		panic(err)
	}
	err = db.Create("text", &text{})
	if err != nil {
		panic(err)
	}
	c, _ := db.Count("text")
	logrus.Printf("[Diana]读取%d条小作文", c)
}

// AddText 添加小作文
func AddText(txt string) error {
	s := md5.Sum(binutils.StringToBytes(txt))
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
