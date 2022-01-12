package score

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/zbputils/process"
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
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		os.RemoveAll(cachePath)
		_ = os.MkdirAll(cachePath, 0755)
		sdb = initialize(dbfile)
		log.Println("[score]加载score数据库")
	}()
}
