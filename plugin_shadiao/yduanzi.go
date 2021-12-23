package shadiao

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"strings"
)

func init() {
	engine.OnFullMatch("讲个段子").SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			return
		}
		data, err := web.ReqWith(yduanziURL, "POST", yduanziReferer, ua)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		text := gjson.Get(helper.BytesToString(data), "duanzi").String()
		text = strings.Replace(text, "<br>", "\n", -1)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})
}
