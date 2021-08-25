// Package shindan 基于 https://shindanmaker.com 的测定小功能
package shindan

import (
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/shindanmaker"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// 限制调用频率
	limit = rate.NewManager(time.Minute*5, 5)
)

func init() {
	zero.OnPrefix("异世界转生", number(587874)).SetBlock(true).FirstPriority().Handle(handle)
	zero.OnPrefix("今天是什么少女", number(162207)).SetBlock(true).FirstPriority().Handle(handle)
	zero.OnPrefix("卖萌", number(360578)).SetBlock(true).FirstPriority().Handle(handle)
	zero.OnPrefix("抽老婆", number(1075116)).SetBlock(true).FirstPriority().Handle(handle)
}

// shindanmaker 处理函数
func handle(ctx *zero.Ctx) {
	if !limit.Load(ctx.Event.UserID).Acquire() {
		ctx.SendChain(message.Text("请稍后重试0x0..."))
		return
	}
	// 获取名字
	name := ctx.State["args"].(string)
	if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
		qq, _ := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("nickname").Str
	} else if name == "" {
		name = ctx.Event.Sender.NickName
	}
	// 调用接口
	text, err := shindanmaker.Shindanmaker(ctx.State["id"].(int64), name)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
	}
	ctx.Send(text)
}

// 传入 shindanmaker id
func number(id int64) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		ctx.State["id"] = id
		return true
	}
}
