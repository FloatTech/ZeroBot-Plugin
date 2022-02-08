package bookreview

import (
	"os"

	log "github.com/sirupsen/logrus"

	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const dbpath = "data/BookReview/"
const dbfile = dbpath + "bookreview.db"

var db = &sql.Sqlite{DBPath: dbfile}

// 加载数据库
func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		// os.RemoveAll(dbpath)
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, true)
		err := db.Create("bookreview", &book{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("bookreview")
		if err != nil {
			panic(err)
		}
		log.Printf("[bookreview]读取%d条书评", n)
	}()
}
