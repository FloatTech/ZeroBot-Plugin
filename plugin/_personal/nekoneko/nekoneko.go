// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.Register("neko", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "云吸猫",
		Help:             "   - neko",
	}).OnKeyword("neko").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://api.thecatapi.com/v1/images/search")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image(gjson.ParseBytes(data).Get("0.url").String()))
		})
}
