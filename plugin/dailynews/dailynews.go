// Package dailynews 今日早报
package dailynews

import (
	"encoding/base64"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const api = "https://uapis.cn/api/v1/daily/news-image"

func init() {
	engine := control.Register("dailynews", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "今日早报",
		Help:             "- 今日早报",
		PrivateDataFolder: "dailynews",
	})

	engine.OnFullMatch(`今日早报`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData(api)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + base64.StdEncoding.EncodeToString(data)))
		})
}
