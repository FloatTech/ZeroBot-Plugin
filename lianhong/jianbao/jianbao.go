// Package jianbao 每日60秒读懂世界
package jianbao

import (
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("jianbao", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "查看每日60秒读懂世界简报\n" +
			"- 简报\n",
	})
	engine.OnFullMatch("简报").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://bh.ayud.top/mrjb")
			if err != nil {
				ctx.SendChain(message.Text("获取简报失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
}
