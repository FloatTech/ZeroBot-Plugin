// Package cybercat 云养猫
package cybercat

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnFullMatch("买猫", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if userInfo != (catinfo{}) && userInfo.Name != "" {
			id = ctx.SendChain(message.Reply(id), message.Text("你居然背着你家喵喵出来找小三!"))
			if rand.Intn(100) == 50 {
				process.SleepAbout1sTo2s()
				if rand.Intn(100) == 1 && catdata.del(gidStr, uidStr) == nil {
					ctx.SendChain(message.Reply(id), message.Text("喔,天啊!你家喵喵带着所有猫粮离家出走了!\n你失去了所有!"))
				} else {
					userInfo = catinfo{
						User: ctx.Event.UserID,
						Food: userInfo.Food,
					}
					err := catdata.insert(gidStr, userInfo)
					if err != nil {
						ctx.SendChain(message.Text("[ERROR]:", err))
						return
					}
					ctx.SendChain(message.Reply(id), message.Text("喔,天啊!你家喵喵离家出走了!\n你失去了喵喵!"))
					return
				}
			}
			return
		}
		userInfo.User = ctx.Event.UserID
		money := wallet.GetWalletOf(ctx.Event.UserID)
		if money < 100 {
			ctx.SendChain(message.Reply(id), message.Text("一只喵喵官方售价100哦;\n你身上没有足够的钱,快去赚钱吧~"))
			return
		}
		money = 100
		messageText := ""
		if rand.Intn(10) == 5 {
			money = rand.Intn(50) + 50
			messageText = "你前往的喵喵店时发现正好有活动,\n一只喵喵现在只需要" + strconv.Itoa(money) + ";\n"
		}
		// 随机属性生成
		typeOfcat := catType[rand.Intn(len(catType))]
		satiety := rand.Intn(90)
		mood := rand.Intn(100)
		weight := rand.Intn(10) + 1

		id = ctx.SendChain(message.Reply(id), message.Text(messageText, "你在喵喵店看到了一只喵喵,经过询问后得知他当前的信息为:",
			"\n品种: "+typeOfcat, "\n当前饱食度: "+strconv.Itoa(satiety), "\n当前心情: "+strconv.Itoa(mood), "\n当前体重: "+strconv.Itoa(weight),
			"\n你是否想要买这只喵喵呢?(回答“是/否”)"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule("^(是|否)$"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		approve := false
		over := time.NewTimer(60 * time.Second)
		for {
			select {
			case <-over.C:
				ctx.SendChain(message.Reply(id), message.Text("你考虑的时间太长了,喵喵店都关门了!下次再来买哦~"))
				//cancel()
				return
			case c := <-recv:
				switch c.Event.Message.String() {
				case "否":
					ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("欢迎你的下次光临"))
				default:
					approve = true
				}
			}
			if approve {
				//cancel()
				break
			}
		}
		ctx.SendChain(message.Reply(id), message.Text("喵喵对你喵喵了两句,貌似是想让你给它取名呢!\n请发送“喵喵叫xxx”给它取名吧~"))
		nameRecv, nameCancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule("^喵喵叫.*"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer nameCancel()
		approve = false
		over = time.NewTimer(30 * time.Second)
		for {
			select {
			case <-over.C:
				ctx.SendChain(message.Reply(id), message.Text("你考虑的时间太长了!可以发送“喵喵改名叫xxx”进行再次改名喔"))
				userInfo.Name = typeOfcat
				approve = true
			case c := <-nameRecv:
				id = c.Event.MessageID
				userInfo.Name = strings.ReplaceAll(c.Event.Message.String(), "喵喵叫", "")
				if rand.Intn(5) == 1 {
					mood += rand.Intn(5)
					if mood > 100 {
						mood = 100
					}
					ctx.SendChain(message.Reply(id), message.Text("这只喵喵好像很喜欢这个名字"))
				}
				approve = true
			}
			if approve {
				break
			}
		}
		userInfo.Type = typeOfcat
		userInfo.Satiety = satiety
		userInfo.Mood = mood
		userInfo.Weight = weight
		if wallet.InsertWalletOf(ctx.Event.UserID, -money) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("恭喜你买了一只喵喵"))
	})
	engine.OnRegex(`^买((\d+)袋)?猫粮$`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		mun := 1
		if ctx.State["regex_matched"].([]string)[2] != "" {
			mun, _ = strconv.Atoi(ctx.State["regex_matched"].([]string)[2])
		}
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		userInfo.User = ctx.Event.UserID
		money := wallet.GetWalletOf(ctx.Event.UserID)
		if money < 10 {
			ctx.SendChain(message.Reply(id), message.Text("一盒猫粮官方售价10哦;\n你身上没有足够的钱,快去赚钱吧~"))
			return
		}
		foodmoney := 10
		messageText := ""
		if rand.Intn(10) < 3 {
			foodmoney = rand.Intn(5) + 5
			messageText = "你前往的喵喵店时发现正好有活动,\n一袋猫粮现在只需要" + strconv.Itoa(money) + ";\n"
		}
		foodmoney *= mun
		if money < foodmoney {
			ctx.SendChain(message.Reply(id), message.Text("你身上没有足够的钱买这么多猫粮,快去赚钱吧~"))
			return
		}
		userInfo.Food = 5 * mun
		if wallet.InsertWalletOf(ctx.Event.UserID, foodmoney) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text(messageText, "你购买了", mun, "袋,共计", money))
	})
}
