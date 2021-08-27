// Package aiwife 随机老婆
package aiwife

import (
	"fmt"
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bed = "https://www.thiswaifudoesnotexist.net/example-%d.jpg"
)

func init() { // 插件主体
	rand.Seed(time.Now().UnixMicro())
	zero.OnFullMatchGroup([]string{"waifu", "随机waifu"}).SetPriority(10).
		Handle(func(ctx *zero.Ctx) {
			miku := rand.Intn(100000) + 1
			ctx.SendChain(message.At(ctx.Event.UserID), message.Image(fmt.Sprintf(bed, miku)))
		})
}
