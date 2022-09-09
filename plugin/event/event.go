// Package event 好友申请以及群聊邀请事件处理
package event

import (
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	base14 "github.com/fumiama/go-base16384"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("event", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "好友申请以及群聊邀请事件处理，默认发送给主人列表第一位\n" +
			" - [开启|关闭]自动同意[申请|邀请|主人]\n" +
			" - [同意|拒绝][申请|邀请][flag]\n" +
			"flag跟随事件一起发送, 默认同意主人的事件",
	})
	engine.On("request/group/invite").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if ok {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag := ctx.Event.Flag
				es := base14.EncodeString(flag)
				comment := ctx.Event.Comment
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				data := c.GetData(-su)
				groupid := ctx.Event.GroupID
				groupname := ctx.GetGroupInfo(groupid, true).Name
				logrus.Info("[event]收到来自[", username, "](", userid, ")的群聊邀请，群:[", groupname, "](", groupid, ")")
				if data&2 == 2 || data&1 != 1 && zero.SuperUserPermission(ctx) {
					ctx.SetGroupAddRequest(flag, "invite", true, "")
					ctx.SendPrivateForwardMessage(su, message.Message{message.CustomNode(username, userid,
						"已自动同意在"+now+"收到来自"+
							"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")的群聊邀请"+
							"\n群聊:["+groupname+"]("+strconv.FormatInt(groupid, 10)+")"+
							"\n验证信息:\n"+comment+
							"\nflag:"+es)})
					return
				}
				ctx.SendPrivateForwardMessage(su,
					message.Message{message.CustomNode(username, userid,
						"在"+now+"收到来自"+
							"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")的群聊邀请"+
							"\n群聊:["+groupname+"]("+strconv.FormatInt(groupid, 10)+")"+
							"\n验证信息:\n"+comment+
							"\n请在下方复制flag并在前面加上:"+
							"\n同意/拒绝邀请，来决定同意还是拒绝"),
						message.CustomNode(username, userid, es)})
			}
		})
	engine.On("request/friend").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if ok {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag := ctx.Event.Flag
				es := base14.EncodeString(flag)
				comment := ctx.Event.Comment
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				data := c.GetData(-su)
				logrus.Info("[event]收到来自[", username, "](", userid, ")的好友申请")
				if data&4 == 4 || data&1 != 1 && zero.SuperUserPermission(ctx) {
					ctx.SetFriendAddRequest(flag, true, "")
					ctx.SendPrivateForwardMessage(su, message.Message{message.CustomNode(username, userid,
						"已自动同意在"+now+"收到来自"+
							"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")"+
							"\n的好友请求:"+comment+
							"\nflag:"+es)})
					return
				}
				ctx.SendPrivateForwardMessage(su,
					message.Message{message.CustomNode(username, userid,
						"在"+now+"收到来自"+
							"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")"+
							"\n的好友请求:"+comment+
							"\n请在下方复制flag并在前面加上:"+
							"\n同意/拒绝申请，来决定同意还是拒绝"),
						message.CustomNode(username, userid, es)})
			}
		})
	engine.OnRegex(`^(同意|拒绝)(申请|邀请)\s*([一-踀]+[㴁-㴆]?)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			cmd := ctx.State["regex_matched"].([]string)[1]
			org := ctx.State["regex_matched"].([]string)[2]
			flag := ctx.State["regex_matched"].([]string)[3]
			other := ctx.State["regex_matched"].([]string)[4]
			es := base14.DecodeString(flag)
			switch cmd {
			case "同意":
				switch org {
				case "申请":
					ctx.SetFriendAddRequest(es, true, other)
					ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
				case "邀请":
					ctx.SetGroupAddRequest(es, "invite", true, "")
					ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
				}
			case "拒绝":
				switch org {
				case "申请":
					ctx.SetFriendAddRequest(es, false, "")
					ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
				case "邀请":
					ctx.SetGroupAddRequest(es, "invite", false, other)
					ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
				}
			}
		})
	engine.OnRegex(`^(开启|关闭)自动同意(申请|邀请|主人)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			su := zero.BotConfig.SuperUsers[0]
			option := ctx.State["regex_matched"].([]string)[1]
			from := ctx.State["regex_matched"].([]string)[2]
			data := c.GetData(-su)
			switch option {
			case "开启":
				switch from {
				case "申请":
					data |= 4
				case "邀请":
					data |= 2
				case "主人":
					data &= 7
				}
			case "关闭":
				switch from {
				case "申请":
					data &= 3
				case "邀请":
					data &= 5
				case "主人":
					data |= 1
				}
			}
			err := c.SetData(-su, data)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Text("已设置自动同意" + from + "为" + option))
		})
}
