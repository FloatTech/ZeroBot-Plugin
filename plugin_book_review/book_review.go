// Package bookreview 书评
package bookreview

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

func init() {
	engine := control.Register("bookreview", &control.Options{
		DisableOnDefault: false,
		Help:             "哀伤雪刃推书记录\n- 书评[xxx]\n- 随机书评",
	})

	// 中文、英文、数字但不包括下划线等符号
	engine.OnRegex("^书评([\u4E00-\u9FA5A-Za-z0-9]{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			b := getBookReviewByKeyword(ctx.State["regex_matched"].([]string)[1])
			ctx.SendChain(message.Text(b.BookReview))
		})

	engine.OnFullMatch("随机书评").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			br := getRandomBookReview()
			ctx.SendChain(message.Text(br.BookReview))
		})
}
