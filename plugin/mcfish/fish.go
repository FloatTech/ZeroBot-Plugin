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
	engine.OnFullMatchGroup([]string{"进行钓鱼", "🎣"}, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		ok, err := dbdata.updateFishInfo(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.1]:", err))
			return
		}
		if !ok {
			ctx.SendChain(message.Text("今天你已经进行", fishLimit, "次钓鱼了.\n游戏虽好,但请不要沉迷。"))
			return
		}
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
		}
		msg := ""
		if equipInfo.Equip != "美西螈" {
			equipInfo.Durable--
			err = dbdata.updateUserEquip(equipInfo)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.5]:", err))
				return
			}
			if equipInfo.Durable < 10 {
				msg = "\n(你的鱼竿耐久仅剩" + strconv.Itoa(equipInfo.Durable) + ")"
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
			msg = "\n(美西螈吃掉了一条" + fishNmae + ")"
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
			getFishMaxDx = 10
			getFishMinDx = 0
		}

		fishDx := rand.Intn(11)
		fishDy := rand.Intn(11)
		for i := 0; i < 10; i++ {
			switch rand.Intn(4) {
			case 0:
				fishDx--
				if fishDx < 0 {
					fishDx = 10
				}
			case 1:
				fishDx++
				if fishDx > 10 {
					fishDx = 0
				}
			case 2:
				fishDy--
				if fishDy < 0 {
					fishDy = 10
				}
			default:
				fishDy++
				if fishDy > 10 {
					fishDx = 0
				}
			}
		}
		if fishDx < getFishMinDx || fishDx > getFishMaxDx || fishDy < getFishMinDy || fishDy > getFishMaxDy {
			ctx.SendChain(message.At(uid), message.Text("很遗憾你没有钓到鱼", msg))
			return
		}
		dice := rand.Intn(100)
		var fishName string
		switch {
		case dice <= bookProbability:
			dice := rand.Intn(100)
			picName := "book"
			switch {
			case dice == 0:
				picName = "美西螈"
				fishName = "美西螈"
			case dice == 1:
				picName = "唱片"
				fishName = "唱片"
			case dice < 41 && dice > 1:
				fishName = "海之眷顾"
			default:
				fishName = "诱钓"
			}
			books, err := dbdata.getUserThingInfo(uid, fishName)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.6]:", err))
				return
			}
			if len(books) == 0 {
				books = append(books, article{
					Duration: time.Now().Unix(),
					Type:     "article",
					Name:     fishName,
				})
			}
			if fishName == "美西螈" {
				books[0].Type = "pole"
				books[0].Other = "999/0/0/0"
			}
			books[0].Number++
			if equipInfo.Equip == "美西螈" && fishName != "美西螈" {
				books[0].Number += 2
			}
			err = dbdata.updateUserThingInfo(uid, books[0])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.7]:", err))
				return
			}
			pic, err := engine.GetLazyData(picName+".png", false)
			if err != nil {
				logrus.Warnln("[mcfish]error:", err)
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", books[0].Name, msg))
				return
			}
			ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", books[0].Name, msg), message.ImageBytes(pic))
			return
		case dice > bookProbability && dice <= poleProbability:
			dice := rand.Intn(100)
			switch {
			case dice >= 30:
				fishName = "木竿"
			case dice >= 10 && dice < 30:
				fishName = "铁竿"
			case dice >= 4 && dice < 10:
				fishName = "金竿"
			case dice >= 1 && dice < 4:
				fishName = "钻石竿"
			default:
				fishName = "下界合金竿竿竿"
			}
			newPole := article{
				Duration: time.Now().Unix(),
				Type:     "pole",
				Name:     fishName,
				Number:   1,
				Other:    strconv.Itoa(rand.Intn(equipAttribute[fishName])+1) + "/" + strconv.Itoa(rand.Intn(10)) + "/" + strconv.Itoa(rand.Intn(3)) + "/" + strconv.Itoa(rand.Intn(2)),
			}
			err = dbdata.updateUserThingInfo(uid, newPole)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.8]:", err))
				return
			}
			pic, err := engine.GetLazyData(fishName+".png", false)
			if err != nil {
				logrus.Warnln("[mcfish]error:", err)
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", fishName, msg))
				return
			}
			ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", fishName, msg), message.ImageBytes(pic))
			return
		case dice >= wasteProbability:
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
			pic, err := engine.GetLazyData(waste+".png", false)
			if err != nil {
				logrus.Warnln("[mcfish]error:", err)
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", waste, msg))
				return
			}
			ctx.SendChain(message.At(uid), message.Text("你钓到了", waste, "\n为水源净化做出了贡献,\n给予", money, "奖励金", msg), message.ImageBytes(pic))
			return
		default:
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
					Duration: time.Now().Unix(),
					Type:     "fish",
					Name:     fishName,
				})
			}
			fishes[0].Number++
			if equipInfo.Equip == "美西螈" || equipInfo.Equip == "三叉戟" {
				fishes[0].Number += 2
			}
			err = dbdata.updateUserThingInfo(uid, fishes[0])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at fish.go.11]:", err))
				return
			}
			pic, err := engine.GetLazyData(fishName+".png", false)
			if err != nil {
				logrus.Warnln("[mcfish]error:", err)
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", fishName, msg))
				return
			}
			ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", fishName, msg), message.ImageBytes(pic))
			return
		}
	})
}
