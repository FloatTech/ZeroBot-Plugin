package modules

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(mute{})
}

type mute struct{}

func (mute) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "mute",
		Version:    "0.0.1",
		Details:    "禁言",
	}
}

func (mute) Start() { // 插件主体
	unWholeBan := zero.OnRegex("^解除全员禁言$", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupWholeBan(event.GroupID, false)
				zero.Send(event, "全员自闭结束~")
				return zero.SuccessResponse
			},
		)
	unWholeBan.Priority = 1
	unWholeBan.Block = true

	wholeBan := zero.OnRegex("^开启全员禁言$", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupWholeBan(event.GroupID, true)
				zero.Send(event, "全员自闭开始~")
				return zero.SuccessResponse
			},
		)
	wholeBan.Priority = 2
	wholeBan.Block = true

	unban := zero.OnRegex("^解除禁言.*?(\\d+)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupBan(event.GroupID, GetInt(state, 1), 0)
				zero.Send(event, "小黑屋释放成功~")
				return zero.SuccessResponse
			},
		)
	unban.Priority = 3
	unban.Block = true

	ban := zero.OnRegex("^禁言.*?(\\d+).*?\\s(\\d+)", zero.OnlyGroup, zero.AdminPermission).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupBan(event.GroupID, GetInt(state, 1), GetInt(state, 2)*60)
				zero.Send(event, "小黑屋收留成功~")
				return zero.SuccessResponse
			},
		)
	ban.Priority = 4
	ban.Block = true

	selfBan := zero.OnRegex("^我要自闭.*?(\\d+)分钟", zero.OnlyGroup).
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				zero.SetGroupBan(event.GroupID, event.UserID, GetInt(state, 1)*60)
				zero.Send(event, "那我就不手下留情了")
				return zero.SuccessResponse
			},
		)
	selfBan.Priority = 5
	selfBan.Block = true
}
