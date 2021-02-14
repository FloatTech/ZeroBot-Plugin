package setutime

import (
	"fmt"

	"bot/music/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(musicSelector{}) // 注册插件
}

type musicSelector struct{} // musicSelector 点歌

func (_ musicSelector) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "MusicSelector",
		Version:    "0.0.1",
		Details:    "点歌",
	}
}

func (_ musicSelector) Start() { // 插件主体
	// TODO 根据PID搜图
	zero.OnRegex(`点歌(.*)`).SetBlock(true).SetPriority(50).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			music, err := utils.CloudMusic(state["regex_matched"].([]string)[1])
			if err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			// TODO 发送搜索结果
			zero.Send(
				event,
				fmt.Sprintf(
					"[CQ:music,type=%s,url=%s,audio=%s,title=%s,content=%s,image=%s]",
					music.Type,
					music.Url,
					music.Audio,
					music.Title,
					music.Content,
					music.Image,
				),
			)
			return zero.FinishResponse
		})
}
