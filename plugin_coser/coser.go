// Package coser images
package coser

import (
	"time"

	"github.com/tidwall/gjson"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"
)

var (
	engine = control.Register("coser", &control.Options{
		DisableOnDefault: false,
		Help:             "三次元小姐姐\n- coser\n",
	})
	prio     = 20
	ua       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36"
	coserURL = "http://ovooa.com/API/cosplay/api.php"
	limit    = rate.NewManager(time.Minute, 5)
)

func init() {
	engine.OnFullMatch("coser", zero.OnlyGroup).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.GroupID).Acquire() {
				ctx.SendChain(message.Text("请稍后重试0x0..."))
				return
			}
			ctx.SendChain(message.Text("少女祈祷中......"))
			data, err := web.ReqWith(coserURL, "GET", "", ua)
			if err != nil {
				log.Println("err为:", err)
			}
			var m message.Message
			text := gjson.Get(helper.BytesToString(data), "data.Title").String()
			m = append(m,
				message.CustomNode(
					zero.BotConfig.NickName[0],
					ctx.Event.SelfID,
					text,
				))
			gjson.Get(helper.BytesToString(data), "data.data").ForEach(func(_, value gjson.Result) bool {
				imgcq := `[CQ:image,file=` + value.String() + `]`
				m = append(m,
					message.CustomNode(
						zero.BotConfig.NickName[0],
						ctx.Event.SelfID,
						imgcq),
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
