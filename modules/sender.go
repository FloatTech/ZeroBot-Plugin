package modules

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(sender{})
}

type sender struct{}

func (sender) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "sender",
		Version:    "0.0.1",
		Details:    "设置群名片群头衔",
	}
}

func (sender) Start() { // 插件主体
	promoteManager := zero.OnRegex("^群聊转发.*?(\\d+)\\s(.*)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SendGroupMessage(GetInt(state, 1), GetStr(state, 2))
				zero.Send(event, "complete!")
				return zero.SuccessResponse
			},
		)
	promoteManager.Priority = 30
	promoteManager.Block = true

	cancleManager := zero.OnRegex("^私聊转发.*?(\\d+)\\s(.*)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SendPrivateMessage(GetInt(state, 1), GetStr(state, 2))
				zero.Send(event, "complete!")
				return zero.SuccessResponse
			},
		)
	cancleManager.Priority = 31
	cancleManager.Block = true
}
