// Package event 好友申请以及群聊邀请事件处理
package event

import (
	"encoding/binary"
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
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "好友申请和群聊邀请事件处理",
		Help: "- [开启|关闭]自动同意[申请|邀请|主人]\n" +
			"- [同意|拒绝][申请|邀请][flag]\n" +
			"Tips: 信息默认发送给主人列表第一位, 默认同意所有主人的事件, flag跟随事件一起发送",
	})
	engine.On("request/group/invite").SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if ok {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag, err := strconv.ParseInt(ctx.Event.Flag, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				var buf [8]byte
				binary.BigEndian.PutUint64(buf[:], uint64(flag))
				es := base14.EncodeToString(buf[1:])
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				data := (storage)(c.GetData(-su))
				groupname := ctx.GetThisGroupInfo(true).Name
				groupid := ctx.Event.GroupID
				logrus.Info("[event]收到来自[", username, "](", userid, ")的群聊邀请，群:[", groupname, "](", groupid, ")")
				if data.isinviteon() || (!data.ismasteroff() && zero.SuperUserPermission(ctx)) {
					ctx.SetGroupAddRequest(ctx.Event.Flag, "invite", true, "")
					ctx.SendPrivateForwardMessage(su, message.Message{message.CustomNode(username, userid,
						"已自动同意在"+now+"收到来自"+
							"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")的群聊邀请"+
							"\n群聊:["+groupname+"]("+strconv.FormatInt(groupid, 10)+")"+
							"\nflag:"+es)})
					return
				}
				ctx.SendPrivateForwardMessage(su,
					message.Message{message.CustomNode(username, userid,
						"在"+now+"收到来自"+
							"\n用户:["+username+"]("+strconv.FormatInt(userid, 10)+")的群聊邀请"+
							"\n群聊:["+groupname+"]("+strconv.FormatInt(groupid, 10)+")"+
							"\n请在下方复制flag并在前面加上:"+
							"\n同意/拒绝邀请，来决定同意还是拒绝"),
						message.CustomNode(username, userid, es)})
			}
		})
	engine.On("request/friend").SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if ok {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag, err := strconv.ParseInt(ctx.Event.Flag, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				var buf [8]byte
				binary.BigEndian.PutUint64(buf[:], uint64(flag))
				es := base14.EncodeToString(buf[1:])
				comment := ctx.Event.Comment
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				data := (storage)(c.GetData(-su))
				logrus.Info("[event]收到来自[", username, "](", userid, ")的好友申请")
				if data.isapplyon() || (!data.ismasteroff() && zero.SuperUserPermission(ctx)) {
					ctx.SetFriendAddRequest(ctx.Event.Flag, true, "")
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
	engine.OnRegex(`^(同意|拒绝)(申请|邀请)\s*([一-踀]{4})\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			cmd := ctx.State["regex_matched"].([]string)[1]
			org := ctx.State["regex_matched"].([]string)[2]
			es := ctx.State["regex_matched"].([]string)[3]
			other := ctx.State["regex_matched"].([]string)[4]
			var buf [8]byte
			copy(buf[1:], base14.DecodeFromString(es))
			flag := strconv.FormatInt(int64(binary.BigEndian.Uint64(buf[:])), 10)
			ok := cmd == "同意"
			switch org {
			case "申请":
				ctx.SetFriendAddRequest(flag, ok, other)
				ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
			case "邀请":
				ctx.SetGroupAddRequest(flag, "invite", ok, other)
				ctx.SendPrivateMessage(su, message.Text("已", cmd, org))
			}
		})
	engine.OnRegex(`^(开启|关闭)自动同意(申请|邀请|主人)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			su := zero.BotConfig.SuperUsers[0]
			option := ctx.State["regex_matched"].([]string)[1]
			from := ctx.State["regex_matched"].([]string)[2]
			data := (storage)(c.GetData(-su))
			switch from {
			case "申请":
				data.setapply(option == "开启")
			case "邀请":
				data.setinvite(option == "开启")
			case "主人":
				data.setmaster(option == "关闭")
			}
			err := c.SetData(-su, int64(data))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("已设置自动同意" + from + "为" + option))
		})
}
