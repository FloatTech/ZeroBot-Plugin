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
		Help:             "反闪照、反撤回",
	})
	engine.OnRegex(`^\[CQ:image.*`, zero.KeywordRule("type=flash")).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			raw := ctx.Event.RawMessage
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			username := ctx.CardOrNickName(uid)
			groupname := ctx.GetGroupInfo(gid, true).Name
			img := strings.ReplaceAll(raw, "type=flash,", "")
			text := "捕捉到了一个闪照！\n" +
				"时间:" + now + "\n" +
				"来自用户:[" + username + "](" + strconv.FormatInt(uid, 10) + ")\n" +
				"来自群聊:[" + groupname + "](" + strconv.FormatInt(gid, 10) + ")\n" +
				"以下是原图:"
			fmsg := make(message.Message, 0, 10)
			fmsg = append(fmsg, message.CustomNode(username, uid, text))
			fmsg = append(fmsg, message.CustomNode(username, uid, img))
			ctx.SendPrivateForwardMessage(su, fmsg)
		})
	engine.On("notice/group_recall").SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			raw := ctx.GetMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64))).Elements.String()
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			username := ctx.CardOrNickName(uid)
			groupname := ctx.GetGroupInfo(gid, true).Name
			text := "捕捉到了一条撤回的消息！\n" +
				"时间:" + now + "\n" +
				"来自用户:[" + username + "](" + strconv.FormatInt(uid, 10) + ")\n" +
				"来自群聊:[" + groupname + "](" + strconv.FormatInt(gid, 10) + ")\n" +
				"以下是源消息："
			fmsg := make(message.Message, 0, 10)
			fmsg = append(fmsg, message.CustomNode(username, uid, text))
			fmsg = append(fmsg, message.CustomNode(username, uid, raw))
			ctx.SendPrivateForwardMessage(su, fmsg)
		})
}
