// Package omikuji 浅草寺求签
package omikuji

import (
	"fmt"
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bed = "https://codechina.csdn.net/u011570312/senso-ji-omikuji/-/raw/main/%d_%d.jpg"
)

func init() { // 插件主体
	rand.Seed(time.Now().UnixMicro())
	zero.OnFullMatchGroup([]string{"求签", "运势", "占卜"}, zero.OnlyToMe).SetPriority(10).
		Handle(func(ctx *zero.Ctx) {
			miku := rand.Intn(100) + 1
			ctx.SendChain(message.At(ctx.Event.UserID), message.Image(fmt.Sprintf(bed, miku, 0)), message.Image(fmt.Sprintf(bed, miku, 1)))
		})
}
