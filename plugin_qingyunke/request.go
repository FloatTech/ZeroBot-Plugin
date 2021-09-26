package qingyunke

// TODO: 移动到 manager 搭配自动验证使用

/*
import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

//自动同意加群，加好友
func init() {
	zero.OnRequest().SetBlock(false).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if ctx.Event.RequestType == "friend" {
			ctx.SetFriendAddRequest(ctx.Event.Flag, true, "")
		}
		if ctx.Event.RequestType == "group" && ctx.Event.SubType == "invite" {
			ctx.SetGroupAddRequest(ctx.Event.Flag, "invite", true, "我爱你，mua~")
		}
	})
}
*/
