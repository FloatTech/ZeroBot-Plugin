// Package cybercat 云养猫
package cybercat

import (
	"fmt"
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
	engine.OnFullMatchGroup([]string{"猫猫状态", "喵喵状态"}, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
		/***************************************************************/
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
		}
		if subtime > 1 {
			userInfo.Satiety -= subtime
			userInfo = userInfo.settleOfWeight()
			userInfo = userInfo.settleOfMood()
		}
		if userInfo.Weight < 0 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有喂猫猫,", userInfo.Name, "已经饿死了..."))
			return
		} else if userInfo.Weight > 200 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有看猫猫,", userInfo.Name, "已经暴饮暴食撑死了..."))
			return
		}
		/***************************************************************/
		userInfo = userInfo.settleOfData()
		workStauts := "休闲中"
		money, workEnd := userInfo.settleOfWork(gidStr)
		if !workEnd {
			workStauts = "工作中"
		} else if money > 0 {
			workStauts = "从工作回来休息中\n	   为你赚了" + strconv.Itoa(money)
		}
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "当前信息如下:\n"),
			message.Image(userInfo.Picurl),
			message.Text("品种: "+userInfo.Type,
				"\n饱食度: ", strconv.FormatFloat(userInfo.Satiety, 'f', 0, 64),
				"\n心情: ", userInfo.Mood,
				"\n体重: ", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64),
				"\n状态:", workStauts,
				"\n\n你的剩余猫粮(斤): ", strconv.FormatFloat(userInfo.Food, 'f', 2, 64)))
	})
	engine.OnRegex(`^喂猫((\d+(.\d+)?)斤猫粮)?$`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
		_, workEnd := userInfo.settleOfWork(gidStr)
		if !workEnd {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "还在努力打工,没有回来呢"))
			return
		}
		/***************************************************************/
		food := 1.0
		if ctx.State["regex_matched"].([]string)[2] != "" {
			food, _ = strconv.ParseFloat(ctx.State["regex_matched"].([]string)[2], 64)
		}
		if userInfo.Food == 0 || userInfo.Food < food {
			ctx.SendChain(message.Reply(id), message.Text("铲屎官你已经没有足够的猫粮了"))
			return
		}
		/***************************************************************/
		if userInfo.Food > 0 && (rand.Intn(10) == 1 || userInfo.Satiety < 10) {
			eat := (userInfo.Food - food) * rand.Float64()
			userInfo = userInfo.settleOfSatiety(eat)
		}
		/***************************************************************/
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
		}
		if subtime < 8 {
			userInfo.Mood -= 5
			if userInfo.Mood < 0 {
				userInfo.Mood = 0
			}
			if rand.Intn(3) < 0 || userInfo.Mood > 80 {
				_ = catdata.insert(gidStr, userInfo)
				ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情吃东西"))
				return
			}
		}
		if subtime > 1 {
			userInfo.Satiety -= subtime
			userInfo = userInfo.settleOfWeight()
		}
		if userInfo.Weight < 0 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有喂猫猫,", userInfo.Name, "已经饿死了..."))
			return
		} else if userInfo.Weight > 200 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有看猫猫,", userInfo.Name, "已经暴饮暴食撑死了..."))
			return
		}
		/***************************************************************/
		userInfo = userInfo.settleOfMood()
		if userInfo.Satiety > 10 && rand.Intn(100) > zbmath.Max(userInfo.Mood*2-userInfo.Mood/2, 50) {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情吃东西"))
			return
		}
		userInfo = userInfo.settleOfSatiety(food)
		/***************************************************************/
		userInfo = userInfo.settleOfWeight()
		if userInfo.Weight > 200 {
			ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于太胖了,已经难以存活去世了..."))
			return
		}
		/***************************************************************/
		userInfo.LastTime = time.Now().Unix()
		userInfo = userInfo.settleOfData()
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text("猫猫吃完了\n", userInfo.Name, "当前信息如下:\n"),
			message.Image(userInfo.Picurl),
			message.Text("品种: "+userInfo.Type,
				"\n饱食度: ", strconv.FormatFloat(userInfo.Satiety, 'f', 0, 64),
				"\n心情: ", userInfo.Mood,
				"\n体重: ", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64),
				"\n\n你的剩余猫粮(斤): ", fmt.Sprintf("%1.1f", userInfo.Food)))
	})
	engine.OnRegex(`^猫猫打工(([1-9])小时)?$`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
		_, workEnd := userInfo.settleOfWork(gidStr)
		if !workEnd {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "还在努力打工,没有回来呢"))
			return
		}
		/***************************************************************/
		workTime := 1 + rand.Intn(9)
		if ctx.State["regex_matched"].([]string)[2] != "" {
			workTime, _ = strconv.Atoi(ctx.State["regex_matched"].([]string)[2])
		}
		/***************************************************************/
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
		}
		userInfo.Satiety -= subtime
		userInfo = userInfo.settleOfWeight()
		if userInfo.Weight < 0 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有喂猫猫,", userInfo.Name, "已经饿死了..."))
			return
		}
		/***************************************************************/
		userInfo = userInfo.settleOfMood()
		if userInfo.Satiety > 10 && rand.Intn(100) > zbmath.Max(userInfo.Mood*2-userInfo.Mood/2, 50) {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情去工作"))
			return
		}
		/***************************************************************/
		userInfo.Work = time.Now().Unix()*10 + int64(workTime)
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "开始去打工了"))
	})
}

// 食物 & 饱食度结算
/*
	饱食度 = 食物 * 10
		  = 1 * 10
		  = 10
	一袋猫粮 = 5斤食物
*/
func (data *catInfo) settleOfSatiety(food float64) catInfo {
	data.Food -= food
	if food > 0 {
		if data.Mood > 50 && rand.Intn(data.Mood) < data.Mood/3 {
			food *= 4
		}
		data.Satiety += food * 10
	}
	return *data
}

// 体重结算
/*
	饱食度大于80可以长大
	体重 = (饱食度 - 50)/100
		= (80 - 50)/100
		= 0.3
*/
func (data *catInfo) settleOfWeight() catInfo {
	if data.Satiety > 80 {
		data.Weight += (data.Satiety - 50) / 100
	}
	return *data
}

// 心情结算
/*
	饱食度越高心情越好，体重越重越不好
	心情 = 10 + 饱食度 * 0.9 - 体重 * 0.1
		// 50
		= 10 + 50 * 0.9 - 100 * 0.1
		= 10 + 45 - 10
		= 45
		// 100
		= 10 + 100 * 0.9 - 100 * 0.1
		= 10 + 90 - 10
		= 90
*/
func (data *catInfo) settleOfMood() catInfo {
	data.Mood = 10 + int(data.Satiety*0.9-data.Weight*0.1)
	return *data
}

// 整体数据结算
func (data *catInfo) settleOfData() catInfo {
	if data.Satiety > 100 {
		data.Satiety = 100
	} else if data.Satiety < 0 {
		data.Satiety = 0
	}
	if data.Mood > 100 {
		data.Mood = 100
	} else if data.Mood < 0 {
		data.Mood = 0
	}
	return *data
}

// 打工结算
func (data *catInfo) settleOfWork(gid string) (int, bool) {
	workTime := data.Work % 10
	if workTime <= 0 {
		return 0, true
	}
	lastTime := time.Unix(data.Work/10, 0)
	subtime := time.Since(lastTime).Hours()
	if subtime < float64(workTime) {
		return 0, false
	}
	getFood := 5 * rand.Float64() // 工作餐
	data.Satiety += getFood * 10
	data.Work = 0
	if catdata.insert(gid, *data) != nil {
		return 0, true
	}
	getmoney := 10 + rand.Intn(10*int(workTime))
	if wallet.InsertWalletOf(data.User, getmoney) != nil {
		return 0, true
	}
	return getmoney, true
}
