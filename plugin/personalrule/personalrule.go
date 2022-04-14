package personalrule

import (
	"fmt"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnRegex(`^\[CQ:xml`, zero.OnlyGroup, zero.KeywordRule("serviceID=\"60\"")).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupKick(
				ctx.Event.GroupID,
				ctx.Event.UserID,
				false,
			)
			nickname := ctx.GetGroupMemberInfo(
				ctx.Event.GroupID,
				ctx.Event.UserID,
				false,
			).Get("nickname").Str
			ctx.SetGroupBan(ctx.Event.GroupID, ctx.Event.UserID, 7*24*60*60)
			ctx.SendChain(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(fmt.Sprintf("检测到 [%s](%d) 发送了干扰性消息,已处理", nickname, ctx.Event.UserID)))...)
			ctx.DeleteMessage(ctx.Event.MessageID.(message.MessageID))
		})
	zero.OnRegex(`^来(.*)涩图`, zero.OnlyGroup, zero.KeywordRule("114514")).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image("https://gchat.qpic.cn/gchatpic_new/1770747317/1049468946-3068097579-76A49478EFA68B4750B10B96917F7B58/0?term=3"))
		})

	/*
		zero.OnRegex(`^\[CQ:reply,id=(.*)](.*)`, zero.OnlyToMe, zero.KeywordRule("recall")).SetBlock(true).SetPriority(1000).
			Handle(func(ctx *zero.Ctx) {
				var msg_id message.MessageID
				fmt.Sscanf(ctx.State["regex_matched"].([]string)[1], "%d", &(msg_id))
				fmt.Println(msg_id)
				ctx.DeleteMessage(msg_id)
			})
	*/

	/*
		zero.OnRegex(`^run(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var cmd = message.UnescapeCQText(ctx.State["regex_matched"].([]string)[1])
			ctx.SendPrivateMessage(ctx.Event.UserID, message.ParseMessageFromString(cmd))
		})
	*/
}
