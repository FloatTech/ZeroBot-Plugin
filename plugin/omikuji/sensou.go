// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

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
			miku, err := bangoToday(ctx.Event.UserID)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Image(fmt.Sprintf(bed, miku, 0)),
				message.Image(fmt.Sprintf(bed, miku, 1)),
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
			log.Printf("[kuji]读取%d条签文", n)
			return true
		},
	)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			bg, err := bangoToday(ctx.Event.UserID)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			kujiBytes, err := text.RenderToBase64(getKujiByBango(bg), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if id := ctx.SendChain(message.At(ctx.Event.UserID), message.Image("base64://"+helper.BytesToString(kujiBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
		})
}

func bangoToday(uid int64) (uint8, error) {
	today, err := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
	if err != nil {
		return 0, err
	}
	seed := uid + today
	r := rand.New(rand.NewSource(seed))
	return uint8(r.Intn(100) + 1), nil
}
