package omikuji

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

const (
	dbpath = "data/omikuji/"
	dbfile = dbpath + "signature.db"
	dburl  = "https://codechina.csdn.net/anto_july/bookreview/-/raw/master/signature.db?inline=false"
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
		err := db.Create("signature", &signature{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("signature")
		if err != nil {
			panic(err)
		}
		log.Printf("[signature]读取%d条签文", n)
	}()

}
