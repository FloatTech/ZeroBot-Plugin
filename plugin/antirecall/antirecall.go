// Package antirecall 反闪照、反撤回
package antirecall

import (
	"fmt"
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
			raw := ctx.Event.RawMessage
			img := strings.ReplaceAll(raw, "type=flash,", "")
			img = message.UnescapeCQCodeText(img)
			text := "闪照捕捉测试"
			ctx.Send(message.ParseMessageFromString(text + img))
		})
	engine.On("notice/group_recall").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			raw := ctx.GetMessage(ctx.Event.MessageID.(message.MessageID)).Elements
			nickname := ctx.Event.Sender.NickName
			uid := ctx.Event.UserID
			msg := fmt.Sprintf("撤回捕捉测试：[%s](%d)\n原消息：%s", nickname, uid, raw)
			ctx.Send(msg)
		})
}
