// Package cybercat 云养猫
package cybercat

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	zbmath "github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex(`^(喵喵|猫猫)(PK|pk)\s*\[CQ:at,qq=(\d+).*`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
		ctx.SendChain(message.Text("等待对方回应。\n(请对方回复“去吧猫猫”接受PK)"))
		duelID, _ := strconv.ParseInt(duelStr, 10, 64)
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule("^去吧猫猫$"), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(duelID)).Repeat()
		defer cancel()
		approve := false
		over := time.NewTimer(60 * time.Second)
		for {
			select {
			case <-over.C:
				ctx.SendChain(message.Reply(id), message.Text("对方没回应,PK取消"))
				// cancel()
				return
			case <-recv:
				over.Stop()
				// break
				approve = true
			}
			if approve {
				break
			}
		}
		/***************************************************************/
		if userInfo.Satiety < 50 && rand.Intn(100) > zbmath.Max(userInfo.Mood, 80) {
			ctx.SendChain(message.Text(userInfo.Name, "好像并没有心情PK\n", duelInfo.Name, "获得了比赛胜利"))
			money := 10 + rand.Intn(duelInfo.Mood)
			if wallet.InsertWalletOf(duelID, money) == nil {
				ctx.SendChain(message.At(duelID), message.Text("你家的喵喵为你赢得了", money))
			}
			userInfo.ArenaTime = time.Now().Unix()
			err = catdata.insert(gidStr, userInfo)
			if err == nil {
				userInfo.ArenaTime = time.Now().Unix()
				err = catdata.insert(gidStr, userInfo)
			}
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
			return
		}
		if duelInfo.Satiety < 50 && rand.Intn(100) > zbmath.Max(duelInfo.Mood, 80) {
			ctx.SendChain(message.Text(duelInfo.Name, "好像并没有心情PK\n", userInfo.Name, "获得了比赛胜利"))
			money := 10 + rand.Intn(userInfo.Mood)
			if wallet.InsertWalletOf(userInfo.User, money) == nil {
				ctx.SendChain(message.At(userInfo.User), message.Text("你家的喵喵为你赢得了", money))
			}
			userInfo.ArenaTime = time.Now().Unix()
			err = catdata.insert(gidStr, userInfo)
			if err == nil {
				userInfo.ArenaTime = time.Now().Unix()
				err = catdata.insert(gidStr, userInfo)
			}
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
			}
			return
		}
		/***************************************************************/
		winLine := math.Min(userInfo.Weight, duelInfo.Weight)
		fat := false
		if winLine == duelInfo.Weight {
			fat = true // 判断用户的是否比对手的重
		}
		winerWeight := (userInfo.Weight + duelInfo.Weight) * rand.Float64()
		messageText := []message.MessageSegment{}
		switch {
		case fat && winerWeight <= (winLine-5): //重,但对面赢了
			messageText = append(messageText, message.Text("天啊,", duelInfo.Weight, "kg的", duelInfo.Name, "完美的借力打力,将", userInfo.Weight, "kg的", userInfo.Name, "打趴下了"))
			if rand.Float64()*100 < math.Max(20, userInfo.Weight) {
				userInfo.Weight -= math.Min(1, duelInfo.Weight/10) * rand.Float64()
				messageText = append(messageText, message.Text("\n"), message.At(userInfo.User), message.Text(userInfo.Name, "在PK中受伤了\n在医疗中心治愈过程中体重降低至", fmt.Sprintf("%1.2f", userInfo.Weight)))

			}
			money := 10 + rand.Intn(zbmath.Min(30, duelInfo.Mood))
			if wallet.InsertWalletOf(duelInfo.User, money) == nil {
				messageText = append(messageText, message.Text("\n"), message.At(duelInfo.User), message.Text(duelInfo.Name, "为你赢得了", money))
			}
		case fat && winerWeight >= (winLine+15): //重,且赢了
			messageText = append(messageText, message.Text(userInfo.Weight, "kg的", userInfo.Name, "以绝对的体型碾压住了", duelInfo.Weight, "kg的", duelInfo.Name))
			if rand.Float64()*100 < math.Min(20, duelInfo.Weight) {
				duelInfo.Weight -= math.Min(1, duelInfo.Weight/10) * rand.Float64()
				messageText = append(messageText, message.Text("\n"), message.At(duelInfo.User), message.Text(duelInfo.Name, "在PK中受伤了\n在医疗中心治愈过程中体重降低至", fmt.Sprintf("%1.2f", duelInfo.Weight)))

			}
			money := 10 + rand.Intn(zbmath.Min(30, userInfo.Mood))
			if wallet.InsertWalletOf(userInfo.User, money) == nil {
				messageText = append(messageText, message.Text("\n"), message.At(userInfo.User), message.Text(userInfo.Name, "为你赢得了", money))
			}
		case !fat && winerWeight <= (winLine-5): //轻,且赢了
			ctx.SendChain(message.Text("天啊,", userInfo.Weight, "kg的", userInfo.Name, "完美的借力打力,将", duelInfo.Weight, "kg的", duelInfo.Name, "打趴下了"))
			if rand.Float64()*100 < math.Max(20, duelInfo.Weight) {
				duelInfo.Weight -= math.Min(1, userInfo.Weight/10) * rand.Float64()
				messageText = append(messageText, message.Text("\n"), message.At(duelInfo.User), message.Text(duelInfo.Name, "在PK中受伤了\n在医疗中心治愈过程中体重降低至", fmt.Sprintf("%1.2f", duelInfo.Weight)))

			}
			money := 10 + rand.Intn(zbmath.Min(30, userInfo.Mood))
			if wallet.InsertWalletOf(userInfo.User, money) == nil {
				messageText = append(messageText, message.Text("\n"), message.At(userInfo.User), message.Text(userInfo.Name, "为你赢得了", money))
			}
		case !fat && winerWeight >= (winLine+15): //轻,但对面赢了
			messageText = append(messageText, message.Text(duelInfo.Weight, "kg的", duelInfo.Name, "以绝对的体型碾压住了", userInfo.Weight, "kg的", userInfo.Name))
			if rand.Float64()*100 < math.Min(20, userInfo.Weight) {
				userInfo.Weight -= math.Min(1, userInfo.Weight/10) * rand.Float64()
				messageText = append(messageText, message.Text("\n"), message.At(userInfo.User), message.Text(userInfo.Name, "在PK中受伤了\n在医疗中心治愈过程中体重降低至", fmt.Sprintf("%1.2f", userInfo.Weight)))
			}
			money := 10 + rand.Intn(zbmath.Min(30, duelInfo.Mood))
			if wallet.InsertWalletOf(duelInfo.User, money) == nil {
				messageText = append(messageText, message.Text("\n"), message.At(duelInfo.User), message.Text(duelInfo.Name, "为你赢得了", money))
			}
		default:
			messageText = append(messageText, message.Text(duelInfo.Name, "和", userInfo.Name, "并没有打架的意愿呢\nPK结束"))
		}
		userInfo.ArenaTime = time.Now().Unix()
		err = catdata.insert(gidStr, userInfo)
		if err == nil {
			duelInfo.ArenaTime = time.Now().Unix()
			err = catdata.insert(gidStr, duelInfo)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
		}
		ctx.SendChain(messageText...)
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
		}
		messageText := []message.MessageSegment{}
		for i, info := range infoList {
			if i > 9 {
				break
			} else if i != 0 {
				messageText = append(messageText, message.Text("\n"))
			}
			messageText = append(messageText, message.Text(
				i+1, ".", info.Name, "\n",
				"体重：", fmt.Sprintf("%1.2f", info.Weight), "kg\n",
				"主人:", ctx.CardOrNickName(info.User),
			))
		}
		ctx.SendChain(messageText...)
	})
}
