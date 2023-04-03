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
	engine.OnRegex(`^买(.*猫)$`, zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		if now := time.Now().Hour(); now >= 6 && now <= 20 {
			return true
		}
		ctx.SendChain(message.Text("猫店已经关门了,早上六点后再来吧"))
		return false
	}, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		/*******************************************************/
		if userInfo != (catInfo{}) && userInfo.Name != "" {
			id = ctx.SendChain(message.Reply(id), message.Text("你居然背着你家喵喵出来找小三!"))
			if rand.Intn(100) < 20 {
				process.SleepAbout1sTo2s()
				if rand.Intn(50) == 30 {
					if catdata.del(gidStr, uidStr) == nil {
						ctx.SendChain(message.Reply(id), message.Text("喔,天啊!你家喵喵带着所有猫粮离家出走了!\n你失去了所有!"))
					}
				} else if catdata.delcat(gidStr, uidStr) == nil {
					ctx.SendChain(message.Reply(id), message.Text("喔,天啊!你家喵喵离家出走了!\n你失去了喵喵!"))
				}
			}
			return
		}
		/*******************************************************/
		lastTime := time.Unix(userInfo.LastTime, 0).Day()
		if lastTime != time.Now().Day() {
			userInfo.Work = 0
			userInfo.LastTime = 0
		}
		/*******************************************************/
		userInfo.User = ctx.Event.UserID
		typeOfcat := ctx.State["regex_matched"].([]string)[1]
		if userInfo.LastTime != 0 && typeOfcat == "猫" {
			ctx.SendChain(message.Reply(id), message.Text("抱歉,一天只能去猫店一次"))
			return
		} else if userInfo.Work > 1 {
			ctx.SendChain(message.Reply(id), message.Text("抱歉,一天只能选购两次"))
			return
		}
		/*******************************************************/
		if typeOfcat == "猫" {
			userInfo.LastTime = time.Now().Unix()
		} else {
			userInfo.Work++
		}
		if err = catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		/*******************************************************/
		if wallet.GetWalletOf(ctx.Event.UserID) < 100 {
			ctx.SendChain(message.Reply(id), message.Text("一只喵喵官方售价100哦;\n你身上没有足够的钱,快去赚钱吧~"))
			return
		}
		messageText := make(message.Message, 0, 3)
		messageText = append(messageText, message.Reply(id))
		money := 100
		if rand.Intn(10) == 5 {
			money = rand.Intn(50) + 50
			messageText = append(messageText, message.Text("你前往的喵喵店时发现正好有活动,\n一只喵喵现在只需要", money, "\n------------------------\n"))
		}
		/*******************************************************/
		if typeOfcat == "猫" {
			nameMap := make([]string, 0, len(catBreeds))
			for zhName := range catBreeds {
				nameMap = append(nameMap, zhName)
			}
			if rand.Intn(100) >= 90 {
				nameMap = append(nameMap, "猫娘")
			}
			nameList := make([]int, 0, 5)
			for i := 0; i < 5; i++ {
				nameList = append(nameList, rand.Intn(len(nameMap)))
			}
			messageText = append(messageText, message.Text("当前猫店售卖以下几只猫:",
				"\n1.", nameMap[nameList[0]],
				"\n2.", nameMap[nameList[1]],
				"\n3.", nameMap[nameList[2]],
				"\n4.", nameMap[nameList[3]],
				"\n5.", nameMap[nameList[4]],
				"\n请发送对应序号进行购买或“取消”取消购买"))
			ctx.Send(messageText)
			typeRecv, typeCancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule("^([1-5]|取消)$"), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(userInfo.User)).Repeat()
			defer typeCancel()
			approve := false
			over := time.NewTimer(60 * time.Second)
			for {
				select {
				case <-over.C:
					ctx.SendChain(message.Reply(id), message.Text("你考虑的时间太长了,喵喵店都关门了!下次再来买哦~"))
					return
				case c := <-typeRecv:
					over.Stop()
					switch c.Event.Message.String() {
					case "取消":
						ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("欢迎你的下次光临"))
						return
					default:
						index, _ := strconv.Atoi(c.Event.Message.String())
						typeOfcat = nameMap[nameList[index-1]]
						approve = true
					}
				}
				if approve {
					break
				}
			}
		}
		/*******************************************************/
		picurl, _ := getPicByBreed(catBreeds[typeOfcat])
		satiety := 90 * rand.Float64() // 饱食度
		mood := 50 + rand.Intn(50)     // 心情
		weight := 2 + 8*rand.Float64() // 体重
		/*******************************************************/
		messageText = message.Message{message.Reply(id)}
		messageText = append(messageText, message.Text("经过询问后得知它当前的信息为:\n"),
			message.Image(picurl),
			message.Text("品种: ", typeOfcat,
				"\n当前饱食度: ", strconv.FormatFloat(satiety, 'f', 0, 64),
				"\n当前心情: ", mood,
				"\n当前体重: ", strconv.FormatFloat(weight, 'f', 2, 64),
				"\n\n你想要买这只猫猫,\n请发送“叫xxx”为它取个名字吧~\n(发送“否”取消购买)"))
		ctx.Send(messageText)
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule("^(叫.*|否)$"), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(userInfo.User)).Repeat()
		defer cancel()
		approve := false
		over := time.NewTimer(90 * time.Second)
		for {
			select {
			case <-over.C:
				ctx.SendChain(message.Reply(id), message.Text("你考虑的时间太长了,喵喵店都关门了!下次再来买哦~"))
				return
			case c := <-recv:
				over.Stop()
				switch c.Event.Message.String() {
				case "否":
					ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("欢迎你的下次光临"))
					return
				default:
					id = c.Event.MessageID
					userInfo.Name = strings.ReplaceAll(c.Event.Message.String(), "叫", "")
					if userInfo.Name == "" || len(userInfo.Name) > 15 {
						over.Reset(90 * time.Second)
						ctx.SendChain(message.Reply(id), message.Text("请输入正确的猫名"))
						continue
					}
					approve = true
				}
			}
			if approve {
				break
			}
		}
		messageText = message.Message{message.Reply(id)}
		if rand.Intn(5) == 1 {
			mood += rand.Intn(30)
			if mood > 100 {
				mood = 100
			}
			messageText = append(messageText, message.Text("这只喵喵好像很喜欢这个名字,\n"))
		}
		userInfo.Type = typeOfcat
		userInfo.Satiety = satiety
		userInfo.Mood = mood
		userInfo.Weight = weight
		userInfo.LastTime = 0
		userInfo.Work = 0
		userInfo.Picurl = picurl
		if err = wallet.InsertWalletOf(ctx.Event.UserID, -money); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if err = catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		messageText = append(messageText, message.Text("恭喜你买了一只喵喵"))
		ctx.Send(messageText)
	})
	engine.OnRegex(`^买((\d+)袋)?猫粮$`, zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		if now := time.Now().Hour(); now >= 6 && now <= 20 {
			return true
		}
		ctx.SendChain(message.Text("猫店已经关门了,早上六点后再来吧"))
		return false
	}, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		/*******************************************************/
		mun := 1.0
		if ctx.State["regex_matched"].([]string)[2] != "" {
			mun, _ = strconv.ParseFloat(ctx.State["regex_matched"].([]string)[2], 64)
			if mun > 100 {
				ctx.SendChain(message.Reply(id), message.Text("猫猫店库存只有100袋,无法供给"))
				return
			}
			if mun < 1 {
				ctx.SendChain(message.Reply(id), message.Text("请输入正确的数量"))
				return
			}
		}
		/*******************************************************/
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if userInfo.Food > 50 {
			ctx.SendChain(message.Reply(id), message.Text("你家的猫粮已经装满仓库(上限50斤)了!"))
			return
		}
		/*******************************************************/
		if wallet.GetWalletOf(ctx.Event.UserID) < 10*int(mun) {
			ctx.SendChain(message.Reply(id), message.Text("一袋猫粮官方售价10哦;\n你身上没有足够的钱,快去赚钱吧~"))
			return
		}
		messageText := make(message.Message, 0, 3)
		messageText = append(messageText, message.Reply(id))
		userInfo.User = ctx.Event.UserID
		foodmoney := 10
		if rand.Intn(10) < 3 {
			foodmoney = rand.Intn(5) + 5
			messageText = append(messageText, message.Text("你前往的喵喵店时发现正好有活动,\n一袋猫粮现在只需要", foodmoney, ";\n"))
		}
		foodmoney *= int(mun)
		userInfo.Food += 5 * mun
		if wallet.InsertWalletOf(ctx.Event.UserID, -foodmoney) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if err = catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		messageText = append(messageText, message.Text("你购买了", mun, "袋,共计", foodmoney, "\n当前猫粮有", strconv.FormatFloat(userInfo.Food, 'f', 2, 64), "斤"))
		ctx.Send(messageText)
	})
	engine.OnPrefixGroup([]string{"喵喵改名叫", "猫猫改名叫"}, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if userInfo == (catInfo{}) || userInfo.Name == "" {
			ctx.SendChain(message.Reply(id), message.Text("铲屎官你还没有属于你的主子喔,快去买一只吧!"))
			return
		}
		newName := strings.TrimSpace(ctx.State["args"].(string))
		switch {
		case newName == "":
			userInfo.Name = userInfo.Type
		case len(newName) > 6*3:
			ctx.SendChain(message.Reply(id), message.Text("请输入正确的名字"))
			return
		default:
			userInfo.Name = newName
		}
		if err = catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text("修改成功"))
	})
	engine.OnPrefix("上传猫猫照片", zero.OnlyGroup, getdb, func(ctx *zero.Ctx) bool {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		if userInfo == (catInfo{}) || userInfo.Name == "" {
			ctx.SendChain(message.Reply(id), message.Text("铲屎官你还没有属于你的主子喔,快去买一只吧!"))
			return false
		}
		if userInfo.Type != "猫娘" {
			ctx.SendChain(message.Reply(id), message.Text("只有猫娘才能资格更换图片喔"))
			return false
		}
		return zero.MustProvidePicture(ctx)
	}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		userInfo, _ := catdata.find(gidStr, uidStr)
		userInfo.Picurl = ctx.State["image_url"].([]string)[0]
		if err := catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
}
