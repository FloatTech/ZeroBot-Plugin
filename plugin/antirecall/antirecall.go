// Package antirecall 反闪照、反撤回
package antirecall

import (
	"strconv"
	"strings"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("antirecall", &control.Options{
		DisableOnDefault: true,
		Help:             "反闪照、反撤回",
	})
	engine.OnRegex(`^\[CQ:image,file=`, zero.KeywordRule("type=flash")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			supersu := zero.BotConfig.SuperUsers[0]
			raw := ctx.Event.RawMessage
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			username := ctx.CardOrNickName(uid)
			groupname := ctx.GetGroupInfo(gid, true).Name
			img := strings.ReplaceAll(raw, "type=flash,", "")
			text := message.UnescapeCQCodeText("捕捉到了一个闪照！\n" +
				"来自用户:[" + username + "](" + strconv.FormatInt(uid, 10) + ")\n" +
				"来自群聊:[" + groupname + "](" + strconv.FormatInt(gid, 10) + ")\n" +
				"以下是原图：\n" + img)
			ctx.SendPrivateMessage(supersu, message.ParseMessageFromString(text))
		})
	engine.On("notice/group_recall").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			supersu := zero.BotConfig.SuperUsers[0]
			raw := ctx.GetMessage(ctx.Event.MessageID.(message.MessageID)).Elements.String()
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			username := ctx.CardOrNickName(uid)
			groupname := ctx.GetGroupInfo(gid, true).Name
			text := message.UnescapeCQCodeText("捕捉到了一条消息！\n" +
				"来自用户:[" + username + "](" + strconv.FormatInt(uid, 10) + ")\n" +
				"来自群聊:[" + groupname + "](" + strconv.FormatInt(gid, 10) + ")\n" +
				"以下是源消息：\n" + raw)
			ctx.SendPrivateMessage(supersu, message.ParseMessageFromString(text))
		})
}
