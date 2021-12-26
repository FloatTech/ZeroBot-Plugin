package cpstory

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	dbpath = "data/CpStory/"
	dbfile = dbpath + "cp.db"
)

var db = &sql.Sqlite{DBPath: dbfile}

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		// os.RemoveAll(dbpath)
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, true)
		err := db.Create("cp_story", &CpStory{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("cp_story")
		if err != nil {
			panic(err)
		}
		log.Printf("[cpstory]读取%d条故事", n)
	}()
}
