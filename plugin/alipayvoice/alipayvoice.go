// Package alipayvoice 支付宝到账语音
package alipayvoice

import (
	b64 "encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	alipayvoiceURL = "https://mm.cqu.cc/share/zhifubaodaozhang/mp3/%v.mp3"
)

func init() { // 插件主体
	engine := control.Register("alipayvoice", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "支付宝到账语音",
		Help:              "- 支付宝到账 1",
		PrivateDataFolder: "alipayvoice",
	})

	// 开启
	engine.OnPrefix(`支付宝到账`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if moneyCount, err := strconv.ParseFloat(strings.TrimSpace(args), 64); err == nil && moneyCount > 0 {
				if data, err := web.GetData(fmt.Sprintf(alipayvoiceURL, moneyCount)); err == nil {
					ctx.SendChain(message.Record("base64://" + b64.StdEncoding.EncodeToString(data)))
				}
			}
		})
}
