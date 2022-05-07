// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
)

const bed = "https://gitcode.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"

func init() { // 插件主体
	engine := control.Register("omikuji", &control.Options{
		DisableOnDefault: false,
		Help: "浅草寺求签\n" +
			"- 求签 | 占卜\n- 解签",
		PublicDataFolder: "Omikuji",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnFullMatchGroup([]string{"求签", "占卜"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			i := ctxext.RandSenderPerDayN(ctx, 100) + 1
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Image(fmt.Sprintf(bed, i, 0)),
				message.Image(fmt.Sprintf(bed, i, 1)),
			)
		})
	engine.OnFullMatch("解签", ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			db.DBPath = engine.DataFolder() + "kuji.db"
			_, err := engine.GetLazyData("kuji.db", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = db.Create("kuji", &kuji{})
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			n, err := db.Count("kuji")
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			logrus.Infof("[kuji]读取%d条签文", n)
			return true
		},
	)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			kujiBytes, err := text.RenderToBase64(
				getKujiByBango(
					uint8(ctxext.RandSenderPerDayN(ctx, 100)+1),
				),
				text.FontFile, 400, 20,
			)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if id := ctx.SendChain(message.At(ctx.Event.UserID), message.Image("base64://"+helper.BytesToString(kujiBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
		})
}
