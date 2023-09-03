// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "摸鱼人日历",
		Help: "- /启用 moyucalendar\n" +
			"- /禁用 moyucalendar\n" +
			"- 记录在\"30 8 * * *\"触发的指令\n" +
			"   - 摸鱼人日历",
	}).OnFullMatch("摸鱼人日历").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://api.vvhan.com/api/moyu")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
}
