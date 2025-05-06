// Package kfccrazythursday 疯狂星期四
package kfccrazythursday

import (
	"encoding/json"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	crazyURL = "https://api.pearktrue.cn/api/kfc/"
)

type crazyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Text string `json:"text"`
}

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "疯狂星期四",
		Help:             "疯狂星期四\n",
	})
	engine.OnFullMatch("疯狂星期四").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(crazyURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		var resp crazyResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			ctx.SendChain(message.Text("JSON解析失败: ", err))
			return
		}

		if resp.Code != 200 {
			ctx.SendChain(message.Text("API返回错误: ", resp.Msg))
			return
		}

		ctx.SendChain(message.Text(resp.Text))
	})
}
