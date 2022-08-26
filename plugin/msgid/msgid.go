package msgid

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("msgid", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "获取消息ID, 可以用于CQ码等.\n" +
			"- ID",
	})

	engine.OnRegex(`(CQ:reply,id=)(\d+|-\d+).*(?:\]| )ID$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			list := ctx.State["regex_matched"].([]string)
			ctx.SendChain(message.Text(list[2]))

		})

	engine.OnFullMatch("ID").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(`请对希望获取消息ID的目标消息回复"ID"以获取消息ID.`))
		})

}
