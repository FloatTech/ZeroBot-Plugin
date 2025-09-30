// Package nsfwauto 图片合规性审查的自动版本
package nsfwauto

import (
	"github.com/FloatTech/AnimeAPI/nsfw"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const hso = "https://gchat.qpic.cn/gchatpic_new//--4234EDEC5F147A4C319A41149D7E0EA9/0"

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "nsfw图片自动识别",
		Help:             "- 当图片属于非 neutral 类别时自动发送评价",
	}).OnMessage(zero.HasPicture).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["image_url"].([]string)
			if len(url) > 0 {
				process.SleepAbout1sTo2s()
				p, err := nsfw.Classify(url[0])
				if err != nil {
					return
				}
				process.SleepAbout1sTo2s()
				autojudge(ctx, p)
			}
		})
}

func autojudge(ctx *zero.Ctx, p *nsfw.Picture) {
	if p.Neutral > 0.3 {
		return
	}
	c := ""
	if p.Drawings > 0.3 {
		c = "二次元"
	} else {
		c = "三次元"
	}
	i := 0
	if p.Hentai > 0.3 {
		c += " hentai"
		i++
	}
	if p.Porn > 0.3 {
		c += " porn"
		i++
	}
	if p.Sexy > 0.3 {
		c += " hso"
		i++
	}
	if i > 0 {
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(c, "\n"), message.Image(hso)))
	}
}
