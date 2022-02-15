// Package jandan 煎蛋网无聊图
package jandan

import (
	"fmt"
	"hash/crc64"
	"regexp"
	"strconv"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"
	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	api = "http://jandan.net/pic"
)

func init() {
	engine := control.Register("jandan", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "煎蛋网无聊图\n- 来份屌图\n- 更新屌图\n",
		PublicDataFolder: "Jandan",
	})

	go func() {
		dbpath := engine.DataFolder()
		db.DBPath = dbpath + "pics.db"
		defer order.DoneOnExit()()
		_, _ = file.GetLazyData(db.DBPath, false, false)
		err := db.Create("picture", &picture{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("picture")
		if err != nil {
			panic(err)
		}
		logrus.Printf("[jandan]读取%d张图片", n)
	}()

	engine.OnFullMatch("来份屌图").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			u, err := getRandomPicture()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Image(u))
		})

	engine.OnFullMatch("更新屌图", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女更新中...")
			webpageURL := api
			doc, err := htmlquery.LoadURL(webpageURL)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			re := regexp.MustCompile(`\d+`)
			pageTotal, err := strconv.Atoi(re.FindString(htmlquery.FindOne(doc, "//*[@id='comments']/div[2]/div/span[@class='current-comment-page']/text()").Data))
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
		LOOP:
			for i := 0; i < pageTotal; i++ {
				logrus.Infoln("[jandan]", fmt.Sprintf("处理第%d/%d页...", i, pageTotal))
				doc, err = htmlquery.LoadURL(webpageURL)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				picList, err := htmlquery.QueryAll(doc, "//*[@class='view_img_link']")
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				if len(picList) != 0 {
					for _, v := range picList {
						u := "https:" + v.Attr[0].Val
						i := crc64.Checksum(binary.StringToBytes(u), crc64.MakeTable(crc64.ISO))
						mu.RLock()
						ok := db.CanFind("picture", "where id="+strconv.FormatUint(i, 10))
						mu.RUnlock()
						if !ok {
							mu.Lock()
							_ = db.Insert("picture", &picture{ID: i, URL: u})
							mu.Unlock()
						} else {
							// 开始重复，说明之后都是重复
							break LOOP
						}
					}
				}
				if i != pageTotal-1 {
					webpageURL = "https:" + htmlquery.FindOne(doc, "//*[@id='comments']/div[@class='comments']/div[@class='cp-pagenavi']/a[@class='previous-comment-page']").Attr[1].Val
				}
			}
			ctx.Send("更新完成!")
		})
}
