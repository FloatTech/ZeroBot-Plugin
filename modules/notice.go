package modules

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(notice{})
}

type notice struct{}

func (notice) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "notice",
		Version:    "0.0.1",
		Details:    "进群退群提醒 戳一戳",
	}
}

func (notice) Start() { // 插件主体
	notify := zero.OnNotice().
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				if event.NoticeType == "notify" {
					time.Sleep(time.Second * 1)
					zero.Send(event, "请不要戳我 >_<")
				}
				return zero.SuccessResponse
			},
		)
	notify.Priority = 40
	notify.Block = true

	increase := zero.OnNotice().
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				if event.NoticeType == "group_increase" {
					zero.Send(event, "欢迎~")
				}
				return zero.SuccessResponse
			},
		)
	increase.Priority = 41
	increase.Block = true

	decrease := zero.OnNotice().
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				if event.NoticeType == "group_decrease" {
					zero.Send(event, "有人跑路了")
				}
				return zero.SuccessResponse
			},
		)
	decrease.Priority = 42
	decrease.Block = true
}
