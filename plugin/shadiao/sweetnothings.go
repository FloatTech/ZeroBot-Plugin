package shadiao

import (
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/zbputils/ctxext"
)

func init() {
	engine.OnFullMatch("来碗绿茶").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.RequestDataWith(web.NewDefaultClient(), chayiURL, "GET", loveliveReferer, ua, nil)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		text := gjson.Get(helper.BytesToString(data), "returnObj.content").String()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})

	engine.OnFullMatch("渣我").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.RequestDataWith(web.NewDefaultClient(), ganhaiURL, "GET", loveliveReferer, ua, nil)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		text := gjson.Get(helper.BytesToString(data), "returnObj.content").String()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})
}
