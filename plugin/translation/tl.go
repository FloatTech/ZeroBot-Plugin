// Package translation 翻译
package translation

import (
	"github.com/FloatTech/AnimeAPI/tl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "单词翻译",
		Help:             ">TL [好|good]",
	}).OnRegex(`^>TL\s(-.{1,10}? )?(.*)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := []string{ctx.State["regex_matched"].([]string)[2]}
			data, err := tl.Translate(msg[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", data))
				return
			}
			ctx.SendChain(message.Text(data))
		})
}
