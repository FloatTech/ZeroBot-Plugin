// Package cybercat 云养猫
package cybercat

import (
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
	engine.OnRegex(`^喂猫((\d+(.\d+)?)斤猫粮)?|猫猫状态$`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
		/*************************判断指令类型**************************************/
		cmd := false // false表示在喂猫
		if ctx.State["regex_matched"].([]string)[0] == "猫猫状态" {
			cmd = true
		}
		/**************************获取工作状态*************************************/
		stauts := "休闲中"
		money, workEnd := userInfo.settleOfWork(gidStr)
		switch {
		case !cmd && !workEnd:
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "还在努力打工,没有回来呢"))
			return
		case cmd && !workEnd:
			overwork := time.Unix(userInfo.Work/10, 0).Add(time.Hour * time.Duration(userInfo.Work%10))
			stauts = overwork.Format("工作中\n(将在01月02日15:04下班)")
		case cmd && money > 0:
			stauts = "从工作回来休息中\n	为你赚了" + strconv.Itoa(money)
		}
		now := time.Now().Hour()
		if !cmd && ((now < 6 || (now > 8 && now < 11) || (now > 14 && now < 17) || now > 21) && (userInfo.Satiety > 50 || rand.Intn(3) != 1)) {
			if userInfo.Satiety > 50 {
				ctx.SendChain(message.Text("猫猫拍了拍饱饱的肚子表示并不饿呢"))
				return
			}
			ctx.SendChain(message.Text("猫猫只想和你一起吃传统早中晚饭咧"))
			return
		}
		/****************************计算食物数量***********************************/
		food := 0.0
		if !cmd {
			stauts = "刚刚的食物很美味"
			// 如果没有指定猫粮就 （1 + 猫粮/5*x ）斤猫粮
			if ctx.State["regex_matched"].([]string)[2] != "" {
				food, _ = strconv.ParseFloat(ctx.State["regex_matched"].([]string)[2], 64)
			} else {
				food = 1.0 + math.Max(userInfo.Food-1, 0)/5*rand.Float64()
			}
			switch {
			case userInfo.Food == 0 || userInfo.Food < food:
				ctx.SendChain(message.Reply(id), message.Text("铲屎官你已经没有足够的猫粮了"))
				return
			// 如果猫粮太多就只吃一点，除非太饿了
			case food > 5 && (rand.Intn(10) < 8 || userInfo.Satiety < 30):
				food = 5
				stauts = "食物实在太多了!"
			case food < 0.5:
				ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "骂骂咧咧的走了"))
				return
			}
		}
		/****************************空闲时间猫体力的减少计算***********************************/
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
			userInfo.LastTime = time.Unix(userInfo.LastTime, 0).Add(time.Duration(subtime) * time.Hour).Unix()
		}
		// 频繁喂猫减少心情
		if !cmd && subtime < 8 {
			userInfo.Mood -= 5
			if userInfo.Mood < 0 {
				userInfo.Mood = 0
			}
			if rand.Intn(10) < 6 && subtime < 2 && userInfo.Satiety > 90 {
				if err = catdata.insert(gidStr, userInfo); err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "肚子已经很饱了,吃不动了"))
				return
			}
		}
		// 当饱食度降到0以下，体力减少
		if subtime > 1 {
			userInfo.Satiety -= subtime * 4
			userInfo = userInfo.settleOfWeight()
			userInfo.Mood -= int(subtime)
			userInfo.LastTime = time.Now().Unix()
		}
		/***************************太饿了偷吃************************************/
		userInfo.Food -= food
		if userInfo.Food > 0 && (rand.Intn(10) == 1 || userInfo.Satiety < 10) {
			eat := userInfo.Food / 5 * rand.Float64()
			userInfo = userInfo.settleOfSatiety(eat)
			userInfo.Mood += int(eat)
			userInfo = userInfo.settleOfWeight()
		}
		/***************************整体结算，判断当前的心情是否继续************************************/
		userInfo = userInfo.settleOfData()
		if !cmd && userInfo.Satiety > 80 && rand.Intn(100) > zbmath.Max(userInfo.Mood*2-userInfo.Mood/2, 50) {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情吃东西"))
			return
		}
		/****************************结算食物***********************************/
		userInfo = userInfo.settleOfSatiety(food)
		userInfo = userInfo.settleOfWeight()
		switch {
		case userInfo.Mood <= 0 && rand.Intn(100) < 10:
			if err = catdata.delcat(gidStr, uidStr); err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "和你的感情淡了,选择了离家出走"))
			return
		case userInfo.Weight <= 0 && time.Since(time.Unix(userInfo.LastTime, 0)).Hours() > 72: //三天不喂食就死
			if err = catdata.delcat(gidStr, uidStr); err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于瘦骨如柴,已经难以存活去世了..."))
			return
		case userInfo.Weight >= 100:
			if 100*rand.Float64() > math.Max(userInfo.Weight-100, 10) { // 越胖越容易成功
				if err = catdata.delcat(gidStr, uidStr); err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				ctx.SendChain(message.Reply(id), message.Text("猫猫", userInfo.Name, "由于太胖了,已经难以存活去世了..."))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("渡劫成功!", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64), "kg的",
				userInfo.Name, "进化成猫娘了!\n可以发送“上传猫猫照片”修改图像了喔"))
			userInfo.Type = "猫娘"
			userInfo.Weight = 3 + rand.Float64()*10
		}
		/****************************保存数据***********************************/
		userInfo.LastTime = time.Now().Unix()
		userInfo.Mood += int(userInfo.Satiety)/5 - int(userInfo.Weight)/10
		userInfo = userInfo.settleOfData()
		if err = catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if !cmd && userInfo.Satiety < 80 {
			stauts = "完全没有饱"
		}
		ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "当前信息如下:\n"),
			message.Image(userInfo.Picurl),
			message.Text("品种: "+userInfo.Type,
				"\n饱食度: ", strconv.FormatFloat(userInfo.Satiety, 'f', 0, 64),
				"\n心情: ", userInfo.Mood,
				"\n体重: ", strconv.FormatFloat(userInfo.Weight, 'f', 2, 64),
				"\n状态:", stauts,
				"\n\n你的剩余猫粮(斤): ", strconv.FormatFloat(userInfo.Food, 'f', 2, 64)))
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
		/***************************************************************/
		_, workEnd := userInfo.settleOfWork(gidStr)
		if !workEnd {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "还在努力打工,没有回来呢"))
			return
		}
		if userInfo.Work > 0 && time.Unix(userInfo.Work/10, 0).Day() == time.Now().Day() && rand.Intn(100) < 10 {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "已经很累了,你不能这么资本"))
			return
		}
		/***************************************************************/
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
			userInfo.LastTime = time.Unix(userInfo.LastTime, 0).Add(time.Duration(subtime) * time.Hour).Unix()
		}
		userInfo.Satiety -= subtime
		userInfo.Mood -= int(subtime)
		userInfo = userInfo.settleOfWeight()
		if userInfo.Weight < 0 {
			if err = catdata.delcat(gidStr, uidStr); err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有喂猫猫,", userInfo.Name, "已经饿死了..."))
			return
		}
		/***************************************************************/
		userInfo = userInfo.settleOfData()
		if userInfo.Satiety > 90 && rand.Intn(100) > zbmath.Max(userInfo.Mood*2-userInfo.Mood/2, 50) {
			ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, "好像并没有心情去工作"))
			return
		}
		/***************************************************************/
		workTime := 1 + rand.Intn(9)
		if ctx.State["regex_matched"].([]string)[2] != "" {
			workTime, _ = strconv.Atoi(ctx.State["regex_matched"].([]string)[2])
		}
		userInfo.Work = time.Now().Unix()*10 + int64(workTime)
		if err = catdata.insert(gidStr, userInfo); err != nil {
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
		/***************************************************************/
		subtime := 0.0
		if userInfo.LastTime != 0 {
			lastTime := time.Unix(userInfo.LastTime, 0)
			subtime = time.Since(lastTime).Hours()
			userInfo.LastTime = time.Unix(userInfo.LastTime, 0).Add(time.Duration(subtime) * time.Hour).Unix()
		}
		userInfo.Satiety -= subtime
		userInfo.Mood -= int(subtime)
		userInfo = userInfo.settleOfWeight()
		if userInfo.Weight < 0 {
			if err = catdata.delcat(gidStr, uidStr); err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Reply(id), message.Text("由于你长时间没有喂猫猫,", userInfo.Name, "已经饿死了..."))
			return
		}
		/***************************************************************/
		choose := rand.Intn(2)
		text := "被调教得屁股高跷呢!心情提高至"
		switch choose {
		case 0:
			text = "不耐烦的走掉了,心情降低至"
			userInfo.Mood -= rand.Intn(zbmath.Max(1, userInfo.Mood))
		case 1:
			userInfo.Mood += rand.Intn(100)
		}
		userInfo = userInfo.settleOfData()
		if err = catdata.insert(gidStr, userInfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(id), message.Text(userInfo.Name, text, userInfo.Mood))
	})
}

// 饱食度结算
func (data *catInfo) settleOfSatiety(food float64) catInfo {
	if food > 0 && data.Satiety < 30 && rand.Intn(100) <= data.Mood/3 {
		food *= 4
	}
	data.Satiety += (food * 100 / math.Max(1, data.Weight/2))
	return *data
}

// 体重结算
func (data *catInfo) settleOfWeight() catInfo {
	if data.Weight < 0 {
		satiety := math.Min(-data.Weight*7, data.Satiety)
		data.Weight += satiety
		data.Satiety -= satiety
	}
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
	data.Work = time.Now().Unix() * 10
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
