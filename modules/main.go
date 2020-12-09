package modules

import (
	"fmt"
	"gm/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var Conf = &utils.YamlConfig{}

func init() {
	zero.RegisterPlugin(menu{})
}

type menu struct{}

func (menu) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "menu",
		Version:    "0.0.1",
		Details:    "菜单",
	}
}

func (menu) Start() { // 插件主体
	gmmenu := zero.OnRegex("^群管系统$").
		Handle(
			func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
				menuText := `====群管====
- 禁言@QQ 1
- 解除禁言 @QQ
- 我要自闭 1分钟
- 开启全员禁言
- 解除全员禁言
- 升为管理@QQ
- 取消管理@QQ
- 修改名片@QQ XXX
- 修改头衔@QQ XXX
- 申请头衔 XXX
- 群聊转发 1234 XXX
- 私聊转发 0000 XXX`
				zero.Send(event, fmt.Sprintf("%v", menuText))
				return zero.SuccessResponse
			},
		)
	gmmenu.Priority = 60
	gmmenu.Block = true
}
