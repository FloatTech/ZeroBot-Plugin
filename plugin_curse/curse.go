// Package curse 骂人插件(求骂,自卫)
package curse

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	minLevel = "min"
	maxLevel = "max"
)

func init() {
	engine := control.Register("curse", order.PrioCurse, &control.Options{
		DisableOnDefault: true,
		Help:             "骂人(求骂,自卫)\n- 骂我\n- 大力骂我",
	})

	engine.OnFullMatch("骂我").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		text := getRandomCurseByLevel(minLevel).Text
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})

	engine.OnFullMatch("大力骂我").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		text := getRandomCurseByLevel(maxLevel).Text
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})

	engine.OnKeywordGroup([]string{"他妈", "公交车", "你妈", "操", "屎", "去死", "快死", "我日", "逼", "尼玛", "艾滋", "癌症", "有病", "烦你", "你爹", "屮", "cnm"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := getRandomCurseByLevel(maxLevel).Text
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
		})
}
