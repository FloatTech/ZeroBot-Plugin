package control

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func sendServiceList(ctx *zero.Ctx) {
	model := extension.CommandModel{}
	_ = ctx.Parse(&model)
	managers := cpmp2lstbyprio()
	page, pages, err := serviceListPage(model.Args, len(managers))
	if err != nil {
		ctx.SendChain(message.Text(err.Error()))
		return
	}
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	entries := make([]serviceListEntry, len(managers))
	start, end := (page-1)*serviceListPageSize, min(page*serviceListPageSize, len(managers))
	for i := start; i < end; i++ {
		manager := managers[i]
		state, err := manager.SwitchState(gid)
		if err != nil {
			ctx.SendChain(message.Text("服务列表状态暂不可读取，请检查本机配置数据库。"))
			return
		}
		entries[i] = serviceListEntry{Name: manager.Service, Brief: manager.Options.Brief, Help: manager.Options.Help, Enabled: state.EffectiveEnabled}
	}
	name := "ZeroBot 服务列表"
	if len(zero.BotConfig.NickName) > 0 {
		name = zero.BotConfig.NickName[0]
	}
	if ctx.Send(serviceListMessage(entries, page, pages, name, ctx.Event.SelfID, ctx.Event.GroupID != 0)).ID() == 0 {
		ctx.SendChain(message.Text("服务列表聊天记录发送失败，请稍后重试；也可在本机后台查看插件与用法。"))
	}
}
