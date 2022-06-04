// Package custom 自定义插件
package custom

import (
	"os"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("custom", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "自定义插件集合\n" +
			" - /kill@bot\n" +
			" - 来114514份涩图",
	})
	engine.OnFullMatchGroup([]string{"pause", "restart", "/kill"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			os.Exit(0)
		})
	engine.OnRegex(`^来(.*)涩图`, zero.OnlyGroup, zero.KeywordRule("114514")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image("https://gchat.qpic.cn/gchatpic_new/1770747317/1049468946-3068097579-76A49478EFA68B4750B10B96917F7B58/0?term=3"))
		})
}
