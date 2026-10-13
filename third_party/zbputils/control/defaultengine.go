package control

import (
	_ "unsafe" // for linkname to defaultEngine

	zero "github.com/wdvxdr1123/ZeroBot"
)

//go:linkname defaultEngine github.com/wdvxdr1123/ZeroBot.defaultEngine
var defaultEngine *zero.Engine

func init() {
	defaultEngine.UsePreHandler(
		func(ctx *zero.Ctx) bool {
			// 防止自触发
			return ctx.Event.UserID != ctx.Event.SelfID || ctx.Event.PostType != "message"
		},
		addsyncxstate,
	)
}
