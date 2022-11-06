// Package font 渲染任意文字到图片
package font

import (
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.Register("font", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "渲染任意文字到图片",
		Help:             "- (用[字体])渲染文字xxx\n可选字体: [终末体|终末变体|紫罗兰体|樱酥体|Consolas体|苹方体]",
	}).OnRegex(`^(用.+)?渲染(([0-9]+)?宽度([0-9]+)?大小)?文字([\s\S]+)$`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		fnt := ctx.State["regex_matched"].([]string)[1]
		wide := ctx.State["regex_matched"].([]string)[3]
		size := ctx.State["regex_matched"].([]string)[4]
		txt := ctx.State["regex_matched"].([]string)[5]
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
		if wide == "" {
			wide = "400"
		}
		if size == "" {
			size = "20"
		}
		widen := math.Str2Int64(wide)
		sizen := math.Str2Int64(size)
		b, err := text.RenderToBase64(txt, fnt, int(widen), int(sizen))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(b))); id.ID() == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})
}
