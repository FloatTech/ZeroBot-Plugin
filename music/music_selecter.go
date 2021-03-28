package setutime

import (
	"fmt"

	"bot/music/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	zero.OnRegex(`^点歌(.*)$`).SetBlock(true).SetPriority(50).
		Handle(func(ctx *zero.Ctx) {
			music, err := utils.CloudMusic(ctx.State["regex_matched"].([]string)[1])
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 发送搜索结果
			ctx.Send(
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
			return
		})
}
