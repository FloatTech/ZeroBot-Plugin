// Package base 基础指令
package base

import (
	"os"

	"github.com/FloatTech/floatbox/process"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		ctx := zero.GetBot(1015464740)
		m, ok := control.Lookup("yaner")
		if ok {
			gid := m.GetData(-2504407110)
			if gid != 0 {
				ctx.SendGroupMessage(gid, message.Text("我回来了😊"))
			} else {
				ctx.SendPrivateMessage(2504407110, message.Text("我回来了😊"))
			}
		}
		err := m.SetData(-2504407110, 0)
		if err != nil {
			ctx.SendPrivateMessage(2504407110, message.Text(err))
		}
	}()
	// 重启
	zero.OnFullMatchGroup([]string{"重启", "洗手手"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m, ok := control.Lookup("yaner")
			if ok {
				err := m.SetData(-2504407110, ctx.Event.GroupID)
				if err == nil {
					ctx.SendChain(message.Text("好的"))
				} else {
					ctx.SendPrivateMessage(2504407110, message.Text(err))
				}
			}
			os.Exit(0)
		})
	// 运行 CQ 码
	zero.OnPrefix("run", zero.SuperUserPermission).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			// 可注入，权限为主人
			ctx.Send(message.UnescapeCQCodeText(ctx.State["args"].(string)))
		})
	// 撤回最后的发言
	zero.OnRegex(`^\[CQ:reply,id=(.*)].*`, zero.KeywordRule("多嘴")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取消息id
			mid := ctx.State["regex_matched"].([]string)[1]
			// 撤回消息
			if ctx.Event.Message[1].Data["qq"] != "" {
				var nickname = zero.BotConfig.NickName[0]
				ctx.SendChain(message.Text("9494，要像", nickname, "一样乖乖的才行哟~"))
			} else {
				ctx.SendChain(message.Text("呜呜呜呜"))
			}
			ctx.DeleteMessage(message.NewMessageIDFromString(mid))
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
		})
	zero.OnNotice(func(ctx *zero.Ctx) bool {
		return ctx.Event.NoticeType == "group_recall" || ctx.Event.NoticeType == "friend_recall"
	}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		id, ok := ctx.Event.MessageID.(int64)
		if !ok {
			return
		}
		for _, msg := range zero.GetTriggeredMessages(message.NewMessageIDFromInteger(id)) {
			process.SleepAbout1sTo2s()
			ctx.DeleteMessage(msg)
		}
	})
}
