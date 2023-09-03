// Package jandan 煎蛋网无聊图
package jandan

import (
	"fmt"
	"hash/crc64"
	"regexp"
	"strconv"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	api = "http://jandan.net/pic"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "煎蛋网无聊图",
		Help:             "- 来份[屌|弔|吊]图\n- 更新[屌|弔|吊]图\n",
		PublicDataFolder: "Jandan",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = engine.DataFolder() + "pics.db"
		_, _ = engine.GetLazyData("pics.db", false)
		err := db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Create("picture", &picture{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		n, err := db.Count("picture")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		logrus.Printf("[jandan]读取%d张图片", n)
		return true
	})

	engine.OnRegex(`来份[屌|弔|吊]图`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			u, err := getRandomPicture()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image(u))
		})

	engine.OnRegex(`更新[屌|弔|吊]图`, zero.SuperUserPermission, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女更新中...")
			webpageURL := api
			doc, err := htmlquery.LoadURL(webpageURL)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			re := regexp.MustCompile(`\d+`)
			pageTotal, err := strconv.Atoi(re.FindString(htmlquery.FindOne(doc, "//*[@id='comments']/div[2]/div/span[@class='current-comment-page']/text()").Data))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		LOOP:
			for i := 0; i < pageTotal; i++ {
				logrus.Debugln("[jandan]", fmt.Sprintf("处理第%d/%d页...", i, pageTotal))
				doc, err = htmlquery.LoadURL(webpageURL)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				picList, err := htmlquery.QueryAll(doc, "//*[@class='view_img_link']")
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
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
