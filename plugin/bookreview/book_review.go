// Package bookreview 书评
package bookreview

import (
	"time"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "哀伤雪刃推书书评",
		Help:             "- 书评[xxx]\n- 随机书评",
		PublicDataFolder: "BookReview",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = engine.DataFolder() + "bookreview.db"
		// os.RemoveAll(dbpath)
		_, _ = engine.GetLazyData("bookreview.db", true)
		err := db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Create("bookreview", &book{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		n, err := db.Count("bookreview")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		log.Infof("[bookreview]读取%d条书评", n)
		return true
	})

	// 中文、英文、数字但不包括下划线等符号
	engine.OnRegex("^书评([\u4E00-\u9FA5A-Za-z0-9]{1,25})$", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			b := getBookReviewByKeyword(ctx.State["regex_matched"].([]string)[1])
			data, err := text.RenderToBase64(b.BookReview, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnFullMatch("随机书评", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			br := getRandomBookReview()
			data, err := text.RenderToBase64(br.BookReview, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
