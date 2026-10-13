package aichat

import (
	goba "github.com/fumiama/go-onebot-agent"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// Group owners retain the group-admin action list. Only configured bot
// owners receive the owner list, including cross-group and private actions.
func agentRole(ctx *zero.Ctx) goba.PermRole {
	if zero.SuperUserPermission(ctx) {
		return goba.PermRoleOwner
	}
	if ctx.Event.GroupID > 0 && ctx.Event.Sender != nil && zero.AdminPermission(ctx) {
		return goba.PermRoleAdmin
	}
	return goba.PermRoleUser
}
