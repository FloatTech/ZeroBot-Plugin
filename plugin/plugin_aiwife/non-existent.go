// Package aiwife 随机老婆
package aiwife

import (
	"fmt"
	"math/rand"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	bed = "https://www.thiswaifudoesnotexist.net/example-%d.jpg"
)

func init() { // 插件主体
	control.Register("aiwife", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "AIWife\n" +
			"- waifu | 随机waifu",
	}).ApplySingle(ctxext.DefaultSingle).OnFullMatchGroup([]string{"waifu", "随机waifu"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			miku := rand.Intn(100000) + 1
			ctx.SendChain(message.At(ctx.Event.UserID), message.Image(fmt.Sprintf(bed, miku)))
		})
}
