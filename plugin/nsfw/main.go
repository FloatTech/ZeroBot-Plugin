// Package nsfw 图片合规性审查
package nsfw

import (
	"github.com/FloatTech/AnimeAPI/nsfw"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "nsfw图片识别",
		Help:             "- nsfw打分[图片]",
	}).ApplySingle(ctxext.DefaultSingle)
	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"nsfw打分"}, zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["image_url"].([]string)
			if len(url) > 0 {
				ctx.SendChain(message.Text("少女祈祷中..."))
				p, err := nsfw.Classify(url[0])
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(judge(p))))
			}
		})
}

func judge(p *nsfw.Picture) string {
	if p.Neutral > 0.3 {
		return "普通哦"
	}
	c := ""
	if p.Drawings > 0.3 || p.Neutral < 0.3 {
		c = "二次元"
	} else {
		c = "三次元"
	}
	if p.Hentai > 0.3 {
		c += " hentai"
	}
	if p.Porn > 0.3 {
		c += " porn"
	}
	if p.Sexy > 0.3 {
		c += " hso"
	}
	return c
}
