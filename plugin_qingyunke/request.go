package qingyunke


import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

//自动同意加群，加好友
func init() {
	zero.OnRequest().SetBlock(false).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if ctx.Event.RequestType == "friend" {
			log.Println("加好友")
			ctx.SetFriendAddRequest(ctx.Event.Flag, true, "")
		}
		if ctx.Event.RequestType == "group" && ctx.Event.SubType == "invite" {
			log.Println("加群")
			ctx.SetGroupAddRequest(ctx.Event.Flag, "invite", true, "我爱你，mua~")
		}
	})

}
