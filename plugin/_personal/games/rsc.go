// Package gamesystem ...
package gamesystem

import (
	"math/rand"

	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 载入游戏系统
	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/games/gamesystem"
)

var point = map[string]int{
	"石头": 1,
	"剪刀": 2,
	"布":  3,
}

func init() {
	// 注册游戏信息
	engine, gameManager, err := gamesystem.Register("石头剪刀布", &gamesystem.GameInfo{
		Command: "- @bot[石头｜剪刀｜布]",
		Help:    "和机器人进行猜拳,如果机器人开心了会得到ATRI币",
		Rewards: "奖励范围在0~10之间",
	})
	if err != nil {
		panic(err)
	}
	engine.OnFullMatchGroup([]string{"石头", "剪刀", "布"}, zero.OnlyToMe, func(ctx *zero.Ctx) bool {
		if gameManager.PlayIn(ctx.Event.GroupID) {
			return true
		}
		ctx.SendChain(message.Text("游戏已下架,无法游玩"))
		return false
	}).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			botchoose := 1 + rand.Intn(3)
			switch botchoose {
			case 1:
				ctx.SendChain(message.Text("石头"))
			case 2:
				ctx.SendChain(message.Text("剪刀"))
			case 3:
				ctx.SendChain(message.Text("布"))
			}
			model := ctx.State["matched"].(string)
			result := point[model] - botchoose

			// 如果是石头和布的比较，比较值正负取反
			if int(math.Abs(result)) == 2 {
				result = -(result)
			}
			switch {
			case result < 0:
				ctx.SendChain(message.Text("可恶,你赢了"))
			case result > 0:
				if rand.Intn(5) == 1 {
					money := rand.Intn(11)
					if money > 0 {
						err := wallet.InsertWalletOf(ctx.Event.UserID, money)
						if err == nil {
							ctx.SendChain(message.Text("哈哈,你输了。嗯!~今天运气不错,我很高兴,奖励你 ", money, " 枚ATRI币吧"))
							return
						}
					}
				}
				ctx.SendChain(message.Text("哈哈,你输了"))
			default:
				if rand.Intn(10) == 1 {
					money := rand.Intn(11)
					if money > 0 {
						err := wallet.InsertWalletOf(ctx.Event.UserID, money)
						if err == nil {
							ctx.SendChain(message.Text("你实力不错,我很欣赏你,奖励你 ", money, " 枚ATRI币吧"))
							return
						}
					}
				}
				ctx.SendChain(message.Text("实力可以啊,希望下次再来和我玩"))
			}
		})
}
