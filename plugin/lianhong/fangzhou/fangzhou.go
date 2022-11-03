// Package fangzhou 图片获取集合
package fangzhou

import (
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("fangzhou", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "全部方舟指令\n" +
			"-方舟十连\n",
	})
	engine.OnFullMatch("方舟十连").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://bh.ayud.top/fz")
			if err != nil {
				ctx.SendChain(message.Text("抽卡失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
}
