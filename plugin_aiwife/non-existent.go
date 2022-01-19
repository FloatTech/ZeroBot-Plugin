// Package aiwife 随机老婆
package aiwife

import (
	"fmt"
	"math/rand"
	"time"

	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	bed = "https://www.thiswaifudoesnotexist.net/example-%d.jpg"
)

func init() { // 插件主体
	// TODO: 1.17 特性暂不增加
	// rand.Seed(time.Now().UnixMicro())
	rand.Seed(time.Now().UnixNano())
	control.Register("aiwife", order.PrioAIWife, &control.Options{
		DisableOnDefault: false,
		Help: "AIWife\n" +
			"- waifu | 随机waifu",
	}).OnFullMatchGroup([]string{"waifu", "随机waifu"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := rand.Intn(100000) + 1
			ctx.SendChain(message.At(ctx.Event.UserID), message.Image(fmt.Sprintf(bed, miku)))
		})
}
