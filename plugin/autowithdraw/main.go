// Package autowithdraw 触发者撤回时也自动撤回
package autowithdraw

import (
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "触发者撤回时也自动撤回",
		Help:             "- 撤回一条消息\n",
	}).OnNotice(func(ctx *zero.Ctx) bool {
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
