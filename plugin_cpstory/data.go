package cpstory

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/sql"
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
		err := db.Create("cp_story", &cpstory{})
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
