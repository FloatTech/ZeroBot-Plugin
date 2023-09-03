// Package font 渲染任意文字到图片
package font

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"math/rand"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "渲染任意文字到图片",
		Help:             "- (用[字体])渲染(抖动)文字xxx\n可选字体: [终末体|终末变体|紫罗兰体|樱酥体|Consolas体|粗苹方体|未来荧黑体|Gugi体|八丸体|Impact体|猫啃体|苹方体]",
	}).OnRegex(`^(用.+)?渲染(抖动)?文字([\s\S]+)$`).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		fnt := ctx.State["regex_matched"].([]string)[1]
		txt := ctx.State["regex_matched"].([]string)[3]
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
		case "用粗苹方体":
			fnt = text.BoldFontFile
		case "用未来荧黑体":
			fnt = text.GlowSansFontFile
		case "用Gugi体":
			fnt = text.GugiRegularFontFile
		case "用八丸体":
			fnt = text.HachiMaruPopRegularFontFile
		case "用Impact体":
			fnt = text.ImpactFontFile
		case "用猫啃体":
			fnt = text.MaokenFontFile
		case "用苹方体":
			fallthrough
		default:
			fnt = text.FontFile
		}
		if ctx.State["regex_matched"].([]string)[2] == "" {
			b, err := text.RenderToBase64(txt, fnt, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(b)))
			return
		}
		nilx, nily := 1.0, 8.0
		s := []*image.NRGBA{}
		strlist := strings.Split(txt, "\n")
		data, err := file.GetLazyData(fnt, control.Md5File, true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 获得画布预计
		testcov := gg.NewContext(1, 1)
		if err = testcov.ParseFontFace(data, 30); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 取最长段
		txt = ""
		for _, v := range strlist {
			if len([]rune(v)) > len([]rune(txt)) {
				txt = v
			}
		}
		w, h := testcov.MeasureString(txt)
		for i := 0; i < 10; i++ {
			cov := gg.NewContext(int(w+float64(len([]rune(txt)))*nilx)+40, int(h+nily)*len(strlist)+30)
			cov.SetRGB(1, 1, 1)
			cov.Clear()
			if err = cov.ParseFontFace(data, 30); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			cov.SetColor(color.NRGBA{R: 0, G: 0, B: 0, A: 127})
			for k, v := range strlist {
				for kk, vv := range []rune(v) {
					x, y := cov.MeasureString(string([]rune(v)[:kk]))
					cov.DrawString(string(vv), x+float64(rand.Intn(5))+10+nilx, y+float64(rand.Intn(5))+15+float64(k)*(y+nily))
				}
			}
			s = append(s, imgfactory.Size(cov.Image(), 0, 0).Image())
		}
		var buf bytes.Buffer
		_ = gif.EncodeAll(&buf, imgfactory.MergeGif(5, s))
		ctx.SendChain(message.ImageBytes(buf.Bytes()))
	})
}
