// Package translation 翻译
package translation

import (
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"

	"github.com/FloatTech/zbputils/control/order"
)

func init() {
	control.Register("translation", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "翻译\n" +
			">TL 你好",
	}).OnRegex(`^>TL\s(-.{1,10}? )?(.*)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := []string{ctx.State["regex_matched"].([]string)[2]}
			data, err := web.GetData("https://api.cloolc.club/fanyi?data=" + msg[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			info := gjson.ParseBytes(data)
			repo := info.Get("data.0")
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text(repo.Get("value.0")))
		})
}
