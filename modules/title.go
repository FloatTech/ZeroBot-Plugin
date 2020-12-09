package modules

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(title{})
}

type title struct{}

func (title) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "title",
		Version:    "0.0.1",
		Details:    "设置群名片群头衔",
	}
}

func (title) Start() { // 插件主体
	setCard := zero.OnRegex("^修改名片.*?(\\d+).*?\\s(.*)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupCard(event.GroupID, GetInt(state, 1), GetStr(state, 2))
				zero.Send(event, "嗯！已经修改了")
				return zero.SuccessResponse
			},
		)
	setCard.Priority = 10
	setCard.Block = true

	setTitle := zero.OnRegex("^修改头衔.*?(\\d+).*?\\s(.*)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupSpecialTitle(event.GroupID, GetInt(state, 1), GetStr(state, 2))
				zero.Send(event, "嗯！已经修改了")
				return zero.SuccessResponse
			},
		)
	setTitle.Priority = 11
	setTitle.Block = true

	setSelfTitle := zero.OnRegex("^申请头衔(.*)", zero.OnlyGroup).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupSpecialTitle(event.GroupID, event.UserID, GetStr(state, 2))
				zero.Send(event, "嗯！不错的头衔呢~")
				return zero.SuccessResponse
			},
		)
	setSelfTitle.Priority = 12
	setSelfTitle.Block = true
}
