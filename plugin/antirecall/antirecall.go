// Package antirecall 反闪照、反撤回
package antirecall

import (
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("antirecall", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "反闪照、反撤回",
		Help:             "反闪照、反撤回",
	})
	engine.OnRegex(`^\[CQ:image.*`, zero.KeywordRule("type=flash")).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			raw := ctx.Event.RawMessage
			uid := ctx.Event.UserID
			if uid == su {
				return
			}
			gid := ctx.Event.GroupID
			botid := ctx.Event.SelfID
			botname := zero.BotConfig.NickName[0]
			username := ctx.CardOrNickName(uid)
			msg := make(message.Message, 10)
			msg = append(msg, message.CustomNode(botname, botid, "捕捉到了一个闪照！\n"+"时间:"+now+"\n"))
			if gid != 0 {
				groupname := ctx.GetGroupInfo(gid, true).Name
				msg = append(msg, message.CustomNode(botname, botid, "来自群聊:["+groupname+"]("+strconv.FormatInt(gid, 10)+")\n来自用户:["+username+"]("+strconv.FormatInt(uid, 10)+")\n以下是原图:"))
			} else {
				msg = append(msg, message.CustomNode(botname, botid, "来自私聊:["+username+"]("+strconv.FormatInt(uid, 10)+")\n以下是原图:"))
			}
			img := strings.ReplaceAll(raw, ",type=flash", "")
			msg = append(msg, message.CustomNode(username, uid, img))
			ctx.SendPrivateForwardMessage(su, msg)
		})
	engine.OnNotice().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType != "group_recall" && ctx.Event.NoticeType != "friend_recall" {
				return
			}
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			raw := ctx.GetMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64))).Elements
			rawmsg := raw.String()
			uid := ctx.Event.UserID
			botid := ctx.Event.SelfID
			if uid == su || uid == botid || strings.Contains(rawmsg, ",type=flash") || strings.Contains(rawmsg, "CQ:reply") && strings.Contains(rawmsg, "撤回") && uid == su {
				return
			}
			gid := ctx.Event.GroupID
			botname := zero.BotConfig.NickName[0]
			username := ctx.CardOrNickName(uid)
			msg := make(message.Message, 10)
			msg = append(msg, message.CustomNode(botname, botid, "捕捉到了一条撤回的消息！\n"+"时间:"+now))
			if gid != 0 {
				groupname := ctx.GetGroupInfo(gid, true).Name
				msg = append(msg, message.CustomNode(botname, botid, "来自群聊:["+groupname+"]("+strconv.FormatInt(gid, 10)+")\n来自用户:["+username+"]("+strconv.FormatInt(uid, 10)+")\n以下是源消息："))
			} else {
				msg = append(msg, message.CustomNode(botname, botid, "来自私聊:["+username+"]("+strconv.FormatInt(uid, 10)+")\n以下是源消息："))
			}
			if strings.Contains(rawmsg, "CQ:record") {
				defer ctx.SendChain(raw...)
			}
			msg = append(msg, message.CustomNode(username, uid, raw))
			ctx.SendPrivateForwardMessage(su, msg)
		})
}
