// Package vtbwife 抽vtb老婆
package vtbwife

import (
	"encoding/json"
	"strconv"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const url = "http://8.134.179.136/vtbwife?id="

func init() { // 插件主体
	engine := control.Register("vtbwife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "抽vtb老婆",
		Help:             "- 抽vtb(老婆)",
	})
	engine.OnRegex(`^抽(vtb|VTB)(老婆)?$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		body, err := web.GetData(url + strconv.Itoa(int(ctx.Event.UserID)))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		var r result
		err = json.Unmarshal(body, &r)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		txt := message.Text(
			"\n今天你的VTB老婆是: ", r.Name,
		)
		if id := ctx.SendChain(message.At(ctx.Event.UserID), txt, message.Image(r.Imgurl), message.Text(r.Message)); id.ID() == 0 {
			ctx.SendChain(message.At(ctx.Event.UserID), txt, message.Text("图片发送失败...\n"), message.Text(r.Message))
		}
	})
}

type result struct {
	Code    int    `json:"code"`
	Imgurl  string `json:"imgurl"`
	Name    string `json:"name"`
	Message string `json:"message"`
}
