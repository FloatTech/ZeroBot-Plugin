// Package baidu 百度一下
package baidu

import (
	"net/url"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

func init() {
	control.Register("baidu", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "不会百度吗",
		Help:             "- 百度下[xxx]",
	}).OnPrefix("百度下").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.State["args"].(string)
			if txt != "" {
				ctx.SendChain(message.Text("https://buhuibaidu.me/?s=" + url.QueryEscape(txt)))
			}
		})
}
