// Package translation 翻译
package translation

import (
	"github.com/FloatTech/AnimeAPI/tl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.Register("translation", &control.Options{
		DisableOnDefault: false,
		Help: "翻译\n" +
			">TL 你好",
	}).OnRegex(`^>TL\s(-.{1,10}? )?(.*)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := []string{ctx.State["regex_matched"].([]string)[2]}
			data, err := tl.Translate(msg[0])
			if err != nil {
				ctx.SendChain(message.Text("Error:", data))
				return
			}
			ctx.SendChain(message.Text(data))
		})
}
