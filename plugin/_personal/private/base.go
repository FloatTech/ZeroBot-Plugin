// Package base 基础指令
package base

import (
	"os"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	serviceName = "base"
	botQQ       = 1015464740 // 机器人QQ
)

var engine = control.Register(serviceName, &ctrl.Options[*zero.Ctx]{
	DisableOnDefault:  false,
	Brief:             "基础指令",
	Help:              "- @bot备份代码\n- @bot上传代码\n- @bot检查更新- @bot重启\ntips:检查更新后如果没有问题后需要重启才OK",
	PrivateDataFolder: "base",
	OnDisable: func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Text("宝↗生↘永↗梦↘！！！！"))
	},
})

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		ctx := zero.GetBot(botQQ)
		m, ok := control.Lookup(serviceName)
		if ok {
			gid := m.GetData(-zero.BotConfig.SuperUsers[0])
			if gid != 0 {
				ctx.SendGroupMessage(gid, message.Text("我回来了😊"))
			} else {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("我回来了😊"))
			}
		}
		err := m.SetData(-zero.BotConfig.SuperUsers[0], 0)
		if err != nil {
			ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(err))
		}
	}()
	// 重启
	zero.OnFullMatchGroup([]string{"重启", "洗手手"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m, ok := control.Lookup(serviceName)
			if ok {
				err := m.SetData(-zero.BotConfig.SuperUsers[0], ctx.Event.GroupID)
				if err == nil {
					ctx.SendChain(message.Text("好的"))
				} else {
					ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(err))
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
	zero.OnRegex(`^\[CQ:reply,id=(.*)].*`, zero.KeywordRule("多嘴", "撤回")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取消息id
			mid := ctx.State["regex_matched"].([]string)[1]
			// 撤回消息
			if ctx.Event.Message[1].Data["qq"] != "" {
				var nickname = zero.BotConfig.NickName[0]
				ctx.SendChain(message.Text("9494,要像", nickname, "一样乖乖的才行哟~"))
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
