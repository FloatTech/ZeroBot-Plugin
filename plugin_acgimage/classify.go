// Package acgimage 随机图片与AI点评
package acgimage

import (
	"net/url"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/classify"
	"github.com/FloatTech/AnimeAPI/imgpool"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	lolipxy   = "https://sayuri.fumiama.top/dice?class=0&loli=true&r18=true"
	apihead   = "https://sayuri.fumiama.top/img?path="
	apiheadv6 = "http://aikae.v6.army:62002/img?arg=get&name="
)

var (
	// r18有一定保护，一般不会发出图片
	randapi = "&loli=true&r18=true"
	msgof   = make(map[int64]message.MessageID)
	block   = false
	limit   = rate.NewManager(time.Minute, 5)
)

func init() { // 插件主体
	engine := control.Register("acgimage", order.PrioACGImage, &control.Options{
		DisableOnDefault: false,
		Help: "随机图片与AI点评\n" +
			"- 随机图片(评级大于6的图将私发)\n" +
			"- 直接随机(无r18检测，务必小心，仅管理可用)\n" +
			"- 设置随机图片网址[url]\n" +
			"- 太涩了(撤回最近发的图)\n" +
			"- 评价图片(发送一张图片让bot评分)",
	})
	engine.OnRegex(`^设置随机图片网址(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
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
	engine.OnFullMatch("随机图片", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if limit.Load(ctx.Event.UserID).Acquire() {
				class, dhash, comment, _ := classify.Classify(randapi, true)
				replyClass(ctx, class, dhash, comment, false)
				return
			}
			ctx.SendChain(message.Text("你太快啦!"))
		})
	// 直接随机图片，无r18保护，后果自负。如果出r18图可尽快通过发送"太涩了"撤回
	engine.OnFullMatch("直接随机", ctxext.UserOrGrpAdmin).SetBlock(true).
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
	engine.OnFullMatch("太涩了", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			msg, ok := msgof[ctx.Event.GroupID]
			if ok {
				ctx.DeleteMessage(msg)
				delete(msgof, ctx.Event.GroupID)
			}
		})
	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"评价图片"}, zero.OnlyPublic, ctxext.CmdMatch, ctxext.MustGiven).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			for _, url := range ctx.State["image_url"].([]string) {
				class, dhash, comment, _ := classify.Classify(url, true)
				replyClass(ctx, class, dhash, comment, true)
				break
			}
		})
	engine.OnRegex(`^给你点提示哦：(.*)$`, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			dhash := ctx.State["regex_matched"].([]string)[1]
			if len(dhash) == 5*3 {
				var u string
				if web.IsSupportIPv6 {
					u = apiheadv6 + dhash + ".webp"
				} else {
					u = apihead + dhash
				}

				m, hassent, err := imgpool.NewImage(ctxext.Send(ctx), ctxext.GetMessage(ctx), dhash, u)
				if err == nil && !hassent {
					ctx.SendChain(message.Image(m.String()))
				}
			}
		})
}

func setLastMsg(id int64, msg message.MessageID) {
	msgof[id] = msg
}

func replyClass(ctx *zero.Ctx, class int, dhash string, comment string, isupload bool) {
	if isupload {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(comment))
		return
	}

	b14, err := url.QueryUnescape(dhash)
	if err != nil {
		return
	}

	var u string
	if web.IsSupportIPv6 {
		u = apiheadv6 + dhash + ".webp"
	} else {
		u = apihead + dhash
	}

	var send ctxext.NoCtxSendMsg
	if class > 5 {
		send = ctxext.SendTo(ctx, ctx.Event.UserID)
		if dhash != "" {
			ctx.SendChain(message.Text(comment + "\n给你点提示哦：" + b14))
		} else {
			ctx.SendChain(message.Text(comment))
		}
	} else {
		send = func(msg interface{}) int64 {
			return ctx.Send(append(msg.(message.Message), message.Text(comment))).ID()
		}
	}

	m, hassent, err := imgpool.NewImage(send, ctxext.GetMessage(ctx), b14, u)
	if err == nil && !hassent {
		send(message.Message{message.Image(m.String())})
	}
}
