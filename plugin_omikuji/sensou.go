// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/text"

	"github.com/FloatTech/zbputils/control/order"
)

const bed = "https://gitcode.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"

func init() { // 插件主体
	engine := control.Register("omikuji", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "浅草寺求签\n" +
			"- 求签 | 占卜\n- 解签",
		PublicDataFolder: "Omikuji",
	}).ApplySingle(ctxext.DefaultSingle)

	go func() {
		dbpath := engine.DataFolder()
		db.DBPath = dbpath + "kuji.db"
		defer order.DoneOnExit()()
		_, _ = file.GetLazyData(db.DBPath, false, true)
		err := db.Create("kuji", &kuji{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("kuji")
		if err != nil {
			panic(err)
		}
		log.Printf("[kuji]读取%d条签文", n)
	}()

	engine.OnFullMatchGroup([]string{"求签", "占卜"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := bangoToday(ctx.Event.UserID)
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Image(fmt.Sprintf(bed, miku, 0)),
				message.Image(fmt.Sprintf(bed, miku, 1)),
			)
		})
	engine.OnFullMatchGroup([]string{"解签"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			kujiBytes, err := text.RenderToBase64(getKujiByBango(bangoToday(ctx.Event.UserID)), text.FontFile, 400, 20)
			if err != nil {
				log.Errorln("[omikuji]:", err)
			}
			if id := ctx.SendChain(message.At(ctx.Event.UserID), message.Image("base64://"+helper.BytesToString(kujiBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}

func bangoToday(uid int64) uint8 {
	today, err := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
	if err != nil {
		log.Errorln("string转化为int64格式有问题:", err)
	}
	seed := uid + today
	r := rand.New(rand.NewSource(seed))
	return uint8(r.Intn(100) + 1)
}
