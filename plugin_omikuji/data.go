package omikuji

import (
	"os"

	log "github.com/sirupsen/logrus"

	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	dbpath = "data/Omikuji/"
	dbfile = dbpath + "kuji.db"
)

var db = &sql.Sqlite{DBPath: dbfile}

func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		_, _ = file.GetLazyData(dbfile, false, true)
		err := db.Create("kuji", &kuji{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("kuji")
		if err != nil {
			panic(err)
		}
		log.Printf("[kuji]读取%d条签文", n)
	}()
}
