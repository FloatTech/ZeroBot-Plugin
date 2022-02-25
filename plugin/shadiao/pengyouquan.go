package shadiao

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
)

func init() {
	engine.OnFullMatch("发个朋友圈").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.ReqWith(pyqURL, "GET", pyqReferer, ua)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(helper.BytesToString(data)))
	})
}
