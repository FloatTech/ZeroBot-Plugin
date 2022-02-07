package jandan

import (
	"github.com/FloatTech/ZeroBot-Plugin/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/sql"
	log "github.com/sirupsen/logrus"
	"os"
)

const dbpath = "data/jandan/"
const dbfile = dbpath + "picture.db"

var db = &sql.Sqlite{DBPath: dbfile}

// 加载数据库
func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		err := db.Create("picture", &picture{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("picture")
		if err != nil {
			panic(err)
		}
		log.Printf("[picture]读取%d张图片", n)
	}()
}
