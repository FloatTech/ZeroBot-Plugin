// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

const (
	bed = "https://codechina.csdn.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"
)

func init() { // 插件主体
	rand.Seed(time.Now().UnixNano())
	control.Register("omikuji", &control.Options{
		DisableOnDefault: false,
		Help: "浅草寺求签\n" +
			"- 求签|占卜",
	}).OnFullMatchGroup([]string{"求签", "占卜"}).SetPriority(10).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := rand.Intn(100) + 1
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Image(fmt.Sprintf(bed, miku, 0)),
				message.Image(fmt.Sprintf(bed, miku, 1)),
			)
		})
}
