// Package curse 骂人插件(求骂,自卫)
package curse

import (
	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"time"
)

const (
	prio     = 10
	minLevel = "min"
	maxLevel = "max"
)

var (
	engine = control.Register("curse", &control.Options{
		DisableOnDefault: false,
		Help:             "骂人(求骂,自卫)\n- 骂我\n- 大力骂我\n- @bot 他妈|公交车|你妈|操|屎|去死|快死|日|逼|尼玛|艾滋|癌症|有病|戴套|啊对对对|烦你",
	})
	limit = rate.NewManager(time.Minute, 30)
)

func init() {
	engine.OnFullMatch("骂我").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			return
		}
		process.SleepAbout1sTo2s()
		text := getRandomCurseByLevel(minLevel).Text
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})

	engine.OnFullMatch("大力骂我").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			return
		}
		process.SleepAbout1sTo2s()
		text := getRandomCurseByLevel(maxLevel).Text
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})
	engine.OnKeywordGroup([]string{"他妈", "公交车", "你妈", "操", "屎", "去死", "快死", "日", "逼", "尼玛", "艾滋", "癌症", "有病", "戴套", "啊对对对", "烦你"}, zero.OnlyToMe).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			text := getRandomCurseByLevel(maxLevel).Text
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
		})
}
