package score

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	cachePath = dbpath + "cache/"
	dbpath    = "data/Score/"
	dbfile    = dbpath + "score.db"
)

// SDB 得分数据库
var SDB *ScoreDB

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		//os.RemoveAll(cachePath)
		_ = os.MkdirAll(cachePath, 0755)
		SDB = Initialize(dbfile)
		log.Println("[score]加载score数据库")
	}()
}
