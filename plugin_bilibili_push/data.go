package bilibilipush

import (
	"github.com/FloatTech/zbputils/process"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	dbpath = "data/bilibilipush/"
	dbfile = dbpath + "push.db"
)

// bdb bilibili推送数据库
var bdb *bilibilipushdb

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		bdb = Initialize(dbfile)
		log.Println("[bilibilipush]加载bilibilipush数据库")
	}()
}
