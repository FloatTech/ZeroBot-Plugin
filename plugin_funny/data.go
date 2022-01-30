package funny

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

type joke struct {
	ID   uint32 `db:"id"`
	Text string `db:"text"`
}

const (
	dbpath = "data/Funny/"
	dbfile = dbpath + "jokes.db"
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
		err = db.Create("jokes", &joke{})
		if err != nil {
			panic(err)
		}
		c, _ := db.Count("jokes")
		logrus.Infoln("[funny]加载", c, "个笑话")
	}()
}
