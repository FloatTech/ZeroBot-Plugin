// Package bookreview 书评
package bookreview

import (
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

func init() {
	engine := control.Register("bookreview", order.PrioBookReview, &control.Options{
		DisableOnDefault: false,
		Help:             "哀伤雪刃推书记录\n- 书评[xxx]\n- 随机书评",
	})

	// 中文、英文、数字但不包括下划线等符号
	engine.OnRegex("^书评([\u4E00-\u9FA5A-Za-z0-9]{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			b := getBookReviewByKeyword(ctx.State["regex_matched"].([]string)[1])
			data, err := text.RenderToBase64(b.BookReview, text.FontFile, 400, 20)
			if err != nil {
				log.Println("err:", err)
			}
			if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnFullMatch("随机书评").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			br := getRandomBookReview()
			data, err := text.RenderToBase64(br.BookReview, text.FontFile, 400, 20)
			if err != nil {
				log.Println("err:", err)
			}
			if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
