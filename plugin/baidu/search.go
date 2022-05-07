// Package baidu 百度一下
package baidu

import (
	"net/url"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

func init() {
	control.Register("baidu", &control.Options{
		DisableOnDefault: false,
		Help: "baidu\n" +
			"- 百度下[xxx]",
	}).OnPrefix("百度下").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				ctx.SendChain(message.Text("https://buhuibaidu.me/?s=" + url.QueryEscape(txt)))
			}
		})
}
