package chat

import (
	"fmt"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	zero.OnNotice(zero.OnlyToMe).SetBlock(true).SetPriority(0).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType != "notify" {
				return
			}
			bot, _ := strconv.ParseInt(zero.BotConfig.SelfID, 10, 64)
			if ctx.Event.UserID == bot {
				time.Sleep(time.Second * 1)
				ctx.Send("请不要戳我 >_<")
				ctx.Send(fmt.Sprintf("[CQ:poke,qq=%d]", ctx.Event.OperatorID))
			}
		})
}
