// Package mcfish 钓鱼模拟器
package mcfish

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex("^装备(.*竿)$", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		equipInfo, err := dbdata.getUserEquip(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.1]:", err))
			return
		}
		articles, err := dbdata.getUserThingInfo(uid, ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.2]:", err))
			return
		}
		if len(articles) == 0 {
			ctx.SendChain(message.Text("你的背包不存在该物品"))
			return
		}
		poles := make(map[int]equip, len(articles))
		for i, info := range articles {
			poleInfo := strings.Split(info.Other, "/")
			durable, _ := strconv.Atoi(poleInfo[0])
			maintenance, _ := strconv.Atoi(poleInfo[1])
			induceLevel, _ := strconv.Atoi(poleInfo[2])
			favorLevel, _ := strconv.Atoi(poleInfo[3])
			poles[i] = equip{
				ID:          uid,
				Equip:       info.Name,
				Durable:     durable,
				Maintenance: maintenance,
				Induce:      induceLevel,
				Favor:       favorLevel,
			}
		}
		check := false
		index := 0
		if len(poles) > 1 {
			msg := make(message.Message, 0, 3+len(articles))
			msg = append(msg, message.Reply(ctx.Event.MessageID), message.Text("找到以下鱼竿:\n"))
			for i, info := range poles {
				msg = append(msg, message.Text("[", i, "]耐", info.Durable, "/修", info.Maintenance,
					"/诱", enchantLevel[info.Induce], "/眷顾", enchantLevel[info.Favor],
					"的", info.Equip, "\n"))
			}
			msg = append(msg, message.Text("————————\n输入对应序号进行装备,或回复“取消”取消"))
			ctx.Send(msg)
			// 等待用户下一步选择
			recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(取消|\d+)$`), zero.CheckUser(ctx.Event.UserID)).Repeat()
			defer cancel()
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("等待超时,取消装备"),
						),
					)
					return
				case e := <-recv:
					nextcmd := e.Event.Message.String()
					if nextcmd == "取消" {
						ctx.Send(
							message.ReplyWithMessage(ctx.Event.MessageID,
								message.Text("已取消装备"),
							),
						)
						return
					}
					index, err = strconv.Atoi(nextcmd)
					if err != nil || index > len(articles)-1 {
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
					}
					check = true
				}
				if check {
					break
				}
			}
		}
		newEquipInfo := poles[index]
		packEquip := articles[index]
		packEquip.Number = 0
		err = dbdata.updateUserThingInfo(uid, packEquip)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.3]:", err))
			return
		}
		err = dbdata.updateUserEquip(newEquipInfo)
		if err == nil && equipInfo != (equip{}) {
			oldthing := article{
				Duration: time.Now().Unix(),
				Type:     "pole",
				Name:     equipInfo.Equip,
				Number:   1,
				Other:    strconv.Itoa(equipInfo.Durable) + "/" + strconv.Itoa(equipInfo.Maintenance) + "/" + strconv.Itoa(equipInfo.Induce) + "/" + strconv.Itoa(equipInfo.Favor),
			}
			err = dbdata.updateUserThingInfo(uid, oldthing)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.4]:", err))
			return
		}
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("装备成功"),
			),
		)
	})
	engine.OnFullMatchGroup([]string{"修复鱼竿", "维修鱼竿"}, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		equipInfo, err := dbdata.getUserEquip(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.5]:", err))
			return
		}
		if equipInfo.Equip == "" {
			ctx.SendChain(message.Text("仅能修复装备中的鱼竿"))
			return
		}
		if equipInfo.Maintenance >= 10 {
			ctx.SendChain(message.Text("装备的鱼竿已经达到修复上限"))
			return
		}
		articles, err := dbdata.getUserThingInfo(uid, equipInfo.Equip)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.6]:", err))
			return
		}
		if len(articles) == 0 {
			ctx.SendChain(message.Text("你的背包不存在相同鱼竿进行修复"))
			return
		}
		poles := make(map[int]equip, len(articles))
		for i, info := range articles {
			poleInfo := strings.Split(info.Other, "/")
			durable, _ := strconv.Atoi(poleInfo[0])
			maintenance, _ := strconv.Atoi(poleInfo[1])
			induceLevel, _ := strconv.Atoi(poleInfo[2])
			favorLevel, _ := strconv.Atoi(poleInfo[3])
			poles[i] = equip{
				ID:          uid,
				Equip:       info.Name,
				Durable:     durable,
				Maintenance: maintenance,
				Induce:      induceLevel,
				Favor:       favorLevel,
			}
		}
		index := 0
		check := false
		if len(articles) > 1 {
			msg := make(message.Message, 0, 3+len(articles))
			msg = append(msg, message.Text("找到以下鱼竿:\n"))
			for i, info := range poles {
				msg = append(msg, message.Text(i, ".耐久", info.Durable, "/维修", info.Maintenance,
					"/诱饵", enchantLevel[info.Induce], "/海之眷顾", enchantLevel[info.Favor],
					"的", info.Equip, "\n"))
			}
			msg = append(msg, message.Text("————————\n输入对应序号进行修复,或回复“取消”取消"))
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, msg...))
			// 等待用户下一步选择
			recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(取消|\d+)$`), zero.CheckUser(ctx.Event.UserID)).Repeat()
			defer cancel()
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("等待超时,取消修复"),
						),
					)
					return
				case e := <-recv:
					nextcmd := e.Event.Message.String()
					if nextcmd == "取消" {
						ctx.Send(
							message.ReplyWithMessage(ctx.Event.MessageID,
								message.Text("已取消修复"),
							),
						)
						return
					}
					index, err = strconv.Atoi(nextcmd)
					if err != nil || index > len(articles)-1 {
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
					}
					check = true
				}
				if check {
					break
				}
			}
		}
		newEquipInfo := poles[index]
		number, err := dbdata.getNumberFor(uid, "竿")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.5.1]:", err))
			return
		}
		if number <= 10 {
			number = 8
		} else {
			number = 10
		}
		equipInfo.Durable += newEquipInfo.Durable * number / 10
		if equipInfo.Durable > equipAttribute[equipInfo.Equip] {
			equipInfo.Durable = equipAttribute[equipInfo.Equip]
		}
		if rand.Intn(100) < 50 {
			equipInfo.Induce += newEquipInfo.Induce
			if equipInfo.Induce > 3 {
				equipInfo.Induce = 3
			}
		}
		if rand.Intn(100) < 50 {
			equipInfo.Favor += newEquipInfo.Favor
			if equipInfo.Favor > 3 {
				equipInfo.Favor = 3
			}
		}
		thingInfo := articles[index]
		thingInfo.Number = 0
		err = dbdata.updateUserThingInfo(uid, thingInfo)
		if err == nil {
			equipInfo.Maintenance++
			err = dbdata.updateUserEquip(equipInfo)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.7]:", err))
			return
		}
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("鱼竿修复成功,耐久提高至", equipInfo.Durable),
			),
		)
	})
	engine.OnPrefix("附魔鱼竿", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		equipInfo, err := dbdata.getUserEquip(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.7]:", err))
			return
		}
		if equipInfo == (equip{}) {
			ctx.SendChain(message.Text("仅可对装备的鱼竿进行附魔"))
			return
		}
		book := strings.TrimSpace(ctx.State["args"].(string))
		books, err := dbdata.getUserThingInfo(uid, book)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.8]:", err))
			return
		}
		if len(books) == 0 {
			ctx.SendChain(message.Text("你的背包不存在", book, "进行附魔"))
			return
		}
		bookInfo := books[0]
		bookInfo.Number--
		err = dbdata.updateUserThingInfo(uid, bookInfo)
		number := 0
		if err == nil {
			if rand.Intn(100) > 50 {
				ctx.SendChain(message.Text("附魔失败了"))
				return
			}
			switch book {
			case "诱钓":
				equipInfo.Induce++
				if equipInfo.Induce > 3 {
					equipInfo.Induce = 3
				}
				number = equipInfo.Induce
			case "海之眷顾":
				equipInfo.Favor++
				if equipInfo.Favor > 3 {
					equipInfo.Favor = 3
				}
				number = equipInfo.Favor
			default:
				ctx.SendChain(message.Text("附魔失败了"))
				return
			}
			err = dbdata.updateUserEquip(equipInfo)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pole.go.9]:", err))
			return
		}
		ctx.SendChain(message.Text("附魔成功,", book, "等级提高至", enchantLevel[number]))
	})
}
