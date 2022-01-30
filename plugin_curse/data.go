package curse

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/sql"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	dbpath = "data/Curse/"
	dbfile = dbpath + "curse.db"
)

var (
	db = &sql.Sqlite{DBPath: dbfile}
)

// 加载数据库
func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		_, err := file.GetLazyData(dbfile, false, true)
		if err != nil {
			panic(err)
		}
		err = db.Create("curse", &curse{})
		if err != nil {
			panic(err)
		}
		c, _ := db.Count("curse")
		logrus.Infoln("[curse]加载", c, "条骂人语录")
	}()
}
