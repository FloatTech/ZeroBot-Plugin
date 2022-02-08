// Package jandan 煎蛋网无聊图
package jandan

import (
	"github.com/FloatTech/ZeroBot-Plugin/order"
	"github.com/FloatTech/zbputils/control"
	log "github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	jandanPictureURL = "http://jandan.net/pic"
)

func init() {
	engine := control.Register("jandan", order.PrioJandan, &control.Options{
		DisableOnDefault: false,
		Help:             "煎蛋网无聊图\n- 来份屌图\n- 更新屌图\n",
	})

	engine.OnFullMatch("来份屌图").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			p, err := getRandomPicture()
			if err != nil {
				log.Errorln("[jandan]:", err)
				ctx.SendChain(message.Text("数据库更新中"))
			}
			ctx.SendChain(message.Image(p.PictureURL))
		})

	engine.OnFullMatch("更新屌图", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女更新中......")
			travelWebpage()
		})
}
