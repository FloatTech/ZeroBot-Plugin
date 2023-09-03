// Package alipayvoice 支付宝到账语音
package alipayvoice

import (
	"fmt"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	alipayvoiceURL = "https://mm.cqu.cc/share/zhifubaodaozhang/mp3/%v.mp3"
)

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "支付宝到账语音",
		Help:              "- 支付宝到账 1",
		PrivateDataFolder: "alipayvoice",
	})

	// 开启
	engine.OnPrefix(`支付宝到账`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			ctx.SendChain(message.Record(fmt.Sprintf(alipayvoiceURL, strings.TrimSpace(args))))
		})
}
