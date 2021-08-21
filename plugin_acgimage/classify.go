// Package acgimage 随机图片与AI点评
package acgimage

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/AnimeAPI/classify"
	"github.com/FloatTech/AnimeAPI/picture"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	lolipxy = "https://sayuri.fumiama.top/dice?class=0&loli=true&r18=true"
	apihead = "https://sayuri.fumiama.top/img?path="
)

var (
	botpath, _ = os.Getwd()
	datapath   = botpath + "/data/acgimage/"
	cacheuri   = "file:///" + datapath + "cache"
	// r18有一定保护，一般不会发出图片
	randapi = "&loli=true&r18=true"
	msgof   = make(map[int64]int64)
	block   = false
)

func init() { // 插件主体
	// 初始化 classify
	classify.Init(datapath)
	zero.OnRegex(`^设置随机图片网址(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			if !strings.HasPrefix(url, "http") {
				ctx.Send("URL非法!")
			} else {
				randapi = url
				ctx.Send("设置好啦")
			}
		})
	// 有保护的随机图片
	zero.OnFullMatch("随机图片", zero.OnlyGroup).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if classify.CanVisit(5) {
				go func() {
					class, lastvisit, dhash, comment := classify.Classify(randapi, false)
					replyClass(ctx, dhash, class, false, lastvisit, comment)
				}()
			} else {
				ctx.Send("你太快啦!")
			}
		})
	// 直接随机图片，无r18保护，后果自负。如果出r18图可尽快通过发送"太涩了"撤回
	zero.OnFullMatch("直接随机", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			if block {
				ctx.Send("请稍后再试哦")
			} else if randapi != "" {
				block = true
				var url string
				if randapi[0] == '&' {
					url = lolipxy
				} else {
					url = randapi
				}
				setLastMsg(ctx.Event.GroupID, ctx.Send(message.Image(url).Add("cache", "0")))
				block = false
			}
		})
	// 撤回最后的直接随机图片
	zero.OnFullMatch("太涩了").SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			go cancel(ctx)
		})
	// 上传一张图进行评价
	zero.OnKeywordGroup([]string{"评价图片"}, zero.OnlyGroup, picture.CmdMatch, picture.MustGiven).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中...")
			for _, url := range ctx.State["image_url"].([]string) {
				go func(target string) {
					class, lastvisit, dhash, comment := classify.Classify(target, true)
					replyClass(ctx, dhash, class, true, lastvisit, comment)
				}(url)
			}
		})
	zero.OnRegex(`^给你点提示哦：(.*)$`, zero.OnlyPrivate).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			dhash := ctx.State["regex_matched"].([]string)[1]
			if len(dhash) == 5*3 {
				ctx.Send(message.Image(apihead + dhash))
			}
		})
}

func setLastMsg(id int64, msg int64) {
	msgof[id] = msg
}

func cancel(ctx *zero.Ctx) {
	msg, ok := msgof[ctx.Event.GroupID]
	if ok {
		ctx.DeleteMessage(msg)
		delete(msgof, ctx.Event.GroupID)
	}
}

func replyClass(ctx *zero.Ctx, dhash string, class int, noimg bool, lv int64, comment string) {
	img := message.Image(cacheuri + strconv.FormatInt(lv, 10))
	if class > 5 {
		if dhash != "" && !noimg {
			b14, err3 := url.QueryUnescape(dhash)
			if err3 == nil {
				ctx.Send(comment + "\n给你点提示哦：" + b14)
				ctx.Event.GroupID = 0
				ctx.Send(img)
			}
		} else {
			ctx.Send(comment)
		}
	} else {
		comment := message.Text(comment)
		if !noimg {
			ctx.SendChain(img, comment)
		} else {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), comment)
		}
	}
}
