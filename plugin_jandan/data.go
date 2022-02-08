package jandan

import (
	"os"
	"sync"

	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const dbpath = "data/Jandan/"
const dbfile = dbpath + "pics.db"

var db = &sql.Sqlite{DBPath: dbfile}
var mu sync.RWMutex

type picture struct {
	ID  uint64 `db:"id"`
	URL string `db:"url"`
}

func getRandomPicture() (u string, err error) {
	var p picture
	mu.RLock()
	err = db.Pick("picture", &p)
	mu.RUnlock()
	u = p.URL
	return
}

// 加载数据库
func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, false)
		err := db.Create("picture", &picture{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("picture")
		if err != nil {
			panic(err)
		}
		log.Printf("[jandan]读取%d张图片", n)
	}()
}
