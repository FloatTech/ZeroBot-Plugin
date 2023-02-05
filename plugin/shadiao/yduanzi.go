package shadiao

import (
	"strings"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/zbputils/ctxext"
)

func init() {
	engine.OnFullMatch("讲个段子").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.RequestDataWith(web.NewDefaultClient(), yduanziURL, "POST", yduanziReferer, ua, nil)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		text := gjson.Get(helper.BytesToString(data), "duanzi").String()
		text = strings.ReplaceAll(text, "<br>", "\n")
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})
}
