// Package gamesystem 基于zbp的猜歌插件
package gamesystem

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	// 注册游戏信息
	if err := register("猜硬币", gameinfo{
		Command: "- 创建猜银币\n" +
			"- [加入|开始]游戏\n" +
			"- 我猜x个正面\n" +
			"- 开始投币",
		Help: "每个人宣言银币正面数量后,掷出参游人数的银币",
		Rewards: "正面与宣言的数量相同的场合获得 正面数*10 ATRI币\n" +
			"正面与宣言的数量相差2以内的场合获得 正面数*5 ATRI币\n" +
			"其他的的场合失去 10 ATRI币",
	}); err != nil {
		panic(err)
	}
	engine.OnFullMatch("创建猜银币", zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		err := whichGameRoomIn("猜硬币", ctx.Event.GroupID)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 结束后关闭房间
		defer whichGameRoomOut("猜硬币", ctx.Event.GroupID)
		uid := ctx.Event.UserID
		userScore := wallet.GetWalletOf(uid)
		if userScore < 10 {
			ctx.SendChain(message.Text("你的ATRI币不足以满足该游戏"))
			return
		}
		// 等待对方响应
		ctx.SendChain(message.Text("你开启了猜银币游戏。\n其他人可发送“加入游戏”加入游戏或你发送“开始游戏”开始游戏"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.FullMatchRule("加入游戏", "开始游戏"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		answer := ""
		var answerid int64
		var duel = make(map[int64]int, 2)
		uidlist := []int64{uid}
		duel[uid] = -1
		wait := time.NewTimer(120 * time.Second)
		for {
			select {
			case <-wait.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("时间超时,游戏取消")))
				return
			case c := <-recv:
				answer = c.Event.Message.String()
				answerid = c.Event.UserID
				if answer == "加入游戏" {
					_, ok := duel[answerid]
					if ok {
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("你已经加入了游戏")))
					} else {
						challengScore := wallet.GetWalletOf(answerid)
						if challengScore < 10 {
							ctx.SendChain(message.Text("你的ATRI币不足以满足该游戏"))
							return
						}
						duel[answerid] = -1
						uidlist = append(uidlist, answerid)
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("成功加入游戏,等待开房人开始游戏")))
					}
				}
			}
			if answer == "开始游戏" && answerid == uid {
				break
			}
		}
		wait = time.NewTimer(120 * time.Second)
		ctx.SendChain(message.Text("游戏开始,请参游人员宣言正面硬币数量或开房人发送“开始投币”开始投币"))
		recv, cancel = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^\d{1,`+strconv.Itoa(len(duel))+`}$|开始投币`), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(uidlist...)).Repeat()
		defer cancel()
		mun := 0
		for {
			select {
			case <-wait.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("等待超时,游戏取消\n咕之人扣除 6 ATRI币")))
				for uid, guess := range duel {
					if guess == -1 {
						err := wallet.InsertWalletOf(uid, -6)
						if err != nil {
							ctx.SendChain(message.At(uid), message.Text(serviceErr, err))
						}
					}
				}
				return
			case c := <-recv:
				eventID := c.Event.UserID
				answer = c.Event.Message.String()
				if answer != "开始投币" && duel[eventID] == -1 {
					mun++
					guess, _ := strconv.Atoi(answer)
					duel[eventID] = guess
					ctx.SendChain(message.Text("以记录你宣言的数目:", guess))
				}
			}
			if answer == "开始投币" {
				if mun >= len(duel) {
					break
				}
				ctx.SendChain(message.Text("还有人没有宣言数量喔"))
			}
		}
		positive := 0
		result := "\n"
		for i := 0; i < len(duel); i++ {
			switch rand.Intn(2) {
			case 0:
				result += "反 "
			case 1:
				positive++
				result += "正 "
			}
		}
		ctx.SendChain(message.Text("OK,我要开始投掷银币了～"))
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Text("一共投掷了", len(duel), "枚银币,其中正面的有", positive, "枚正面。\n具体结果如下", result))
		//数据结算
		for uid, guess := range duel {
			switch {
			case guess == positive:
				err := wallet.InsertWalletOf(uid, positive*10)
				if err != nil {
					ctx.SendChain(message.At(uid), message.Text(serviceErr, err))
				}
			case int(math.Abs(guess-positive)) <= 2:
				err := wallet.InsertWalletOf(uid, positive*5)
				if err != nil {
					ctx.SendChain(message.At(uid), message.Text(serviceErr, err))
				}
			default:
				err := wallet.InsertWalletOf(uid, -10)
				if err != nil {
					ctx.SendChain(message.At(uid), message.Text(serviceErr, err))
				}
			}
		}
		ctx.SendChain(message.Text("宣言的数量与正面数相同的玩家将获得 ", positive*10, "ATRI币\n"+
			"宣言的数量与正面数相差2以内的将获得 ", positive*5, "ATRI币\n其他玩家失去 10 ATRI币\n谢谢游玩。"))
	})
}
