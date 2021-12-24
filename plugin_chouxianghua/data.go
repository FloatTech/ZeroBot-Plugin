package chouxianghua

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	dbpath = "data/ChouXiangHua/"
	dbfile = dbpath + "cxh.db"
)

var db = &sql.Sqlite{DBPath: dbfile}

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		// os.RemoveAll(dbpath)
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, true)
		err := db.Create("pinyin", &Pinyin{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("pinyin")
		if err != nil {
			panic(err)
		}
		log.Printf("[chouxianghua]读取%d条拼音", n)
	}()
}
