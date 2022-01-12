// Package acgimage 随机图片与AI点评
package acgimage

import (
	"net/url"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/classify"
	"github.com/FloatTech/AnimeAPI/picture"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
)

const (
	lolipxy   = "https://sayuri.fumiama.top/dice?class=0&loli=true&r18=true"
	apihead   = "https://sayuri.fumiama.top/img?path="
	apiheadv6 = "http://aikae.v6.army:62002/img?arg=get&name="
)

var (
	// r18有一定保护，一般不会发出图片
	randapi = "&loli=true&r18=true"
	msgof   = make(map[int64]int64)
	block   = false
	limit   = rate.NewManager(time.Minute, 5)
)

func init() { // 插件主体
	engine := control.Register("acgimage", &control.Options{
		DisableOnDefault: false,
		Help: "随机图片与AI点评\n" +
			"- 随机图片(评级大于6的图将私发)\n" +
			"- 直接随机(无r18检测，务必小心，仅管理可用)\n" +
			"- 设置随机图片网址[url]\n" +
			"- 太涩了(撤回最近发的图)\n" +
			"- 评价图片(发送一张图片让bot评分)",
	})
	engine.OnRegex(`^设置随机图片网址(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			if !strings.HasPrefix(url, "http") {
				ctx.SendChain(message.Text("URL非法!"))
			} else {
				randapi = url
				ctx.SendChain(message.Text("设置好啦"))
			}
		})
	// 有保护的随机图片
	engine.OnFullMatch("随机图片", zero.OnlyGroup).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if limit.Load(ctx.Event.UserID).Acquire() {
				class, dhash, comment, _ := classify.Classify(randapi, true)
				replyClass(ctx, class, dhash, comment, false)
				return
			}
			ctx.SendChain(message.Text("你太快啦!"))
		})
	// 直接随机图片，无r18保护，后果自负。如果出r18图可尽快通过发送"太涩了"撤回
	engine.OnFullMatch("直接随机", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if block {
				ctx.SendChain(message.Text("请稍后再试哦"))
			} else if randapi != "" {
				block = true
				var url string
				if randapi[0] == '&' {
					url = lolipxy
				} else {
					url = randapi
				}
				setLastMsg(ctx.Event.GroupID, ctx.SendChain(message.Image(url).Add("cache", "0")))
				block = false
			}
		})
	// 撤回最后的直接随机图片
	engine.OnFullMatch("太涩了").SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			msg, ok := msgof[ctx.Event.GroupID]
			if ok {
				ctx.DeleteMessage(msg)
				delete(msgof, ctx.Event.GroupID)
			}
		})
	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"评价图片"}, zero.OnlyGroup, picture.CmdMatch, picture.MustGiven).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			for _, url := range ctx.State["image_url"].([]string) {
				class, dhash, comment, _ := classify.Classify(url, true)
				replyClass(ctx, class, dhash, comment, true)
				break
			}
		})
	engine.OnRegex(`^给你点提示哦：(.*)$`, zero.OnlyPrivate).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			dhash := ctx.State["regex_matched"].([]string)[1]
			if len(dhash) == 5*3 {
				if web.IsSupportIPv6 {
					ctx.SendChain(message.Image(apiheadv6 + dhash + ".webp"))
				} else {
					ctx.SendChain(message.Image(apihead + dhash))
				}
			}
		})
}

func setLastMsg(id int64, msg int64) {
	msgof[id] = msg
}

func replyClass(ctx *zero.Ctx, class int, dhash string, comment string, isupload bool) {
	b14, err := url.QueryUnescape(dhash)
	if err != nil {
		return
	}

	var img message.MessageSegment
	if web.IsSupportIPv6 {
		img = message.Image(apiheadv6 + dhash + ".webp")
	} else {
		img = message.Image(apihead + dhash)
	}

	if class > 5 {
		if dhash != "" && !isupload {
			ctx.SendChain(message.Text(comment + "\n给你点提示哦：" + b14))
			ctx.Event.GroupID = 0
			ctx.SendChain(img)
			return
		}
		ctx.SendChain(message.Text(comment))
		return
	}
	if isupload {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(comment))
		return
	}
	ctx.SendChain(img, message.Text(comment))
}
