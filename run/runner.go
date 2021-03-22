package runner

import (
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	zero.OnRegex(`^run(.*)$`, zero.SuperUserPermission).SetBlock(true).SetPriority(0).
		Handle(func(ctx *zero.Ctx) {
			var cmd = ctx.State["regex_matched"].([]string)[1]
			cmd = strings.ReplaceAll(cmd, "&#91;", "[")
			cmd = strings.ReplaceAll(cmd, "&#93;", "]")
			ctx.Send(cmd)
		})
}
