// Package delreply 撤回消息
package delreply

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 插件主体
func init() {
	engine := control.Register("delreply", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "回复消息\"撤回\"以撤回消息",
	})
	engine.OnRegex(`^\[CQ:reply,id=(-?[0-9]+)\].*`, zero.AdminPermission, zero.KeywordRule("撤回")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.DeleteMessage(message.NewMessageIDFromString(ctx.State["regex_matched"].([]string)[1]))
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
		})
}
