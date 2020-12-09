package modules

import (
	"gm/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(request{})
}

type request struct{}

func (request) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "request",
		Version:    "0.0.1",
		Details:    "设置群名片群头衔",
	}
}

func (request) Start() { // 插件主体
	friendAdd := zero.OnNotice().
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				if event.RequestType == "friend" {
					zero.SendPrivateMessage(utils.Str2Int(Conf.Master[0]), "有人想加我")
				}
				return zero.SuccessResponse
			},
		)
	friendAdd.Priority = 50
	friendAdd.Block = true

	groupAdd := zero.OnNotice().
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				if event.RequestType == "friend" {
					if event.SubType == "add" {
						zero.Send(event, "有人申请加群")
					} else {
						zero.SendPrivateMessage(utils.Str2Int(Conf.Master[0]), "有人想拉我入群")
					}
				}
				return zero.SuccessResponse
			},
		)
	groupAdd.Priority = 51
	groupAdd.Block = true
}
