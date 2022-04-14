package manager

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/math"
)

type storage uint64

func init() {
	// 指定开启某群全群禁言 Usage: 开启全员禁言123456
	zero.OnRegex(`^开启全员禁言.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始"))
		})
	// 指定解除某群全群禁言 Usage: 解除全员禁言123456
	zero.OnRegex(`^解除全员禁言.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束"))
		})
	zero.OnRequest().SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.RequestType == "friend" {
				logrus.Debugln("[manager]收到好友申请, 用户:", ctx.CardOrNickName(ctx.Event.UserID), "(", ctx.Event.UserID, ")", "验证消息:", ctx.Event.Comment, "Flag", ctx.Event.Flag)
				su := zero.BotConfig.SuperUsers[0]
				ctx.SendPrivateMessage(
					su,
					message.Text(
						"在"+
							time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")+
							"收到来自"+
							ctx.CardOrNickName(ctx.Event.UserID)+
							"("+
							strconv.FormatInt(ctx.Event.UserID, 10)+
							")好友请求:"+ctx.Event.Comment+
							"\n输入:\n-通过申请"+ctx.Event.Flag+
							"\n-拒绝申请"+ctx.Event.Flag),
				)
			}
		})
	zero.OnPrefix("通过申请", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetFriendAddRequest(strings.ReplaceAll(ctx.Event.RawMessage, "通过申请", ""), true, "")
			ctx.Send(message.Text("已通过好友申请"))
		})

	zero.OnPrefix("拒绝申请", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetFriendAddRequest(strings.ReplaceAll(ctx.Event.RawMessage, "拒绝申请", ""), false, "")
			ctx.Send(message.Text("已拒绝好友申请"))
		})
}
