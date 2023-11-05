// Package cybercat 云养猫
package cybercat

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	zbmath "github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/fumiama/jieba/util/helper"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex(`^(喵喵|猫猫)(PK|pk)\s*\[CQ:at,qq=(\d+).*`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.MessageID
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		uidStr := strconv.FormatInt(ctx.Event.UserID, 10)
		if ctx.State["regex_matched"].([]string)[3] == uidStr {
			ctx.SendChain(message.Reply(id), message.Text("猫猫歪头看着你表示咄咄怪事哦~"))
			return
		}
		userInfo, err := catdata.find(gidStr, uidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if userInfo == (catInfo{}) || userInfo.Name == "" {
			ctx.SendChain(message.Reply(id), message.Text("铲屎官你还没有属于你的主子喔,快去买一只吧!"))
			return
		}
		lastTime := time.Unix(userInfo.ArenaTime, 0)
		if time.Since(lastTime).Hours() < 24 {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "已经PK过了,让它休息休息吧"))
			return
		}
		duelStr := ctx.State["regex_matched"].([]string)[3]
		duelInfo, err := catdata.find(gidStr, duelStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if duelInfo == (catInfo{}) || duelInfo.Name == "" {
			ctx.SendChain(message.Reply(id), message.Text("他还没有属于他的猫猫,无法PK"))
			return
		}
		lastTime = time.Unix(duelInfo.ArenaTime, 0)
		if time.Since(lastTime).Hours() < 24 {
			ctx.SendChain(message.Reply(id), message.Text(ctx.CardOrNickName(duelInfo.User), "的", duelInfo.Name, "已经PK过了,让它休息休息吧"))
			return
		}
		/***************************************************************/
		ctx.SendChain(message.Text("等待对方回应。(发送“取消”撤回PK)\n请对方发送“去吧猫猫”接受PK或“拒绝”结束PK"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule("^(去吧猫猫|取消|拒绝)$"), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(zbmath.Str2Int64(duelStr), userInfo.User)).Repeat()
		defer cancel()
		approve := false
		over := time.NewTimer(60 * time.Second)
		for {
			select {
			case <-over.C:
				ctx.SendChain(message.Reply(id), message.Text("对方没回应,PK取消"))
				return
			case c := <-recv:
				over.Stop()
				switch {
				case c.Event.Message.String() == "拒绝" && c.Event.UserID == duelInfo.User:
					ctx.SendChain(message.Reply(id), message.Text("对方拒绝了你的PK"))
					return
				case c.Event.Message.String() == "取消" && c.Event.UserID == userInfo.User:
					ctx.SendChain(message.Reply(id), message.Text("你取消了PK"))
					return
				case c.Event.Message.String() == "去吧猫猫" && c.Event.UserID == duelInfo.User:
					approve = true
				}
			}
			if approve {
				break
			}
		}
		/***************************************************************/
		now := time.Now().Unix()
		winer := userInfo
		loser := duelInfo
		/***************************************************************/
		mood := false
		switch {
		case userInfo.Satiety > 50 && rand.Intn(100) > zbmath.Max(userInfo.Mood, 80):
			mood = true
			winer = duelInfo
			loser = userInfo
		case duelInfo.Satiety > 50 && rand.Intn(100) > zbmath.Max(duelInfo.Mood, 80):
			mood = true
		}
		if mood {
			ctx.SendChain(message.Text(loser.Name, "好像并没有心情PK\n", winer.Name, "获得了比赛胜利"))
			money := 10 + rand.Intn(int(winer.Weight))
			if wallet.InsertWalletOf(winer.User, money) == nil {
				ctx.SendChain(message.At(winer.User), message.Text("你家的喵喵为你赢得了", money))
			}
			winer.ArenaTime = now
			err = catdata.insert(gidStr, winer)
			if err == nil {
				loser.ArenaTime = now
				err = catdata.insert(gidStr, loser)
			}
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
			return
		}
		/***************************************************************/
		winLine := math.Min(userInfo.Weight, duelInfo.Weight)
		weightLine := (userInfo.Weight + duelInfo.Weight) * rand.Float64()
		fatLine := false
		if winLine > weightLine-winLine*0.1 && winLine < weightLine+winLine*0.1 {
			fatLine = true
		}
		if fatLine {
			ctx.SendChain(message.Reply(id), message.Text(duelInfo.Name, "和", userInfo.Name, "之间并没有PK的意愿呢\nPK结束"))
			userInfo.ArenaTime = now
			err = catdata.insert(gidStr, userInfo)
			if err == nil {
				duelInfo.ArenaTime = now
				err = catdata.insert(gidStr, duelInfo)
			}
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
		}
		/***************************************************************/
		winer, loser = pkweight(userInfo, duelInfo)
		messageText := make(message.Message, 0, 3)
		if rand.Intn(2) == 0 {
			messageText = append(messageText, message.Text("天啊,",
				strconv.FormatFloat(winer.Weight, 'f', 2, 64), "kg的", winer.Name,
				"完美的借力打力,将", strconv.FormatFloat(loser.Weight, 'f', 2, 64), "kg的", loser.Name, "打趴下了"))
		} else {
			messageText = append(messageText, message.Text("精彩!", strconv.FormatFloat(winer.Weight, 'f', 2, 64), "kg的", winer.Name,
				"利用了PK地形,让", strconv.FormatFloat(loser.Weight, 'f', 2, 64), "kg的", loser.Name, "认输了"))
		}
		money := 10 + rand.Intn(int(winer.Weight))
		if wallet.InsertWalletOf(winer.User, money) == nil {
			messageText = append(messageText, message.Text("\n"), message.At(winer.User), message.Text("\n", winer.Name, "为你赢得了", money))
		} else {
			messageText = append(messageText, message.Text("\n"), message.At(winer.User), message.Text("\n", winer.Name, "受伤了,所赚的钱全拿来疗伤了"))
		}
		if rand.Float64()*100 < math.Max(20, loser.Weight) {
			loser.Weight -= math.Min(1, loser.Weight/10) * rand.Float64()
			messageText = append(messageText, message.Text("\n"), message.At(loser.User),
				message.Text("\n", loser.Name, "在PK中受伤了\n在医疗中心治愈过程中体重降低至", strconv.FormatFloat(loser.Weight, 'f', 2, 64)))
		}
		userInfo.ArenaTime = time.Now().Unix()
		err = catdata.insert(gidStr, winer)
		if err == nil {
			duelInfo.ArenaTime = time.Now().Unix()
			err = catdata.insert(gidStr, loser)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
		}
		ctx.Send(messageText)
	})
	engine.OnFullMatchGroup([]string{"猫猫排行榜", "喵喵排行榜"}, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		gidStr := "group" + strconv.FormatInt(ctx.Event.GroupID, 10)
		infoList, err := catdata.getGroupdata(gidStr)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if len(infoList) == 0 {
			ctx.SendChain(message.Text("没有人养猫哦"))
			return
		}
		messageText := make([]string, 0, 10)
		for i, info := range infoList {
			if i > 9 {
				break
			}
			messageText = append(messageText, []string{
				strconv.Itoa(i+1) + "." + info.Name + "(" + info.Type + ")",
				"体重：" + strconv.FormatFloat(info.Weight, 'f', 2, 64) + "斤",
				"主人:" + ctx.CardOrNickName(info.User), "--------------------",
			}...)
		}
		textPic, err := text.RenderToBase64(strings.Join(messageText, "\n"), text.BoldFontFile, 1080, 50)
		if err != nil {
			return
		}
		ctx.SendChain(message.Image("base64://" + helper.BytesToString(textPic)))
	})
}

// PK参数权重
// 系数= (属性 - 基准属性)/100
// R = (体重 - 基准属性)*0.05 + (心情 - 心情)*K*0.4 + (饱食度 - 饱食度)*R*0.4
/**********************体重绝对优势******************************
	// 纯比体重
	// 17,100,100 Vs 130,100,100
	R1 = (17 - 50) * 0.05 = -1.65
	R2 = (130 - 50) * 0.05 = 4
	R2 > R1,R2赢
	// 如果心情好
	// 17,80,100 Vs 130,40,100
	R1 = (17 - 50) * 0.05 + (80 - 40)*(80 - 50)/100 * 0.4 = -1.65 + 4.8 = 3.15
	R2 = (130 - 50) * 0.05 + (40 - 80)*(40 - 50)/100 * 0.4 = 4 + 1.6 = 5.6
	R2 > R1,R2赢
	// 如果肚子饿
	// 17,80,80 Vs 130,40,40
	R1 = (17 - 50) * 0.05 + (80 - 40)*(80 - 50)/100 * 0.4 + (80 - 40)*(80 - 50)/100 * 0.4 = -1.65 + 4.8 + 4.8 = 7.95
	R2 = (130 - 50) * 0.05 + (40 - 80)*(40 - 50)/100 * 0.4 + (80 - 40)*(80 - 50)/100 * 0.4 = 4 + 1.6 + 1.6 = 7.2
	R1 > R2,R1赢
**********************体重均衡******************************
	// 纯比体重
	// 30,100,100 Vs 60,100,100
	R1 = (30 - 50) * 0.05 = -0.1
	R2 = (60 - 50) * 0.05 = 0.5
	R2 > R1,R2赢
	// 如果心情好
	// 30,80,100 Vs 60,40,100
	R1 = (30 - 50) * 0.05 + (80 - 40)*(80 - 50)/100 * 0.4 = -0.1 + 4.8 = 4.7
	R2 = (60 - 50) * 0.05 + (40 - 80)*(40 - 50)/100 * 0.4 = 0.5 + 1.6 = 2.3
	R1 > R2,R1赢
	// 如果肚子饿
	// 30,80,40 Vs 130,40,80
	R1 = (30 - 50) * 0.05 + (80 - 40)*(80 - 50)/100 * 0.4 + (80 - 40)*(80 - 50)/100 * 0.4 = 4.7 + 1.6= 6.3
	R2 = (130 - 50) * 0.05 + (40 - 80)*(40 - 50)/100 * 0.4 + (80 - 40)*(80 - 50)/100 * 0.4 = 2.3 + 4.8 = 7.1
	R2 > R1,R2赢
*/
func pkweight(player1, player2 catInfo) (winer, loser catInfo) {
	weightOfplayer1 := (player1.Weight-50)*0.05 +
		float64((player1.Mood-player2.Mood)*(player1.Mood-50))*0.4 +
		(player1.Satiety-player2.Satiety)*(player1.Satiety-50)*0.4
	weightOfplayer2 := (player2.Weight-50)*0.05 +
		float64((player2.Mood-player1.Mood)*(player2.Mood-50))*0.4 +
		(player2.Satiety-player1.Satiety)*(player2.Satiety-50)*0.4
	if weightOfplayer1 > weightOfplayer2 {
		return player1, player2
	}
	return player2, player1
}
