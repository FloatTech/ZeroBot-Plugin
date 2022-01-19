// Package aireply AI 回复
package aireply

import (
	"errors"
	"time"

	"github.com/FloatTech/AnimeAPI/aireply"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	serviceName = "aireply"
)

var modes = [...]string{"青云客", "小爱"}

func init() { // 插件主体
	bucket := rate.NewManager(time.Minute, 20) // 接口回复限速器
	engine := control.Register(serviceName, order.PrioAIReply, &control.Options{
		DisableOnDefault: false,
		Help: "人工智能回复\n" +
			"- @Bot 任意文本(任意一句话回复)\n- 设置回复模式[青云客  |  小爱]\n- ",
	})
	// 回复 @和包括名字
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			aireply := aireply.NewAIReply(getReplyMode(ctx))
			if !bucket.Load(ctx.Event.UserID).Acquire() {
				// 频繁触发，不回复
				return
			}
			reply := aireply.Talk(ctx.ExtractPlainText())
			// 回复
			time.Sleep(time.Second * 1)
			if zero.OnlyPublic(ctx) {
				reply = append(reply, message.Reply(ctx.Event.MessageID))
				ctx.Send(reply)
				return
			}
			ctx.Send(reply)
		})
	engine.OnPrefix(`设置回复模式`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			err := setReplyMode(ctx, param)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
		})
}

func setReplyMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range modes {
		if s == name {
			ok = true
			index = int64(i)
			break
		}
	}
	if !ok {
		return errors.New("no such mode")
	}
	m, ok := control.Lookup(serviceName)
	if !ok {
		return errors.New("no such plugin")
	}
	return m.SetData(gid, index)
}

func getReplyMode(ctx *zero.Ctx) (name string) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := control.Lookup(serviceName)
	if ok {
		index := m.GetData(gid)
		if int(index) < len(modes) {
			return modes[index]
		}
	}
	return "青云客"
}
