// Package chrev 英文字符反转
package chrev

import (
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	// 初始化engine
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "英文字符翻转",
		Help:             "例: 翻转 I love you",
	})
	// 处理字符翻转指令
	engine.OnRegex(`^翻转\s*([A-Za-z\s]*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取需要翻转的字符串
			str := ctx.State["regex_matched"].([]string)[1]
			// 将字符顺序翻转
			tmp := strings.Builder{}
			for i := len(str) - 1; i >= 0; i-- {
				tmp.WriteRune(charMap[str[i]])
			}
			// 发送翻转后的字符串
			ctx.SendChain(message.Text(tmp.String()))
		})
}
