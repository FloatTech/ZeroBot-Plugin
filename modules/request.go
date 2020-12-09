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
		Details:    "申请添加好友 加入群聊 邀请群聊",
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
				if event.RequestType == "group" {
					nickname := GetNickname(event.GroupID, event.UserID)
					if event.SubType == "add" {
						zero.Send(event, nickname+"("+utils.Int2Str(event.UserID)+") 申请加群")
					} else {
						zero.SendPrivateMessage(utils.Str2Int(Conf.Master[0]), nickname+"("+utils.Int2Str(event.UserID)+") 邀请我加入群 "+utils.Int2Str(event.GroupID))
					}
				}
				return zero.SuccessResponse
			},
		)
	groupAdd.Priority = 51
	groupAdd.Block = true
}
