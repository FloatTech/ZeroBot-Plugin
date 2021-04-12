package chat

import (
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var poke = rate.NewManager(time.Minute*5, 8) // 戳一戳

func init() { // 插件主体
	var NICKNAME = zero.BotConfig.NickName[0]
	// 被喊名字
	zero.OnFullMatchGroup(zero.BotConfig.NickName).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					NICKNAME + "在此，有何贵干~",
					"(っ●ω●)っ在~",
					"这里是" + NICKNAME + "(っ●ω●)っ",
					NICKNAME + "不在呢~",
				}[rand.Intn(4)],
			))
		})
	// 戳一戳
	zero.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			switch {
			case poke.Load(ctx.Event.UserID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("请不要戳", NICKNAME, " >_<"))
			case poke.Load(ctx.Event.UserID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("喂(#`O′) 戳", NICKNAME, "干嘛！"))
			default:
				// 频繁触发，不回复
			}
			return
		})
}
