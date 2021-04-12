package chat

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

var poke = rate.NewManager(time.Minute*5, 8) // 戳一戳

func init() { // 插件主体
	// 戳一戳
	zero.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			switch {
			case poke.Load(ctx.Event.UserID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				time.Sleep(time.Second * 1)
				ctx.Send("请不要戳我 >_<")
			case poke.Load(ctx.Event.UserID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				ctx.Send("喂(#`O′) 戳我干嘛！")
			default:
				// 频繁触发，不回复
			}
			return
		})
}
