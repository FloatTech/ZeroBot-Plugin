// Package mcfish 钓鱼模拟器
package mcfish

import (
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	lastTime    = 0
	refreshFish = func(ctx *zero.Ctx) bool {
		if len(discount) != 0 && lastTime == time.Now().Day() {
			return true
		}
		ok, err := dbdata.refreshStroeInfo()
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.1]:", err))
			return false
		}
		lastTime = time.Now().Day()
		return ok
	}
)

func init() {
	engine.OnFullMatchGroup([]string{"钓鱼看板", "钓鱼商店"}, getdb, refreshFish).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		infos, err := dbdata.getStoreInfo()
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.2]:", err))
			return
		}
		var picImage image.Image
		if len(infos) == 0 {
			picImage, err = drawStroeEmptyImage()
		} else {
			picImage, err = drawStroeInfoImage(infos)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.3]:", err))
			return
		}
		pic, err := imgfactory.ToBytes(picImage)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.4]:", err))
			return
		}
		ctx.SendChain(message.ImageBytes(pic))
	})
	engine.OnRegex(`^出售(.+(竿|鱼)|河豚|鹦鹉螺|诱钓|海之眷顾)\s*(\d*)$`, getdb, refreshFish).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		thingName := ctx.State["regex_matched"].([]string)[1]
		number, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[3])
		if number == 0 || strings.Contains(thingName, "竿") {
			number = 1
		}
		articles, err := dbdata.getUserThingInfo(uid, thingName)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.5]:", err))
			return
		}
		if len(articles) == 0 {
			ctx.SendChain(message.Text("你的背包不存在该物品"))
			return
		}
		index := 0
		thing := article{}
		if len(articles) > 1 {
			msg := make(message.Message, 3+len(articles))
			msg = append(msg, message.Reply(ctx.Event.MessageID), message.Text("找到以下物品:\n"))
			for i, info := range articles {
				if info.Other != "" {
					msg = append(msg, message.Text(
						strconv.Itoa(i), info.Name, "(", info.Other, ")", "  ", info.Number, "\n"))
				} else {
					msg = append(msg, message.Text(
						strconv.Itoa(i), info.Name, "  ", info.Number, "\n"))
				}

			}
			msg = append(msg, message.Reply(ctx.Event.MessageID), message.Text("\n————————————————\n输入对应序号进行装备,或回复“取消”取消"))
			ctx.Send(msg)
			// 等待用户下一步选择
			sell := false
			recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(取消|\d+)$`), zero.CheckUser(ctx.Event.UserID)).Repeat()
			defer cancel()
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("等待超时,取消出售"),
						),
					)
					return
				case e := <-recv:
					nextcmd := e.Event.Message.String()
					if nextcmd == "取消" {
						ctx.Send(
							message.ReplyWithMessage(ctx.Event.MessageID,
								message.Text("已取消出售"),
							),
						)
						return
					}
					index, err = strconv.Atoi(nextcmd)
					if err != nil || index > len(articles)-1 {
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
						continue
					}
					sell = true
				}
				if sell {
					break
				}
			}
		}
		thing = articles[index]
		thing.Number -= number
		err = dbdata.updateUserThingInfo(uid, thing)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.6]:", err))
			return
		}
		var pice int
		if strings.Contains(thingName, "竿") {
			poleInfo := strings.Split(thing.Other, "/")
			durable, _ := strconv.Atoi(poleInfo[0])
			maintenance, _ := strconv.Atoi(poleInfo[1])
			induceLevel, _ := strconv.Atoi(poleInfo[2])
			favorLevel, _ := strconv.Atoi(poleInfo[3])
			equipPice := (thingPice[thingName]-(equipAttribute[thingName]-durable)-maintenance*2)*discount[thingName]/100 + induceLevel*1000 + favorLevel*2500
			newCommodity := store{
				Duration: time.Now().Unix(),
				Name:     thingName,
				Number:   1,
				Price:    equipPice,
				Other:    thing.Other,
			}
			err = dbdata.updateStoreInfo(newCommodity)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at store.go.7]:", err))
				return
			}
			pice = equipPice * 6 / 10
		} else {
			pice = thingPice[thingName] * discount[thingName] / 100
			things, err1 := dbdata.getStoreThingInfo(thingName)
			if err1 != nil {
				ctx.SendChain(message.Text("[ERROR at store.go.8]:", err1))
				return
			}
			if len(things) == 0 {
				things = append(things, store{
					Duration: time.Now().Unix(),
					Name:     thingName,
					Number:   0,
					Price:    pice,
				})
			}
			things[0].Number += number
			err = dbdata.updateStoreInfo(things[0])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at store.go.9]:", err))
				return
			}
		}
		err = wallet.InsertWalletOf(uid, pice)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.10]:", err))
			return
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("出售成功,你赚到了", pice)))
	})
	engine.OnRegex(`^购买(.+(竿|鱼)|河豚|鹦鹉螺|诱钓|海之眷顾)\s*(\d*)$`, getdb, refreshFish).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		thingName := ctx.State["regex_matched"].([]string)[1]
		number, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[3])
		if number == 0 {
			number = 1
		}
		thingInfos, err := dbdata.getStoreThingInfo(thingName)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.11]:", err))
			return
		}
		if len(thingInfos) == 0 {
			ctx.SendChain(message.Text("当前商店并没有上架该物品"))
			return
		}
		index := 0
		pice := 0
		if len(thingInfos) > 1 {
			msg := make(message.Message, 3+len(thingInfos))
			msg = append(msg, message.Text("找到以下物品:\n"))
			for i, info := range thingInfos {
				if strings.Contains(thingName, "竿") {
					poleInfo := strings.Split(info.Other, "/")
					durable, _ := strconv.Atoi(poleInfo[0])
					maintenance, _ := strconv.Atoi(poleInfo[1])
					induceLevel, _ := strconv.Atoi(poleInfo[2])
					favorLevel, _ := strconv.Atoi(poleInfo[3])
					pice = (thingPice[info.Name]-(equipAttribute[info.Name]-durable)-maintenance*2)*discount[info.Name]/100 + induceLevel*1000 + favorLevel*2500
					msg = append(msg, message.Text(
						strconv.Itoa(i), info.Name, "(", info.Other, ")", "  数量:", info.Number, "  价格:", pice, "\n"))
				} else {
					pice = thingPice[thingName] * discount[thingName] / 100
					msg = append(msg, message.Text(
						strconv.Itoa(i), info.Name, "  数量:", info.Number, "  价格:", pice, "\n"))
				}

			}
			msg = append(msg, message.Text("\n————————————————\n输入对应序号进行装备,或回复“取消”取消"))
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, msg...))
			// 等待用户下一步选择
			sell := false
			recv, cancel := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^(取消|\d+)$`), zero.CheckUser(ctx.Event.UserID)).Repeat()
			defer cancel()
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("等待超时,取消出售"),
						),
					)
					return
				case e := <-recv:
					nextcmd := e.Event.Message.String()
					if nextcmd == "取消" {
						ctx.Send(
							message.ReplyWithMessage(ctx.Event.MessageID,
								message.Text("已取消出售"),
							),
						)
						return
					}
					index, err = strconv.Atoi(nextcmd)
					if err != nil || index > len(thingInfos)-1 {
						ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入正确的序号"))
						continue
					}
					sell = true
				}
				if sell {
					break
				}
			}
		}
		thing := thingInfos[index]
		thing.Number -= number
		err = dbdata.updateStoreInfo(thing)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.12]:", err))
			return
		}
		money := wallet.GetWalletOf(uid)
		if money < pice {
			ctx.SendChain(message.Text("你身上的钱(", money, ")不够支付"))
			return
		}
		err = wallet.InsertWalletOf(uid, -pice)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.13]:", err))
			return
		}
		if strings.Contains(thingName, "竿") {
			newCommodity := article{
				Duration: time.Now().Unix(),
				Name:     thingName,
				Number:   1,
				Other:    thing.Other,
			}
			err = dbdata.updateUserThingInfo(uid, newCommodity)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at store.go.14]:", err))
				return
			}
		} else {
			things, err1 := dbdata.getUserThingInfo(uid, thingName)
			if err1 != nil {
				ctx.SendChain(message.Text("[ERROR at store.go.15]:", err1))
				return
			}
			if len(things) == 0 {
				things = append(things, article{
					Duration: time.Now().Unix(),
					Name:     thingName,
					Number:   0,
				})
			}
			things[0].Number += number
			err = dbdata.updateUserThingInfo(uid, things[0])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR at store.go.16]:", err))
				return
			}
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("你成功花了", pice, "购买了", thingName)))
	})
}

func drawStroeEmptyImage() (picImage image.Image, err error) {
	fontdata, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	canvas := gg.NewContext(1000, 300)
	// 画底色
	canvas.DrawRectangle(0, 0, 1000, 300)
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 边框框
	canvas.DrawRectangle(0, 0, 1000, 300)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	canvas.SetColor(color.Black)
	err = canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString("价格信息")
	canvas.DrawString("价格信息", 10, 10+textH*1.2)
	canvas.DrawLine(10, textH*1.6, textW, textH*1.6)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	if err = canvas.ParseFontFace(fontdata, 50); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored("当前商店并没有上架任何物品", 500, 10+textH*2+50, 0.5, 0)
	return canvas.Image(), nil
}

func drawStroeInfoImage(stroeInfo []store) (picImage image.Image, err error) {
	fontdata, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	canvas := gg.NewContext(1, 1)
	err = canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	titleW, titleH := canvas.MeasureString("价格信息")

	err = canvas.ParseFontFace(fontdata, 50)
	if err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("高度")
	_, nameW := canvas.MeasureString("名称")
	_, numberW := canvas.MeasureString("数量")
	_, priceW := canvas.MeasureString("价格")
	for _, info := range stroeInfo {
		textW, _ := canvas.MeasureString(info.Name + "(" + info.Other + ")")
		if nameW < textW {
			nameW = textW
		}
		textW, _ = canvas.MeasureString(strconv.Itoa(info.Number))
		if numberW < textW {
			numberW = textW
		}
		pice := 0
		if strings.Contains(info.Name, "竿") {
			poleInfo := strings.Split(info.Other, "/")
			durable, _ := strconv.Atoi(poleInfo[0])
			maintenance, _ := strconv.Atoi(poleInfo[1])
			induceLevel, _ := strconv.Atoi(poleInfo[2])
			favorLevel, _ := strconv.Atoi(poleInfo[3])
			pice = (thingPice[info.Name]-(equipAttribute[info.Name]-durable)-maintenance*2)*discount[info.Name]/100 + induceLevel*1000 + favorLevel*2500
		} else {
			pice = thingPice[info.Name] * discount[info.Name] / 100
		}
		textW, _ = canvas.MeasureString(strconv.Itoa(pice))
		if priceW < textW {
			priceW = textW
		}
	}

	bolckW := int(10 + nameW + 50 + numberW + 50 + priceW + 10)
	backY := 10 + int(titleH*2) + 10 + (len(stroeInfo)+2)*int(textH*2) + 10
	canvas = gg.NewContext(bolckW, math.Max(backY, 500))
	// 画底色
	canvas.DrawRectangle(0, 0, float64(bolckW), float64(backY))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 边框框
	canvas.DrawRectangle(0, 0, float64(bolckW), float64(backY))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	// 放字
	canvas.SetColor(color.Black)
	err = canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	canvas.DrawString("价格信息", 10, 10+titleH*1.2)
	canvas.DrawLine(10, titleH*1.6, titleW, titleH*1.6)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	textDy := 10 + titleH*1.7
	if err = canvas.ParseFontFace(fontdata, 50); err != nil {
		return nil, err
	}
	canvas.SetColor(color.Black)

	canvas.DrawStringAnchored("名称", 10+nameW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("数量", 10+nameW+10+numberW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("价格", 10+nameW+10+numberW+50+priceW/2, textDy+textH/2, 0.5, 0.5)

	for _, info := range stroeInfo {
		textDy += textH * 2
		name := info.Name
		if info.Other != "" {
			name += "(" + info.Other + ")"
		}
		numberStr := strconv.Itoa(info.Number)
		pice := 0
		if strings.Contains(name, "竿") {
			poleInfo := strings.Split(info.Other, "/")
			durable, _ := strconv.Atoi(poleInfo[0])
			maintenance, _ := strconv.Atoi(poleInfo[1])
			induceLevel, _ := strconv.Atoi(poleInfo[2])
			favorLevel, _ := strconv.Atoi(poleInfo[3])
			pice = (thingPice[info.Name]-(equipAttribute[info.Name]-durable)-maintenance*2)*discount[info.Name]/100 + induceLevel*1000 + favorLevel*2500
		} else {
			pice = thingPice[info.Name] * discount[info.Name] / 100
		}

		canvas.DrawStringAnchored(name, 10+nameW/2, textDy+textH/2, 0.5, 0.5)
		canvas.DrawStringAnchored(numberStr, 10+nameW+10+numberW/2, textDy+textH/2, 0.5, 0.5)
		canvas.DrawStringAnchored(strconv.Itoa(pice), 10+nameW+10+numberW+50+priceW/2, textDy+textH/2, 0.5, 0.5)
	}
	return canvas.Image(), nil
}
