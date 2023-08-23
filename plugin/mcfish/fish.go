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
	engine.OnFullMatch("进行钓鱼", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
			ctx.SendChain(message.Text("你尚未装备鱼竿,是否花费100购买鱼竿?\n回答\"是\"或\"否\""))
			// 等待用户下一步选择
			recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(是|否)$`), zero.OnlyGroup, zero.CheckUser(ctx.Event.UserID)).Repeat()
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
					buy = true
				}
				if buy {
					break
				}
			}
		}
		waitTime := 120 / (equipInfo.Induce + 1)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你开始去钓鱼了,请耐心等待鱼上钩(预计要", time.Second*time.Duration(waitTime), ")"))
		equipInfo.Durable--
		err = dbdata.updateUserEquip(equipInfo)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.5]:", err))
			return
		}
		timer := time.NewTimer(time.Second * time.Duration(rand.Intn(waitTime)+1))
		for {
			<-timer.C
			timer.Stop()
			// 概率
			wasteProbability := 41 + equipInfo.Favor*10
			poleProbability := 11 + equipInfo.Favor*3
			bookProbability := 1 + equipInfo.Favor*1
			// 钓到鱼的范围
			getFishMaxDy := 8 + equipInfo.Induce
			getFishMinDy := 2 - equipInfo.Induce
			getFishMaxDx := 8 + equipInfo.Induce
			getFishMinDx := 2 - equipInfo.Induce

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
				ctx.SendChain(message.At(uid), message.Text("很遗憾你没有钓到鱼"))
				return
			}
			dice := rand.Intn(100)
			switch {
			case dice <= bookProbability:
				bookName := "诱钓"
				if rand.Intn(2) == 1 {
					bookName = "海之眷顾"
				}
				books, err := dbdata.getUserThingInfo(uid, bookName)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.6]:", err))
					return
				}
				if len(books) == 0 {
					books = append(books, article{
						Duration: time.Now().Unix(),
						Name:     bookName,
					})
				}
				books[0].Number++
				err = dbdata.updateUserThingInfo(uid, books[0])
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.7]:", err))
					return
				}
				pic, err := engine.GetLazyData("book.png", false)
				if err != nil {
					logrus.Warnln("[mcfish]error:", err)
					ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", books[0].Name))
					return
				}
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", books[0].Name), message.ImageBytes(pic))
				return
			case dice > bookProbability && dice <= poleProbability:
				var poleName string
				switch rand.Intn(len(equipAttribute)) {
				case 1:
					poleName = "铁竿"
				case 2:
					poleName = "金竿"
				case 3:
					poleName = "钻石竿"
				case 4:
					poleName = "下界合金竿竿竿"
				default:
					poleName = "木竿"
				}
				newPole := article{
					Duration: time.Now().Unix(),
					Name:     poleName,
					Number:   1,
					Other:    strconv.Itoa(rand.Intn(equipAttribute[poleName])+1) + "/" + strconv.Itoa(rand.Intn(10)) + "/" + strconv.Itoa(rand.Intn(3)) + "/" + strconv.Itoa(rand.Intn(3)),
				}
				err = dbdata.updateUserThingInfo(uid, newPole)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.8]:", err))
					return
				}
				pic, err := engine.GetLazyData(poleName+".png", false)
				if err != nil {
					logrus.Warnln("[mcfish]error:", err)
					ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", poleName))
					return
				}
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", poleName), message.ImageBytes(pic))
				return
			case dice >= wasteProbability:
				waste := wasteList[rand.Intn(len(wasteList))]
				money := 10
				err := wallet.InsertWalletOf(uid, money)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.9]:", err))
					return
				}
				pic, err := engine.GetLazyData(waste+".png", false)
				if err != nil {
					logrus.Warnln("[mcfish]error:", err)
					ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", waste))
					return
				}
				ctx.SendChain(message.At(uid), message.Text("你钓到了", waste, ",为河流净化做出了贡献,给了你", money, "奖励金"), message.ImageBytes(pic))
				return
			default:
				var fishName string
				dice = rand.Intn(100)
				switch {
				case dice >= 70:
					fishName = "鲑鱼"
				case dice >= 20 && dice < 70:
					fishName = "热带鱼"
				case dice >= 6 && dice < 20:
					fishName = "河豚"
				case dice >= 3 && dice < 6:
					fishName = "鹦鹉螺"
				default:
					fishName = "鳕鱼"
				}
				fishes, err := dbdata.getUserThingInfo(uid, fishName)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.10]:", err))
					return
				}
				if len(fishes) == 0 {
					fishes = append(fishes, article{
						Duration: time.Now().Unix(),
						Name:     fishName,
					})
				}
				fishes[0].Number++
				err = dbdata.updateUserThingInfo(uid, fishes[0])
				if err != nil {
					ctx.SendChain(message.Text("[ERROR at fish.go.11]:", err))
					return
				}
				pic, err := engine.GetLazyData(fishName+".png", false)
				if err != nil {
					logrus.Warnln("[mcfish]error:", err)
					ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", fishName))
					return
				}
				ctx.SendChain(message.At(uid), message.Text("恭喜你钓到了", fishName), message.ImageBytes(pic))
				if equipInfo.Durable < 10 {
					ctx.SendChain(message.At(uid), message.Text("你的鱼竿耐久仅剩", equipInfo.Durable))
				}
				return
			}
		}
	})
}
