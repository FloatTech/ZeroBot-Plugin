package modules

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(manage{})
}

type manage struct{}

func (manage) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "manage",
		Version:    "0.0.1",
		Details:    "管理任命 踢出退出群聊",
	}
}

func (manage) Start() { // 插件主体
	promoteManager := zero.OnRegex("^升为管理.*?(\\d+)", zero.OnlyGroup, zero.SuperUserPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupAdmin(event.GroupID, GetInt(state, 1), true)
				nickname := GetNickname(event.GroupID, GetInt(state, 1))
				zero.Send(event, nickname+" 升为了管理~")
				return zero.SuccessResponse
			},
		)
	promoteManager.Priority = 20
	promoteManager.Block = true

	cancleManager := zero.OnRegex("^取消管理.*?(\\d+)", zero.OnlyGroup, zero.SuperUserPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupAdmin(event.GroupID, GetInt(state, 1), false)
				nickname := GetNickname(event.GroupID, GetInt(state, 1))
				zero.Send(event, "残念~ "+nickname+" 暂时失去了管理员的资格")
				return zero.SuccessResponse
			},
		)
	cancleManager.Priority = 21
	cancleManager.Block = true

	kick := zero.OnRegex("^踢出群聊.*?(\\d+)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupKick(event.GroupID, GetInt(state, 1), false)
				nickname := GetNickname(event.GroupID, GetInt(state, 1))
				zero.Send(event, "残念~ "+nickname+"("+GetStr(state, 1)+") 被放逐")
				return zero.SuccessResponse
			},
		)
	kick.Priority = 22
	kick.Block = true

	leave := zero.OnRegex("^退出群聊.*?(\\d+)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupLeave(GetInt(state, 1), true)
				zero.Send(event, "姬气人已主动离开群聊 "+GetStr(state, 1))
				return zero.SuccessResponse
			},
		)
	leave.Priority = 23
	leave.Block = true

}
