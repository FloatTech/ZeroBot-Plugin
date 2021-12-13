// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	log "github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bed = "https://codechina.csdn.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"
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
			userId := ctx.Event.UserID
			today, err := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
			if err != nil {
				log.Errorln("string转化为int64格式有问题:", err)
			}
			seed := userId + today
			rand.Seed(seed)
			miku := rand.Intn(100) + 1
			ctx.SendChain(
				message.At(userId),
				message.Image(fmt.Sprintf(bed, miku, 0)),
				message.Image(fmt.Sprintf(bed, miku, 1)),
			)
		})
	engine.OnFullMatchGroup([]string{"解签"}).SetPriority(10).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userId := ctx.Event.UserID
			today, err := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
			if err != nil {
				log.Errorln("string转化为int64格式有问题:", err)
			}
			seed := userId + today
			rand.Seed(seed)
			miku := rand.Intn(100) + 1
			s := getSignatureById(miku)
			ctx.SendChain(
				message.At(userId),
				message.Text(s.Text),
			)
		})
}
