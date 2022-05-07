// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.Register("moyucalendar", &control.Options{
		DisableOnDefault: true,
		Help: "摸鱼人日历\n" +
			"- /启用 moyucalendar\n" +
			"- /禁用 moyucalendar\n" +
			"- 记录在\"30 8 * * *\"触发的指令\n" +
			"   - 摸鱼人日历",
	}).OnFullMatch("摸鱼人日历").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://api.vvhan.com/api/moyu")
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
}
