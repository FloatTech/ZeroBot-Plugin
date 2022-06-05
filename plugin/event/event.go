// Package event 好友申请
package event

import (
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 来自mayuri的插件
	engine := control.Register("event", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "好友申请以及群聊邀请事件处理，默认发送给主人列表第一位",
	})
	engine.OnRequest().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.RequestType == "friend" {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag := ctx.Event.Flag
				comment := ctx.Event.Comment
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				logrus.Infoln("[manager]收到来自[", username, "](", userid, ")的好友申请")
				id := ctx.SendPrivateMessage(
					su,
					message.Text("在", now,
						"\n收到来自[", username, "](", strconv.FormatInt(userid, 10), ")",
						"\n的好友请求:", comment,
						"\n请在下方复制flag并在前面加上:",
						"\n通过/拒绝邀请，来决定通过还是拒绝"))
				time.Sleep(time.Second * 1)
				ctx.SendPrivateMessage(su, message.ReplyWithMessage(id, message.Text(flag)))
			}
		})
	engine.OnRequest().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.RequestType == "group" && ctx.Event.SubType == "invite" {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag := ctx.Event.Flag
				comment := ctx.Event.Comment
				inviterid := ctx.Event.UserID
				invitername := ctx.CardOrNickName(inviterid)
				groupid := ctx.Event.GroupID
				groupname := ctx.GetGroupInfo(groupid, true).Name
				logrus.Infoln("[manager]收到来自[", invitername, "](", inviterid, ")的群聊邀请，群:[", groupname, "](", groupid, ")")
				id := ctx.SendPrivateMessage(
					su,
					message.Text("在", now,
						"\n收到来自[", invitername, "]("+strconv.FormatInt(inviterid, 10), ")的群聊邀请",
						"\n群聊:[", groupname, "]("+strconv.FormatInt(groupid, 10), ")",
						"\n验证信息:\n", comment,
						"\n请在下方复制flag并在前面加上:",
						"\n通过/拒绝邀请，来决定通过还是拒绝"))
				time.Sleep(time.Second * 1)
				ctx.SendPrivateMessage(su, message.ReplyWithMessage(id, message.Text(flag)))
			}
		})
		/*
			engine.OnRequest().SetBlock(false).
				Handle(func(ctx *zero.Ctx) {
					if ctx.Event.RequestType == "group" && ctx.Event.SubType == "invite" {
						su := zero.BotConfig.SuperUsers[0]
						now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
						inviterid := ctx.Event.UserID
						invitername := ctx.CardOrNickName(inviterid)
						groupid := ctx.Event.GroupID
						groupname := ctx.GetGroupInfo(groupid, true).Name
						logrus.Infoln("[manager]被用户[", invitername, "](", inviterid, ")拉至群:[", groupname, "](", groupid, ")")
						ctx.SendPrivateMessage(
							su,
							message.Text("在", now,
								"\n被用户[", invitername, "]("+strconv.FormatInt(inviterid, 10), ")拉至",
								"\n群聊:[", groupname, "]("+strconv.FormatInt(groupid, 10), ")"))
					}
				})
		*/
	engine.OnRegex(`^通过申请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			rename := ctx.State["regex_matched"].([]string)[2]
			ctx.SetFriendAddRequest(flag, true, rename)
			ctx.SendPrivateMessage(su, message.Text("已通过好友申请"))
		})
	engine.OnRegex(`^拒绝申请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			ctx.SetFriendAddRequest(flag, false, "")
			ctx.SendPrivateMessage(su, message.Text("已拒绝好友申请"))
		})
	engine.OnRegex(`^通过邀请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			ctx.SetGroupAddRequest(flag, "invite", true, "")
			ctx.SendPrivateMessage(su, message.Text("已通过群聊邀请"))
		})
	engine.OnRegex(`^拒绝邀请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			reason := ctx.State["regex_matched"].([]string)[2]
			ctx.SetGroupAddRequest(flag, "invite", false, reason)
			ctx.SendPrivateMessage(su, message.Text("已拒绝群聊邀请"))
		})
	/*
		engine.OnNotice().SetBlock(false).
			Handle(func(ctx *zero.Ctx) {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				subtype := ctx.Event.SubType
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				operatorid := ctx.Event.OperatorID
				operatorname := ctx.CardOrNickName(operatorid)
				groupid := ctx.Event.GroupID
				groupname := ctx.GetGroupInfo(groupid, true).Name
				switch subtype {
				case "kick_me":
					{
						ctx.SendPrivateMessage(
							su,
							message.Text("呜呜呜，我在", now,
								"被[", operatorname, "](", operatorid, ")",
								"丢出了裙[", groupname, "](", groupid, ")"))
					}
				case "kick":
					{
						ctx.SendPrivateMessage(
							su,
							message.Text("好可怕，[", username, "](", userid, ")",
								"在", now,
								"被[", operatorname, "](", operatorid, ")",
								"丢出了裙[", groupname, "](", groupid, ")"))
					}
				}
			})
	*/
}
