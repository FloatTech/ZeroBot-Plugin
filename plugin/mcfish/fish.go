// Package mcfish 钓鱼模拟器
package mcfish

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex(`^进行(([1-5]\d*)次)?钓鱼$`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		fishNumber := 1
		info := ctx.State["regex_matched"].([]string)[2]
		if info != "" {
			number, err := strconv.Atoi(info)
			if err != nil || number > 50 {
				ctx.SendChain(message.Text("请输入正确的次数"))
				return
			}
			fishNumber = number
		}
		residue, err := dbdata.updateFishInfo(uid, fishNumber)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.1]:", err))
			return
		}
		if residue == 0 {
			ctx.SendChain(message.Text("今天你已经进行", fishLimit, "次钓鱼了.\n游戏虽好,但请不要沉迷。"))
			return
		}
		fishNumber = residue
		equipInfo, err := dbdata.getUserEquip(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.2]:", err))
			return
		}
		if equipInfo == (equip{}) {
			ok, err := dbdata.checkEquipFor(uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.2.1]:", err))
				return
			}
			if !ok {
				ctx.SendChain(message.At(uid), message.Text("请装备鱼竿后钓鱼", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你尚未装备鱼竿,是否花费100购买鱼竿?\n回答\"是\"或\"否\""))
			// 等待用户下一步选择
			recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(是|否)$`), zero.CheckUser(ctx.Event.UserID)).Repeat()
			defer cancel()
			buy := false
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("等待超时,取消钓鱼")))
					return
				case e := <-recv:
					nextcmd := e.Event.Message.String()
					if nextcmd == "否" {
						ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("已取消购买")))
						return
					}
					money := wallet.GetWalletOf(uid)
					if money < 100 {
						ctx.SendChain(message.Text("你钱包当前只有", money, "ATRI币,无法完成支付"))
						return
					}
					err = wallet.InsertWalletOf(uid, -100)
					if err != nil {
						ctx.SendChain(message.Text("[ERROR at fish.go.3]:", err))
						return
					}
					equipInfo = equip{
						ID:      uid,
						Equip:   "木竿",
						Durable: 30,
					}
					err = dbdata.updateUserEquip(equipInfo)
					if err != nil {
						ctx.SendChain(message.Text("[ERROR at fish.go.4]:", err))
						return
					}
					err = dbdata.setEquipFor(uid)
					if err != nil {
						ctx.SendChain(message.Text("[ERROR at fish.go.4]:", err))
						return
					}
					buy = true
				}
				if buy {
					break
				}
			}
		} else if equipInfo.Durable < fishNumber {
			fishNumber = equipInfo.Durable
		}
		msg := ""
		if equipInfo.Equip != "美西螈" {
			equipInfo.Durable -= fishNumber
			err = dbdata.updateUserEquip(equipInfo)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.5]:", err))
				return
			}
			if equipInfo.Durable < 10 {
				msg = "(你的鱼竿耐久仅剩" + strconv.Itoa(equipInfo.Durable) + ")"
			}
		} else {
			fishNmae, err := dbdata.pickFishFor(uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.5.1]:", err))
				return
			}
			if fishNmae == "" {
				equipInfo.Durable = 0
				err = dbdata.updateUserEquip(equipInfo)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.5]:", err))
				}
				ctx.SendChain(message.Text("美西螈因为没吃到鱼,钓鱼时一直没回来,你失去了美西螈"))
				return
			}
			msg = "(美西螈吃掉了一条" + fishNmae + ")"
		}
		waitTime := 120 / (equipInfo.Induce + 1)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你开始去钓鱼了,请耐心等待鱼上钩(预计要", time.Second*time.Duration(waitTime), ")"))
		timer := time.NewTimer(time.Second * time.Duration(rand.Intn(waitTime)+1))
		for {
			<-timer.C
			timer.Stop()
			break
		}
		// 概率
		wasteProbability := 41 + equipInfo.Favor*10
		poleProbability := 11 + equipInfo.Favor*3
		bookProbability := 1 + equipInfo.Favor*1
		// 钓到鱼的范围
		number, err := dbdata.getNumberFor(uid, "鱼")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.5.1]:", err))
			return
		}
		getFishMaxDy := 9
		getFishMinDy := 2
		getFishMaxDx := 9
		getFishMinDx := 1
		if number > 100 || equipInfo.Equip == "美西螈" {
			getFishMaxDy = 10
			getFishMinDy = 1
			getFishMaxDx = 10
			getFishMinDx = 0
		}
		// 钓鱼结算
		thingNameList := make(map[string]int)
		picName := ""
		for i := fishNumber; i > 0; i-- {
			fishDx := rand.Intn(11)
			fishDy := rand.Intn(11)
			if fishDx < getFishMinDx || fishDx > getFishMaxDx || fishDy < getFishMinDy || fishDy > getFishMaxDy {
				if fishNumber == 1 {
					ctx.SendChain(message.At(uid), message.Text("很遗憾你没有钓到鱼", msg))
					return
				}
				thingNameList["空竿"]++
				continue
			}
			dice := rand.Intn(100)
			switch {
			case dice >= wasteProbability: // 垃圾
				waste := wasteList[rand.Intn(len(wasteList))]
				money := 10
				if equipInfo.Equip == "美西螈" {
					money *= 3
				}
				err := wallet.InsertWalletOf(uid, money)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.9]:", err))
					return
				}
				picName = waste
				thingNameList[waste]++
				if fishNumber == 1 {
					msg = "为河流净化做出了贡献,\n给予" + strconv.Itoa(money) + "奖励金\n" + msg
				}
			case dice <= bookProbability:
				picName = "book"
				thingName := "诱钓"
				dice := rand.Intn(100)
				switch {
				case dice == 0:
					picName = "美西螈"
					thingName = "美西螈"
				case dice == 1:
					picName = "唱片"
					thingName = "唱片"
				case dice < 41 && dice > 1:
					thingName = "海之眷顾"
				}
				books, err := dbdata.getUserThingInfo(uid, thingName)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.6]:", err))
					return
				}
				if len(books) == 0 {
					books = append(books, article{
						Duration: time.Now().Unix()*100 + int64(i),
						Type:     "article",
						Name:     thingName,
					})
				}
				if thingName == "美西螈" {
					books[0].Type = "pole"
					books[0].Other = "999/0/0/0"
				}
				number := 1
				if equipInfo.Equip == "美西螈" && thingName != "美西螈" {
					number += 2
				}
				books[0].Number += number
				err = dbdata.updateUserThingInfo(uid, books[0])
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.7]:", err))
					return
				}
				thingNameList[thingName] += number
			case dice > bookProbability && dice <= poleProbability:
				poleNmae := "木竿"
				dice := rand.Intn(100)
				switch {
				case dice >= 10 && dice < 30:
					poleNmae = "铁竿"
				case dice >= 4 && dice < 10:
					poleNmae = "金竿"
				case dice >= 1 && dice < 4:
					poleNmae = "钻石竿"
				case dice == 0:
					poleNmae = "下界合金竿竿竿"
				}
				newPole := article{
					Duration: time.Now().Unix()*100 + int64(i),
					Type:     "pole",
					Name:     poleNmae,
					Number:   1,
					Other: strconv.Itoa(rand.Intn(equipAttribute[poleNmae])+1) +
						"/" + strconv.Itoa(rand.Intn(10)) + "/" +
						strconv.Itoa(rand.Intn(3)) + "/" + strconv.Itoa(rand.Intn(2)),
				}
				err = dbdata.updateUserThingInfo(uid, newPole)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.8]:", err))
					return
				}
				picName = poleNmae
				thingNameList[poleNmae]++
			default:
				fishName := ""
				dice = rand.Intn(100)
				switch {
				case dice == 99:
					fishName = "墨鱼"
				case dice >= 30 && dice != 99:
					fishName = "鳕鱼"
				case dice >= 10 && dice < 30:
					fishName = "鲑鱼"
				case dice >= 4 && dice < 10:
					fishName = "热带鱼"
				case dice >= 1 && dice < 4:
					fishName = "河豚"
				default:
					fishName = "鹦鹉螺"
				}
				fishes, err := dbdata.getUserThingInfo(uid, fishName)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.10]:", err))
					return
				}
				if len(fishes) == 0 {
					fishes = append(fishes, article{
						Duration: time.Now().Unix()*100 + int64(i),
						Type:     "fish",
						Name:     fishName,
					})
				}
				number := 1
				if equipInfo.Equip == "美西螈" || equipInfo.Equip == "三叉戟" {
					number += 2
				}
				fishes[0].Number += number
				err = dbdata.updateUserThingInfo(uid, fishes[0])
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.11]:", err))
					return
				}
				picName = fishName
				thingNameList[fishName] += number
			}
		}
		if len(thingNameList) == 1 {
			thingName := ""
			for name := range thingNameList {
				thingName = name
			}
			pic, err := engine.GetLazyData(picName+".png", false)
			if err != nil {
				logrus.Warnln("[mcfish]error:", err)
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", thingName, "\n", msg))
				return
			}
			ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", thingName, "\n", msg), message.ImageBytes(pic))
			return
		}
		msgInfo := make(message.Message, 0, 3+len(thingNameList))
		msgInfo = append(msgInfo, message.Reply(ctx.Event.MessageID), message.Text("你进行了", fishNumber, "次钓鱼,结果如下:\n"))
		for name, number := range thingNameList {
			msgInfo = append(msgInfo, message.Text(name, " : ", number, "\n"))
		}
		msgInfo = append(msgInfo, message.Text(msg))
		ctx.Send(msgInfo)
	})
}
