package plugin_book_review

import (
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

const dburl = "https://codechina.csdn.net/anto_july/bookreview/-/raw/master/bookreview.db"

// 加载数据库
func init() {
	go func() {
		process.SleepAbout1sTo2s()
		// os.RemoveAll(dbpath)
		_ = os.MkdirAll(dbpath, 0755)
		if !file.IsExist(dbfile) { // 如果没有数据库，则从 url 下载
			f, err := os.Create(dbfile)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			resp, err := http.Get(dburl)

			if err == nil {
				defer resp.Body.Close()
				if resp.ContentLength > 0 {
					log.Printf("[bookreview]从镜像下载数据库%d字节...", resp.ContentLength)
					data, err := io.ReadAll(resp.Body)
					if err == nil && len(data) > 0 {
						_, _ = f.Write(data)
						return
					}
					panic(err)
				}
			}
			panic(err)
		}
	}()
}
