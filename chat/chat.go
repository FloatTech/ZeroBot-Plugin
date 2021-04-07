package chat

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	// 戳一戳
	zero.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).SetPriority(0).
		Handle(func(ctx *zero.Ctx) {
			time.Sleep(time.Second * 1)
			ctx.Send("请不要戳我 >_<")
			return
		})
}
