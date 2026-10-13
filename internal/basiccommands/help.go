// Package basiccommands registers built-in commands independently of plugins.
package basiccommands

import (
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// RegisterHelp accepts surrounding whitespace from QQ clients, while preserving
// the existing aliases, exact command names, priority and private/at-me behavior.
func RegisterHelp(banner string) {
	zero.OnMessage(func(ctx *zero.Ctx) bool {
		switch strings.TrimSpace(ctx.MessageString()) {
		case "help", "/help", ".help", "菜单":
			return true
		default:
			return false
		}
	}, zero.OnlyToMe).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text(banner, "\n发送\"/服务列表\"查看功能与用法（聊天记录）\n发送\"/服务列表 2\"翻页；\"/用法name\"查看完整用法\n群内请 @ 机器人"))
	})
}
