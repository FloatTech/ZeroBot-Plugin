package setutime

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	utils "github.com/Yiwen-Chan/ZeroBot-Plugin/setutime/utils"
)

func init() { // 插件主体
	// 根据PID搜图
	zero.OnRegex(`^搜图(\d+)$`).SetBlock(true).SetPriority(30).
		Handle(func(ctx *zero.Ctx) {
			id := utils.Str2Int(ctx.State["regex_matched"].([]string)[1])
			ctx.Send("少女祈祷中......")
			// 获取P站插图信息
			illust := &utils.Illust{}
			if err := illust.IllustInfo(id); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 下载P站插图
			if _, err := illust.PixivPicDown(CACHEPATH); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 发送搜索结果
			ctx.Send(illust.DetailPic)
			return
		})
	// 以图搜图
	zero.OnMessage(FullMatchText("以图搜图", "搜索图片", "以图识图"), MustHasPicture()).SetBlock(true).SetPriority(999).
		Handle(func(ctx *zero.Ctx) {
			// 开始搜索图片
			ctx.Send("少女祈祷中......")
			for _, pic := range ctx.State["image_url"].([]string) {
				fmt.Println(pic)
				if m, err := utils.SauceNaoSearch(pic); err == nil {
					ctx.SendChain(m...) // 返回SauceNAO的结果
					continue
				} else {
					ctx.SendChain(message.Text("ERROR: ", err))
				}
				if m, err := utils.Ascii2dSearch(pic); err == nil {
					ctx.SendChain(m...) // 返回Ascii2d的结果
					continue
				} else {
					ctx.SendChain(message.Text("ERROR: ", err))
				}
			}
			return
		})
}

// FullMatchText 如果信息中文本完全匹配则返回 true
func FullMatchText(src ...string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		for _, elem := range msg {
			if elem.Type == "text" {
				text := elem.Data["text"]
				text = strings.ReplaceAll(text, " ", "")
				text = strings.ReplaceAll(text, "\r", "")
				text = strings.ReplaceAll(text, "\n", "")
				for _, s := range src {
					if text == s {
						return true
					}
				}
			}
		}
		return false
	}
}

// HasPicture 消息含有图片返回 true
func HasPicture() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		url := []string{}
		// 如果是回复信息则将信息替换成被回复的那条
		if msg[0].Type == "reply" {
			id, _ := strconv.Atoi(msg[0].Data["id"])
			msg = ctx.GetMessage(int64(id)).Elements
		}
		// 遍历信息中所有图片
		for _, elem := range msg {
			if elem.Type == "image" {
				url = append(url, elem.Data["url"])
			}
		}
		// 如果有图片就返回true
		if len(url) > 0 {
			ctx.State["image_url"] = url
			return true
		}
		return false
	}
}

// MustHasPicture 消息不存在图片阻塞60秒至有图片，超时返回 false
func MustHasPicture() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if HasPicture()(ctx) {
			return true
		}
		// 没有图片就索取
		ctx.Send("请发送一张图片")
		next := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID), HasPicture())
		recv, cancel := next.Repeat()
		select {
		case e := <-recv:
			cancel()
			newCtx := &zero.Ctx{Event: e, State: zero.State{}}
			if HasPicture()(newCtx) {
				ctx.State["image_url"] = newCtx.State["image_url"]
				return true
			}
			return false
		case <-time.After(time.Second * 60):
			return false
		}
	}
}
