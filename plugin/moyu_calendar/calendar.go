// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
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
			data, err := web.RequestDataWith(web.NewDefaultClient(), "https://api.vvhan.com/api/moyu?type=json", "GET", "", "")
			if err != nil {
				return
			}
			picURL := gjson.Get(string(data), "url").String()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Image(picURL))
		})
}
