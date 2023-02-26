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
		workStauts := "休闲中"
		money, workEnd := userInfo.settleOfWork(gidStr)
		if !workEnd {
			overwork := time.Unix(userInfo.Work/10, 0).Add(time.Hour * time.Duration(userInfo.Work%10))
			workStauts = overwork.Format("工作中\n(将在01月02日15:04下班)")
		} else {
			/***************************************************************/
			if userInfo.Food > 0 && (rand.Intn(10) == 1 || userInfo.Satiety < 10) {
				eat := userInfo.Food / 5 * rand.Float64()
				userInfo = userInfo.settleOfSatiety(eat)
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
				userInfo.Mood -= int(subtime)
				userInfo = userInfo.settleOfData()
			}
		}
		if money > 0 {
			workStauts = "从工作回来休息中\n	为你赚了" + strconv.Itoa(money)
		}
		/***************************************************************/
		if userInfo.Weight <= 0 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于瘦骨如柴,已经难以存活去世了..."))
			return
		} else if userInfo.Weight >= 200 {
			if rand.Intn(100) != 50 {
				if catdata.delcat(gidStr, uidStr) != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于太胖了,已经难以存活去世了..."))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("渡劫成功！", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64), "kg的",
				userInfo.Name, "进化成猫娘了!\n可以发送“上传猫猫照片”修改图像了喔"))
			userInfo.Type = "猫娘"
			userInfo.Weight = 3 + rand.Float64()*10
		}
		userInfo = userInfo.settleOfData()
		userInfo.LastTime = time.Now().Unix()
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
		result := "表示食物很美味呢~"
		switch {
		case food > 5 && rand.Intn(10) < 8:
			food = 5
			result = "并没有选择吃完呢"
		case food < 0.5:
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "骂骂咧咧的走了"))
			return

		}
		/***************************************************************/
		if userInfo.Food > 0 && (rand.Intn(10) == 1 || userInfo.Satiety < 10) {
			eat := (userInfo.Food - food) / 5 * rand.Float64()
			userInfo = userInfo.settleOfSatiety(eat)
			userInfo.Mood += int(eat)
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
			if rand.Intn(10) == 1 || userInfo.Mood > 80 {
				_ = catdata.insert(gidStr, userInfo)
				ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情吃东西"))
				return
			}
		}
		if subtime > 1 {
			userInfo.Satiety -= subtime
			userInfo = userInfo.settleOfWeight()
			userInfo.Mood -= int(subtime)
		}
		/***************************************************************/
		userInfo = userInfo.settleOfData()
		if userInfo.Satiety > 80 && rand.Intn(100) > zbmath.Max(userInfo.Mood*2-userInfo.Mood/2, 50) {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情吃东西"))
			return
		}
		lastSatiety := userInfo.Satiety
		userInfo = userInfo.settleOfSatiety(food)
		/***************************************************************/
		userInfo = userInfo.settleOfWeight()
		if userInfo.Weight <= 0 {
			if catdata.delcat(gidStr, uidStr) != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于瘦骨如柴,已经难以存活去世了..."))
			return
		} else if userInfo.Weight >= 200 {
			if rand.Intn(100) != 50 {
				if catdata.delcat(gidStr, uidStr) != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于太胖了,已经难以存活去世了..."))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("渡劫成功！", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64), "kg的",
				userInfo.Name, "进化成猫娘了!\n可以发送“上传猫猫照片”修改图像了喔"))
			userInfo.Type = "猫娘"
			userInfo.Weight = 3 + rand.Float64()*10
		}
		/***************************************************************/
		userInfo.LastTime = time.Now().Unix()
		userInfo = userInfo.settleOfData()
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if userInfo.Satiety < 80 && userInfo.Satiety-lastSatiety < 30 {
			result = "表示完全没有饱呢!"
		}
		ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, result, "\n------状态------\n",
			"饱食度: ", strconv.FormatFloat(userInfo.Satiety, 'f', 0, 64),
			"\n心情: ", userInfo.Mood,
			"\n体重: ", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64),
			"\n------仓库------",
			"\n剩余猫粮(斤): ", fmt.Sprintf("%1.1f", userInfo.Food)))
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
		if time.Unix(userInfo.Work/10, 0).Day() == time.Now().Day() && rand.Intn(100) != 1 {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "已经很累了,你不能这么资本"))
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
		userInfo = userInfo.settleOfData()
		if userInfo.Mood > 10 && rand.Intn(100) > zbmath.Max(userInfo.Mood*2-userInfo.Mood/2, 50) {
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
	engine.OnFullMatchGroup([]string{"逗猫", "撸猫", "rua猫", "mua猫", "玩猫", "摸猫"}, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
		}
		if userInfo.LastTime != 0 && subtime < 2 && rand.Intn(5) < 3 {
			ctx.SendChain(message.Reply(id), message.Text("刚吃饱没多久的", userInfo.Name, "跑走去睡觉了"))
			return
		}
		/***************************************************************/
		choose := rand.Intn(2)
		text := "被调教得屁股高跷呢!心情提高至"
		switch choose {
		case 0:
			text = "不耐烦的走掉了,心情降低至"
			userInfo.Mood -= rand.Intn(userInfo.Mood)
		case 1:
			userInfo.Mood += rand.Intn(100)
		}
		userInfo = userInfo.settleOfData()
		if catdata.insert(gidStr, userInfo) != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, text, userInfo.Mood))
	})
}

// 饱食度结算
func (data *catInfo) settleOfSatiety(food float64) catInfo {
	if food > 0 && data.Satiety < 30 && rand.Intn(data.Mood+1) <= data.Mood/3 {
		food *= 4
	}
	data.Satiety += food * 100 / math.Max(1, data.Weight)
	return *data
}

// 体重结算
func (data *catInfo) settleOfWeight() catInfo {
	switch {
	case data.Satiety > 100:
		data.Weight += (data.Satiety - 50) / 100
	case data.Satiety < 0:
		data.Weight += data.Satiety / 10
	}
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
	getFood := 5 * rand.Float64()
	mood := rand.Intn(int(workTime))
	if rand.Intn(5) < 3 { // 60%受饿
		getFood = -(getFood + float64(workTime)*rand.Float64())
		mood *= -3
	}
	data.Satiety += getFood * 100 / math.Max(1, data.Weight)
	data.Mood += mood
	data.LastTime = time.Unix(data.LastTime, 0).Add(time.Duration(workTime) * time.Hour).Unix()
	if catdata.insert(gid, *data) != nil {
		return 0, true
	}
	getmoney := 10 + rand.Intn(10*int(workTime))
	if wallet.InsertWalletOf(data.User, getmoney) != nil {
		return 0, true
	}
	return getmoney, true
}
