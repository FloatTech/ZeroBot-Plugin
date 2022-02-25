// Package nsfw 图片合规性审查
package nsfw

import (
	"github.com/FloatTech/AnimeAPI/nsfw"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

func init() {
	engine := control.Register("nsfw", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "nsfw图片识别\n- nsfw打分[图片]",
	}).ApplySingle(ctxext.DefaultSingle)
	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"nsfw打分"}, zero.OnlyGroup, ctxext.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["image_url"].([]string)
			if len(url) > 0 {
				ctx.SendChain(message.Text("少女祈祷中..."))
				p, err := nsfw.Classify(url...)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(judge(p[0]))))
			}
		})
	en := control.Register("nsfwauto", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help:             "nsfw图片自动识别\n- 当图片属于非 neutral 类别时自动发送评价",
	})
	en.OnMessage(ctxext.IsPicExists).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["image_url"].([]string)
			if len(url) > 0 {
				process.SleepAbout1sTo2s()
				p, err := nsfw.Classify(url...)
				if err != nil {
					return
				}
				process.SleepAbout1sTo2s()
				autojudge(ctx, p[0])
			}
		})
}

func judge(p nsfw.Picture) string {
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

func autojudge(ctx *zero.Ctx, p nsfw.Picture) {
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
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(c)))
	}
}
