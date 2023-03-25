// Package wife 抽老婆
package wife

import (
	"encoding/json"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("wife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "- 抽老婆",
		Brief:            "从老婆库抽每日老婆",
		PublicDataFolder: "Wife",
	}).ApplySingle(ctxext.DefaultSingle)
	cards := []string{}
	uriprefix := "file:///" + file.BOTPATH + "/" + engine.DataFolder()
	engine.OnFullMatch("抽老婆", fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			data, err := engine.GetLazyData("wife.json", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &cards)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			logrus.Infof("[wife]加载%d个老婆", len(cards))
			return true
		},
	)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			card := cards[fcext.RandSenderPerDayN(ctx.Event.UserID, len(cards))]
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("今天的二次元老婆是~【", card, "】哒"),
				message.Image(uriprefix+"wives/"+card),
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("今天的二次元老婆是~【", card, "】哒\n【图片发送失败, 请联系维护者】"),
				)
			}
		})
}
