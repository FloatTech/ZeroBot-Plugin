package omikuji

import (
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

const (
	dbpath = "data/Omikuji/"
	dbfile = dbpath + "kuji.db"
	dburl  = "https://codechina.csdn.net/u011570312/ZeroBot-Plugin/-/raw/master/" + dbfile
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
		if !file.IsExist(dbfile) { // 如果没有数据库，则从 url 下载
			f, err := os.Create(dbfile)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			resp, err := http.Get(dburl)

			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			if resp.ContentLength > 0 {
				log.Printf("[omikuji]从镜像下载数据库%d字节...", resp.ContentLength)
				data, err := io.ReadAll(resp.Body)
				if err == nil && len(data) > 0 {
					_, _ = f.Write(data)
				} else {
					panic(err)
				}
			}
		}
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
