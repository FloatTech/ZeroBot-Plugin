package recall

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("recall", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "Reply a message with 'recall' or 'Recall' to recall it.\n" +
			"- Recall",
	})

	engine.OnRegex(`^\[CQ:reply,id=(\d+|-\d+).*(?:\](\s{0,1}))(?:Recall|recall)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regex := ctx.State["regex_matched"].([]string)
			curmsgid := ctx.Event.MessageID.(int64)
			ctx.DeleteMessage(message.NewMessageIDFromString(regex[1]))
			ctx.DeleteMessage(message.NewMessageIDFromInteger(curmsgid))
		})
}
