package score

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/txt2img"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	cachePath = dbpath + "cache/"
	dbpath    = "data/score/"
	dbfile    = dbpath + "score.db"
)

// sdb 得分数据库
var sdb *scoredb

// 加载数据库
func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_, err := file.GetLazyData(txt2img.BoldFontFile, false, true)
		if err != nil {
			panic(err)
		}
		_, err = file.GetLazyData(txt2img.FontFile, false, true)
		if err != nil {
			panic(err)
		}
		_ = os.MkdirAll(dbpath, 0755)
		os.RemoveAll(cachePath)
		_ = os.MkdirAll(cachePath, 0755)
		sdb = initialize(dbfile)
		log.Println("[score]加载score数据库")
	}()
}
