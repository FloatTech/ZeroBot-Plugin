package omikuji

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

const (
	dbpath = "data/Omikuji/"
	dbfile = dbpath + "kuji.db"
)

var db = &sql.Sqlite{DBPath: dbfile}

func init() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()
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
