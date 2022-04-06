// Package inject 注入指令
package inject

import (
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	en := control.Register("inject", &control.Options{
		DisableOnDefault: false,
		Help: "注入指令\n" +
			"- run[CQ码]",
	})
	// 运行 CQ 码
	en.OnPrefix("run", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 可注入，权限为主人
			ctx.Send(message.UnescapeCQCodeText(ctx.State["args"].(string)))
		})
}
