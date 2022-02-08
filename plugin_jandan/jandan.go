// Package jandan 煎蛋网无聊图
package jandan

import (
	"math/rand"
	"time"

	"github.com/FloatTech/zbputils/control"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	jandanPictureURL = "http://jandan.net/pic"
)

var (
	pictureList []string
)

func init() {
	engine := control.Register("jandan", order.PrioJandan, &control.Options{
		DisableOnDefault: false,
		Help:             "煎蛋网无聊图\n- 来份屌图\n- 更新屌图\n",
	})

	engine.OnFullMatch("来份屌图").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			rand.Seed(time.Now().Unix())
			ctx.SendChain(message.Image(pictureList[rand.Intn(len(pictureList))]))
		})

	engine.OnFullMatch("更新屌图", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女更新中......")
			travelWebpage()
		})
}
