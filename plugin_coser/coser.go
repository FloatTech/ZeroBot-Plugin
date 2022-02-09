// Package coser images
package coser

import (
	"github.com/tidwall/gjson"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"

	"github.com/FloatTech/zbputils/control/order"
)

var (
	ua       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36"
	coserURL = "http://ovooa.com/API/cosplay/api.php"
)

func init() {
	control.Register("coser", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "三次元小姐姐\n- coser",
	}).ApplySingle(ctxext.DefaultSingle).OnFullMatch("coser", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中......"))
			data, err := web.ReqWith(coserURL, "GET", "", ua)
			if err != nil {
				log.Println("err为:", err)
			}
			var m message.Message
			text := gjson.Get(helper.BytesToString(data), "data.Title").String()
			m = append(m,
				message.CustomNode(
					ctx.Event.Sender.NickName,
					ctx.Event.UserID,
					text,
				))
			gjson.Get(helper.BytesToString(data), "data.data").ForEach(func(_, value gjson.Result) bool {
				m = append(m,
					message.CustomNode(
						ctx.Event.Sender.NickName,
						ctx.Event.UserID,
						[]message.MessageSegment{
							message.Image(value.String()),
						}),
				)
				return true
			})

			if id := ctx.SendGroupForwardMessage(
				ctx.Event.GroupID,
				m).Get("message_id").Int(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
