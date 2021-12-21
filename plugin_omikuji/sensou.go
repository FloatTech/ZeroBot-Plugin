// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bed = "https://gitcode.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"
)

var (
	engine = control.Register("omikuji", &control.Options{
		DisableOnDefault: false,
		Help: "浅草寺求签\n" +
			"- 求签|占卜\n- 解签",
	})
)

func init() { // 插件主体
	engine.OnFullMatchGroup([]string{"求签", "占卜"}).SetPriority(10).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := bangoToday(ctx.Event.UserID)
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Image(fmt.Sprintf(bed, miku, 0)),
				message.Image(fmt.Sprintf(bed, miku, 1)),
			)
		})
	engine.OnFullMatchGroup([]string{"解签"}).SetPriority(10).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text(getKujiByBango(bangoToday(ctx.Event.UserID))),
			)
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
