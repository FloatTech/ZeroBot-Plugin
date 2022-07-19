// Package event 好友申请以及群聊邀请事件处理
package event

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type config struct {
	AutoAcceptFriendAdd   bool `json:"AutoAcceptFriendAdd"`
	AutoAcceptGroupInvite bool `json:"AutoAcceptGroupInvite"`
}

var (
	cfg = config{
		AutoAcceptFriendAdd:   false,
		AutoAcceptGroupInvite: false,
	}
)

func init() {
	engine := control.Register("event", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "好友申请以及群聊邀请事件处理，默认发送给主人列表第一位\n" +
			" - [开启|关闭]自动同意申请\n" +
			" - [开启|关闭]自动同意邀请",
		PrivateDataFolder: "event",
	})
	path := engine.DataFolder()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
	cfgFile := engine.DataFolder() + "config.json"
	if file.IsExist(cfgFile) {
		reader, err := os.Open(cfgFile)
		if err == nil {
			err = json.NewDecoder(reader).Decode(&cfg)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
		err = reader.Close()
		if err != nil {
			panic(err)
		}
	} else {
		err = saveConfig(cfgFile)
		if err != nil {
			panic(err)
		}
	}
	engine.OnRequest().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.RequestType == "friend" {
				su := zero.BotConfig.SuperUsers[0]
				now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
				flag := ctx.Event.Flag
				comment := ctx.Event.Comment
				userid := ctx.Event.UserID
				username := ctx.CardOrNickName(userid)
				logrus.Infoln("[event]收到来自[", username, "](", userid, ")的好友申请")
				if cfg.AutoAcceptFriendAdd {
					ctx.SetFriendAddRequest(flag, true, "")
					ctx.SendPrivateMessage(su,
						message.Text("已自动同意在", now, "收到来自",
							"\n用户:[", username, "](", strconv.FormatInt(userid, 10), ")",
							"\n的好友请求:", comment,
							"\nflag:", flag))
					return
				}
				id := ctx.SendPrivateMessage(su,
					message.Text("在", now, "收到来自",
						"\n用户:[", username, "](", strconv.FormatInt(userid, 10), ")",
						"\n的好友请求:", comment,
						"\n请在下方复制flag并在前面加上:",
						"\n同意/拒绝申请，来决定同意还是拒绝"))
				process.SleepAbout1sTo2s()
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
				logrus.Infoln("[event]收到来自[", invitername, "](", inviterid, ")的群聊邀请，群:[", groupname, "](", groupid, ")")
				if cfg.AutoAcceptGroupInvite {
					ctx.SetGroupAddRequest(flag, "invite", true, "")
					ctx.SendPrivateMessage(su,
						message.Text("已自动同意在", now, "收到来自",
							"\n用户:[", invitername, "](", strconv.FormatInt(inviterid, 10), ")的群聊邀请",
							"\n群聊:[", groupname, "](", strconv.FormatInt(groupid, 10), ")",
							"\n验证信息:\n", comment,
							"\nflag:", flag))
					return
				}
				id := ctx.SendPrivateMessage(su,
					message.Text("在", now, "收到来自",
						"\n用户:[", invitername, "](", strconv.FormatInt(inviterid, 10), ")的群聊邀请",
						"\n群聊:[", groupname, "](", strconv.FormatInt(groupid, 10), ")",
						"\n验证信息:\n", comment,
						"\n请在下方复制flag并在前面加上:",
						"\n同意/拒绝邀请，来决定同意还是拒绝"))
				process.SleepAbout1sTo2s()
				ctx.SendPrivateMessage(su, message.ReplyWithMessage(id, message.Text(flag)))
			}
		})
	engine.OnRegex(`^同意申请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			rename := ctx.State["regex_matched"].([]string)[2]
			ctx.SetFriendAddRequest(flag, true, rename)
			ctx.SendPrivateMessage(su, message.Text("已同意好友申请"))
		})
	engine.OnRegex(`^拒绝申请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			ctx.SetFriendAddRequest(flag, false, "")
			ctx.SendPrivateMessage(su, message.Text("已拒绝好友申请"))
		})
	engine.OnRegex(`^同意邀请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			ctx.SetGroupAddRequest(flag, "invite", true, "")
			ctx.SendPrivateMessage(su, message.Text("已同意群聊邀请"))
		})
	engine.OnRegex(`^拒绝邀请\s*(\d+)\s*(.*)$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			flag := ctx.State["regex_matched"].([]string)[1]
			reason := ctx.State["regex_matched"].([]string)[2]
			ctx.SetGroupAddRequest(flag, "invite", false, reason)
			ctx.SendPrivateMessage(su, message.Text("已拒绝群聊邀请"))
		})
	engine.OnRegex(`^(.*)自动同意申请$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			switch option {
			case "开启", "打开", "启用":
				cfg.AutoAcceptFriendAdd = true
			case "关闭", "关掉", "禁用":
				cfg.AutoAcceptFriendAdd = false
			default:
				return
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("已设置自动同意好友申请为" + option))
			} else {
				ctx.SendChain(message.Text("ERROR:", err))
			}
		})
	engine.OnRegex(`^(.*)自动同意邀请$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			switch option {
			case "开启", "打开", "启用":
				cfg.AutoAcceptGroupInvite = true
			case "关闭", "关掉", "禁用":
				cfg.AutoAcceptGroupInvite = false
			default:
				return
			}
			err = saveConfig(cfgFile)
			if err == nil {
				ctx.SendChain(message.Text("已设置自动同意群聊邀请为" + option))
			} else {
				ctx.SendChain(message.Text("ERROR:", err))
			}
		})
}

func saveConfig(cfgFile string) (err error) {
	if reader, err := os.Create(cfgFile); err == nil {
		err = json.NewEncoder(reader).Encode(&cfg)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}
