// Package ctxext zero ctx 扩展
package ctxext

import (
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// NickName 从 args 获取名字
func NickName(ctx *zero.Ctx) (name string) {
	name = ctx.State["args"].(string)
	if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
		qq, _ := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("nickname").Str
	} else if name == "" {
		name = ctx.Event.Sender.NickName
	}
	return
}

// CardOrNickName 从 uid 获取名字
func CardOrNickName(ctx *zero.Ctx, uid int64) (name string) {
	name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, uid, false).Get("card").String()
	if name == "" {
		name = ctx.GetStrangerInfo(uid, false).Get("nickname").String()
	}
	return
}
