package bilibilipush

import (
	"github.com/FloatTech/zbputils/process"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	cachePath = dbpath + "cache/"
	dbpath    = "data/bilibilipush/"
	dbfile    = dbpath + "push.db"
)

// bdb bilibili推送数据库
var bdb *bilibilipushdb

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		os.RemoveAll(cachePath)
		_ = os.MkdirAll(cachePath, 0755)
		bdb = initialize(dbfile)
		log.Println("[bilibilipush]加载bilibilipush数据库")
	}()
}
