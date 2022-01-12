package aireply

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/mockingbird"
	control "github.com/FloatTech/zbputils/control"
)

const ttsprio = 250

func init() {
	limit := rate.NewManager(time.Second*10, 1)

	control.Register("mockingbird", &control.Options{
		DisableOnDefault: false,
		Help:             "拟声鸟\n- @Bot 任意文本(任意一句话回复)",
	}).OnMessage(zero.OnlyToMe, func(ctx *zero.Ctx) bool {
		return limit.Load(ctx.Event.UserID).Acquire()
	}).SetBlock(true).SetPriority(ttsprio).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			ctx.SendChain(mockingbird.Speak(ctx.Event.UserID, func() string {
				return r.TalkPlain(msg)
			}))
		})
}
