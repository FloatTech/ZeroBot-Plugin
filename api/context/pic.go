package context

import (
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

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
				ctx.Event.MessageID = newCtx.Event.MessageID
				return true
			}
			return false
		case <-time.After(time.Second * 60):
			return false
		}
	}
}
