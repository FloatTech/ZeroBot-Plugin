// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
)

const bed = "https://gitea.seku.su/fumiama/senso-ji-omikuji/raw/branch/main/"

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "浅草寺求签",
		Help:             "- 求签 | 占卜\n- 解签",
		PublicDataFolder: "Omikuji",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnFullMatchGroup([]string{"求签", "占卜"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			i := fcext.RandSenderPerDayN(ctx.Event.UserID, 100) + 1
			img0, err := engine.GetCustomLazyData(bed, fmt.Sprintf("%d_%d.jpg", i, 0))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			img1, err := engine.GetCustomLazyData(bed, fmt.Sprintf("%d_%d.jpg", i, 1))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.ImageBytes(img0),
				message.ImageBytes(img1),
			)
		})
	engine.OnFullMatch("解签", fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			db.DBPath = engine.DataFolder() + "kuji.db"
			_, err := engine.GetLazyData("kuji.db", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = db.Open(time.Hour)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = db.Create("kuji", &kuji{})
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			n, err := db.Count("kuji")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			logrus.Infof("[kuji]读取%d条签文", n)
			return true
		},
	)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			kujiBytes, err := text.RenderToBase64(
				getKujiByBango(
					uint8(fcext.RandSenderPerDayN(ctx.Event.UserID, 100)+1),
				),
				text.FontFile, 400, 20,
			)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.At(ctx.Event.UserID), message.Image("base64://"+helper.BytesToString(kujiBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
