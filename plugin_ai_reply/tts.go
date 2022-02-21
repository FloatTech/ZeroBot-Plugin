package aireply

import (
	"github.com/pkumza/numcn"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"regexp"
	"strconv"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/tts/mockingbird"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"

	"github.com/FloatTech/zbputils/control/order"
)

var (
	reNumber = "(\\-|\\+)?\\d+(\\.\\d+)?"
)

func init() {
	control.Register("mockingbird", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "拟声鸟\n- @Bot 任意文本(任意一句话回复)",
	}).OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			ctx.SendChain(message.Record(mockingbird.NewMockingBirdTTS(1).Speak(ctx.Event.UserID, func() string {
				reply := r.TalkPlain(msg, zero.BotConfig.NickName[0])
				re := regexp.MustCompile(reNumber)
				reply = re.ReplaceAllStringFunc(reply, func(s string) string {
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						log.Errorln("[mockingbird]:", err)
						return s
					}
					return numcn.EncodeFromFloat64(f)
				})
				log.Println("[mockingbird]:", reply)
				return reply
			})))
		})
}
