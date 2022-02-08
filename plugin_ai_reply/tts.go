package aireply

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/mockingbird"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

func init() {
	control.Register("mockingbird", order.PrioMockingBird, &control.Options{
		DisableOnDefault: false,
		Help:             "拟声鸟\n- @Bot 任意文本(任意一句话回复)",
	}).OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			ctx.SendChain(message.Record(mockingbird.Speak(ctx.Event.UserID, func() string {
				return r.TalkPlain(msg, zero.BotConfig.NickName[0])
			})))
		})
}
