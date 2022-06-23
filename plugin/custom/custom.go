// Package custom 自定义插件
package custom

import (
	"os"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("custom", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "自定义插件集合\n" +
			" - /kill\n" +
			" - 来114514份涩图\n" +
			" - /发送公告",
	})
	engine.OnFullMatchGroup([]string{"pause", "restart", "/kill"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			os.Exit(0)
		})
	engine.OnRegex(`^来(.*)涩图`, zero.OnlyGroup, zero.KeywordRule("114514")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image("https://gchat.qpic.cn/gchatpic_new/1770747317/1049468946-3068097579-76A49478EFA68B4750B10B96917F7B58/0?term=3"))
		})
	engine.OnCommand("发送公告", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, ctx.CheckSession())
			recv, stop := next.Repeat()
			defer stop()
			ctx.SendChain(message.Text("请输入公告内容"))
			var step int
			var origin string
			for {
				select {
				case <-time.After(time.Second * 60):
					ctx.SendChain(message.Text("时间太久啦！不发了！"))
					return
				case c := <-recv:
					switch step {
					case 0:
						origin = c.Event.RawMessage
						ctx.SendChain(message.Text("请输入\"确定\"或者\"取消\"来决定是否发送此公告"))
						step++
					case 1:
						msg := c.Event.Message.ExtractPlainText()
						if msg != "确定" && msg != "取消" {
							ctx.SendChain(message.Text("请输入\"确定\"或者\"取消\"哟"))
							continue
						}
						if msg == "确定" {
							ctx.SendChain(message.Text("正在发送..."))
							zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
								for _, g := range ctx.GetGroupList().Array() {
									gid := g.Get("group_id").Int()
									ctx.SendGroupMessage(gid, origin)
								}
								return true
							})
							return
						}
						ctx.SendChain(message.Text("已经取消发送了哟~"))
						return
					}
				}
			}
		})
}
