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
		Help:             "好友申请，默认发送给主人列表第一位",
	})
	engine.OnRequest().SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.RequestType == "friend" {
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				comment := ctx.Event.Comment
				flag := ctx.Event.Flag
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				logrus.Infoln("[manager]收到好友申请, 用户:", username, "(", userid, ")", "验证消息:", comment, "Flag", flag)
				su := zero.BotConfig.SuperUsers[0]
				ctx.SendPrivateMessage(
					su,
					message.Text(
						"在"+now+"收到来自"+username+"("+strconv.FormatInt(userid, 10)+")好友请求:\n"+comment+"\n输入:\n-通过申请"+flag+"\n-拒绝申请"+flag),
				)
			}
		})
	engine.OnRequest().SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.RequestType == "group" {
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				groupid := ctx.Event.GroupID
				groupname := ctx.GetGroupInfo(groupid, true).Name
				comment := ctx.Event.Comment
				flag := ctx.Event.Flag
				inviterid := ctx.Event.UserID
				invitername := ctx.CardOrNickName(inviterid)
				logrus.Infoln("[manager]收到", "来自", invitername, "(", inviterid, ")", "的", "群邀请\n群:", groupname, "(", groupid, ")", "\n验证消息:", comment, "\nFlag", flag)
				su := zero.BotConfig.SuperUsers[0]
				ctx.SendPrivateMessage(
					su,
					message.Text("在"+now+"收到来自"+invitername+"("+strconv.FormatInt(inviterid, 10)+")群邀请\n群:"+groupname+"("+strconv.FormatInt(groupid, 10)+")"+"\n验证信息:\n"+comment+"\n输入:\n-通过邀请"+flag+"\n-拒绝邀请"+flag),
				)
			}
		})

	engine.OnRegex(`^通过申请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetFriendAddRequest(ctx.State["regex_matched"].([]string)[1], true, ctx.State["regex_matched"].([]string)[2])
			ctx.Send(message.Text("已通过好友申请"))
		})

	engine.OnRegex(`^拒绝申请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetFriendAddRequest(ctx.State["regex_matched"].([]string)[1], false, "")
			ctx.Send(message.Text("已拒绝好友申请"))
		})
	engine.OnRegex(`^通过邀请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAddRequest(ctx.State["regex_matched"].([]string)[1], "invite", true, "")
			ctx.Send(message.Text("已通过群邀请"))
		})

	engine.OnRegex(`^拒绝邀请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAddRequest(ctx.State["regex_matched"].([]string)[1], "invite", false, ctx.State["regex_matched"].([]string)[2])
			ctx.Send(message.Text("已拒绝群邀请"))
		})

	engine.OnNotice().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			eventtype := ctx.Event.NoticeType
			userid := ctx.Event.UserID
			username := ctx.CardOrNickName(userid)
			operatorid := ctx.Event.OperatorID
			operatorname := ctx.CardOrNickName(operatorid)
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			groupid := ctx.Event.GroupID
			groupname := ctx.GetGroupInfo(groupid, true).Name
			switch eventtype {
			case "kickme":
				{
					ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0],
						message.Text("呜呜呜，我在", now, "被", operatorname, "(", operatorid, ")", "丢出了裙", groupname, "(", groupid, ")"))
				}
			case "kick":
				{
					ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0],
						message.Text("好可怕，", username, "(", userid, ")", "在", now, "被", operatorname, "(", operatorid, ")", "丢出了裙", groupname, "(", groupid, ")"))
				}
			}
		})
}
