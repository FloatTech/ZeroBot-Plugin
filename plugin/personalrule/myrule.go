package personalrule

import (
	"fmt"
	"os"

	control "github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("mine", &control.Options{
		DisableOnDefault: false,
		Help: "Poweroff\n" +
			"- pause",
	})
	engine.OnFullMatchGroup([]string{"pause", "restart", "kill"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			os.Exit(0)
		})
	engine.OnRegex(`^\[CQ:xml`, zero.OnlyGroup, zero.KeywordRule("serviceID=\"60\"")).SetBlock(true).
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
	engine.OnRegex(`^来(.*)涩图`, zero.OnlyGroup, zero.KeywordRule("114514")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Image("https://gchat.qpic.cn/gchatpic_new/1770747317/1049468946-3068097579-76A49478EFA68B4750B10B96917F7B58/0?term=3"))
		})
}
