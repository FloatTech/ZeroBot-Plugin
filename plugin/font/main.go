// Package font 渲染任意文字到图片
package font

import (
	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.Register("font", &control.Options{
		DisableOnDefault: false,
		Help:             "渲染任意文字到图片\n- (用[终末体|终末变体|紫罗兰体|樱酥体|Consolas体|苹方体])渲染文字xxx",
	}).OnRegex(`^(用.+)?渲染文字([\s\S]+)$`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		fnt := ctx.State["regex_matched"].([]string)[1]
		txt := ctx.State["regex_matched"].([]string)[2]
		switch fnt {
		case "用终末体":
			fnt = text.SyumatuFontFile
		case "用终末变体":
			fnt = text.NisiFontFile
		case "用紫罗兰体":
			fnt = text.VioletEvergardenFontFile
		case "用樱酥体":
			fnt = text.SakuraFontFile
		case "用Consolas体":
			fnt = text.ConsolasFontFile
		case "用苹方体":
			fallthrough
		default:
			fnt = text.FontFile
		}
		b, err := text.RenderToBase64(txt, fnt, 400, 20)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Image("base64://" + binary.BytesToString(b)))
	})
}
