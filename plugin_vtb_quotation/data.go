package vtbquotation

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
)

const pburl = "https://codechina.csdn.net/u011570312/ZeroBot-Plugin/-/raw/master/" + dbpath

// 加载数据库
func init() {
	if !file.IsExist(dbpath) { // 如果没有数据库，则从 url 下载
		f, err := os.Create(dbpath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		resp, err := http.Get(pburl)
		if err == nil {
			defer resp.Body.Close()
			if resp.ContentLength > 0 {
				log.Printf("[vtb]从镜像下载数据库%d字节...", resp.ContentLength)
				data, err := io.ReadAll(resp.Body)
				if err == nil && len(data) > 0 {
					_, _ = f.Write(data)
				}
				panic(err)
			}
		}
		panic(err)
	}
}
