// Package mcfish 钓鱼模拟器
package mcfish

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"strconv"
	"strings"
	"sync"

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

func init() {
	engine.OnFullMatch("钓鱼背包", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		equipInfo, err := dbdata.getUserEquip(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pack.go.1]:", err))
			return
		}
		articles, err := dbdata.getUserPack(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pack.go.2]:", err))
			return
		}
		pic, err := drawPackImage(uid, equipInfo, articles)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pack.go.3]:", err))
			return
		}
		ctx.SendChain(message.ImageBytes(pic))
	})
	engine.OnRegex(`^消除绑定诅咒(\d*)$`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		number, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
		if number == 0 {
			number = 1
		}
		number1, err := dbdata.getNumberFor(uid, "宝藏诅咒")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.3.1]:", err))
			return
		}
		if number1 == 0 {
			ctx.SendChain(message.Text("你没有绑定任何诅咒"))
			return
		}
		if number1 < number {
			number = number1
		}
		number2, err := dbdata.getNumberFor(uid, "净化书")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.3.2]:", err))
			return
		}
		if number2 < number {
			ctx.SendChain(message.Text("你没有足够的解除诅咒的道具"))
			return
		}
		articles, err := dbdata.getUserThingInfo(uid, "净化书")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.3.3]:", err))
			return
		}
		articles[0].Number -= number
		err = dbdata.updateUserThingInfo(uid, articles[0])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.3.4]:", err))
			return
		}
		articles, err = dbdata.getUserThingInfo(uid, "宝藏诅咒")
		if err != nil {
			ctx.SendChain(message.Text("消除失败,净化书销毁了\n[ERROR at store.go.3.5]:", err))
			return
		}
		articles[0].Number -= number
		err = dbdata.updateUserThingInfo(uid, articles[0])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at store.go.3.5]:", err))
			return
		}
		ctx.SendChain(message.Text("消除成功"))
	})
	engine.OnFullMatch("当前装备概率明细", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		equipInfo, err := dbdata.getUserEquip(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pack.go.1]:", err))
			return
		}
		number, err := dbdata.getNumberFor(uid, "鱼")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at fish.go.5.1]:", err))
			return
		}
		msg := make(message.Message, 0, 20+len(thingList))
		msg = append(msg, message.At(uid), message.Text("\n大类概率:\n"))
		probableList := make([]int, 4)
		for _, info := range articlesInfo.ZoneInfo {
			switch info.Name {
			case "treasure":
				probableList[0] = info.Probability
			case "pole":
				probableList[1] = info.Probability
			case "fish":
				probableList[2] = info.Probability
			case "waste":
				probableList[3] = info.Probability
			}
		}
		if number > 100 || equipInfo.Equip == "美西螈" { // 放大概率
			probableList = []int{2, 8, 35, 45}
		}
		if equipInfo.Favor > 0 {
			probableList[0] += equipInfo.Favor
			probableList[1] += equipInfo.Favor
			probableList[2] += equipInfo.Favor
			probableList[3] -= equipInfo.Favor * 3
		}
		probable := probableList[0]
		msg = append(msg, message.Text("宝藏 : ", probableList[0], "%\n"))
		probable += probableList[1]
		msg = append(msg, message.Text("鱼竿 : ", probableList[1], "%\n"))
		probable += probableList[2]
		msg = append(msg, message.Text("鱼类 : ", probableList[2], "%\n"))
		probable += probableList[3]
		msg = append(msg, message.Text("垃圾 : ", probableList[3], "%\n"))
		msg = append(msg, message.Text("合计 : ", probable, "%\n"))
		msg = append(msg, message.Text("-----------\n宝藏概率:\n"))
		for _, name := range treasureList {
			msg = append(msg, message.Text(name, " : ",
				strconv.FormatFloat(float64(probabilities[name].Max-probabilities[name].Min)*float64(probableList[0])/100, 'f', 2, 64),
				"%\n"))
		}
		msg = append(msg, message.Text("-----------\n鱼竿概率:\n"))
		for _, name := range poleList {
			if name != "美西螈" {
				msg = append(msg, message.Text(name, " : ",
					strconv.FormatFloat(float64(probabilities[name].Max-probabilities[name].Min)*float64(probableList[1])/100, 'f', 2, 64),
					"%\n"))
			} else if name == "美西螈" {
				msg = append(msg, message.Text(name, " : ",
					strconv.FormatFloat(float64(probabilities[name].Max-probabilities[name].Min)*float64(probableList[0])/100, 'f', 2, 64),
					"%\n"))
			}
		}
		msg = append(msg, message.Text("-----------\n鱼类概率:\n"))
		for _, name := range fishList {
			if name != "海豚" {
				msg = append(msg, message.Text(name, " : ",
					strconv.FormatFloat(float64(probabilities[name].Max-probabilities[name].Min)*float64(probableList[2])/100, 'f', 2, 64),
					"%\n"))
			} else if name == "海豚" {
				msg = append(msg, message.Text(name, " : ",
					strconv.FormatFloat(float64(probabilities[name].Max-probabilities[name].Min)*float64(probableList[0])/100, 'f', 2, 64),
					"%\n"))
			}
		}
		msg = append(msg, message.Text("-----------"))
		ctx.Send(msg)
	})
}

func drawPackImage(uid int64, equipInfo equip, articles []article) (imagePicByte []byte, err error) {
	fontdata, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	var (
		wg         sync.WaitGroup
		equipBlock image.Image // 装备信息
		packBlock  image.Image // 背包信息
	)
	wg.Add(1)
	// 绘制ID
	go func() {
		defer wg.Done()
		if equipInfo == (equip{}) {
			equipBlock, err = drawEquipEmptyBlock(fontdata)
		} else {
			equipBlock, err = drawEquipInfoBlock(equipInfo, fontdata)
		}
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制基本信息
	go func() {
		defer wg.Done()
		if len(articles) == 0 {
			packBlock, err = drawArticleEmptyBlock(fontdata)
		} else {
			packBlock, err = drawArticleInfoBlock(uid, articles, fontdata)
		}
		if err != nil {
			return
		}
	}()
	wg.Wait()
	if equipBlock == nil || packBlock == nil {
		err = errors.New("生成图片失败,数据缺失")
		return
	}
	// 计算图片高度
	backDX := 1020
	backDY := 10 + equipBlock.Bounds().Dy() + 10 + packBlock.Bounds().Dy() + 10
	canvas := gg.NewContext(backDX, backDY)

	// 画底色
	canvas.DrawRectangle(0, 0, float64(backDX), float64(backDY))
	canvas.SetRGBA255(150, 150, 150, 255)
	canvas.Fill()
	canvas.DrawRectangle(10, 10, float64(backDX-20), float64(backDY-20))
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	canvas.DrawImage(equipBlock, 10, 10)
	canvas.DrawImage(packBlock, 10, 10+equipBlock.Bounds().Dy()+10)

	return imgfactory.ToBytes(canvas.Image())
}

// 绘制装备栏区域
func drawEquipEmptyBlock(fontdata []byte) (image.Image, error) {
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
	err := canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString("装备信息")
	canvas.DrawString("装备信息", 10, 10+textH)
	canvas.DrawLine(10, textH*1.2, textW, textH*1.2)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	if err = canvas.ParseFontFace(fontdata, 50); err != nil {
		return nil, err
	}
	canvas.DrawString("没有装备任何鱼竿", 50, 10+textH*2+50)
	return canvas.Image(), nil
}
func drawEquipInfoBlock(equipInfo equip, fontdata []byte) (image.Image, error) {
	canvas := gg.NewContext(1, 1)
	err := canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	_, titleH := canvas.MeasureString("装备信息")
	err = canvas.ParseFontFace(fontdata, 50)
	if err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("装备信息")

	backDY := math.Max(int(10+titleH*2+(textH*2)*4+10), 300)

	canvas = gg.NewContext(1000, backDY)
	// 画底色
	canvas.DrawRectangle(0, 0, 1000, float64(backDY))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 边框框
	canvas.DrawRectangle(0, 0, 1000, float64(backDY))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	getAvatar, err := engine.GetLazyData(equipInfo.Equip+".png", false)
	if err != nil {
		return nil, err
	}
	equipPic, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return nil, err
	}
	picDy := float64(backDY) - 10 - titleH*2
	equipPic = imgfactory.Size(equipPic, int(picDy)-10, int(picDy)-10).Image()
	canvas.DrawImage(equipPic, 10, 10+int(titleH)*2)

	// 放字
	canvas.SetColor(color.Black)
	if err = canvas.ParseFontFace(fontdata, 100); err != nil {
		return nil, err
	}
	titleW, titleH := canvas.MeasureString("装备信息")
	canvas.DrawString("装备信息", 10, 10+titleH*1.2)
	canvas.DrawLine(10, titleH*1.6, titleW, titleH*1.6)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	textDx := picDy + 10
	textDy := 10 + titleH*2
	if err = canvas.ParseFontFace(fontdata, 75); err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString(equipInfo.Equip)
	canvas.DrawStringAnchored(equipInfo.Equip, textDx+textW/2, textDy+textH/2, 0.5, 0.5)

	textDy += textH * 1.5
	if err = canvas.ParseFontFace(fontdata, 50); err != nil {
		return nil, err
	}
	textW, textH = canvas.MeasureString("维修次数")
	durable := strconv.Itoa(equipInfo.Durable)
	valueW, _ := canvas.MeasureString("100")
	barW := 1000 - textDx - textW - 10 - valueW - 10

	canvas.DrawStringAnchored("装备耐久", textDx+textW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawRectangle(textDx+textW+5, textDy, barW, textH*1.2)
	canvas.SetRGB255(150, 150, 150)
	canvas.Fill()
	canvas.SetRGB255(0, 0, 0)
	durableW := barW * float64(equipInfo.Durable) / float64(durationList[equipInfo.Equip])
	canvas.DrawRectangle(textDx+textW+5, textDy, durableW, textH*1.2)
	canvas.SetRGB255(102, 102, 102)
	canvas.Fill()
	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored(durable, textDx+textW+5+barW+5+valueW/2, textDy+textH/2, 0.5, 0.5)

	textDy += textH * 2
	maintenance := strconv.Itoa(equipInfo.Maintenance)
	canvas.DrawStringAnchored("维修次数", textDx+textW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawRectangle(textDx+textW+5, textDy, barW, textH*1.2)
	canvas.SetRGB255(150, 150, 150)
	canvas.Fill()
	canvas.SetRGB255(0, 0, 0)
	canvas.DrawRectangle(textDx+textW+5, textDy, barW*float64(equipInfo.Maintenance)/10, textH*1.2)
	canvas.SetRGB255(102, 102, 102)
	canvas.Fill()
	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored(maintenance, textDx+textW+5+barW+5+valueW/2, textDy+textH/2, 0.5, 0.5)

	textDy += textH * 3
	canvas.DrawString(" 附魔: 诱钓"+enchantLevel[equipInfo.Induce]+"  海之眷顾"+enchantLevel[equipInfo.Favor], textDx, textDy)
	return canvas.Image(), nil
}

// 绘制背包信息区域
func drawArticleEmptyBlock(fontdata []byte) (image.Image, error) {
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
	err := canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString("背包信息")
	canvas.DrawString("背包信息", 10, 10+textH*1.2)
	canvas.DrawLine(10, textH*1.6, textW, textH*1.6)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	if err = canvas.ParseFontFace(fontdata, 50); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored("背包没有存放任何东西", 500, 10+textH*2+50, 0.5, 0)
	return canvas.Image(), nil
}
func drawArticleInfoBlock(uid int64, articles []article, fontdata []byte) (image.Image, error) {
	canvas := gg.NewContext(1, 1)
	err := canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	titleW, titleH := canvas.MeasureString("背包信息")
	front := 45.0
	err = canvas.ParseFontFace(fontdata, front)
	if err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("高度")

	nameWOfFiest := 0.0
	nameWOfSecond := 0.0
	for i, info := range articles {
		textW, _ := canvas.MeasureString(info.Name + "(" + info.Other + ")")
		if i%2 == 0 && textW > nameWOfFiest {
			nameWOfFiest = textW
		} else if textW > nameWOfSecond {
			nameWOfSecond = textW
		}
	}
	valueW, _ := canvas.MeasureString("10000")

	if (10+nameWOfFiest+10+valueW+10)+(10+nameWOfSecond+10+valueW+10) > 980 {
		front = 32.0
		err = canvas.ParseFontFace(fontdata, front)
		if err != nil {
			return nil, err
		}
		_, textH = canvas.MeasureString("高度")

		nameWOfFiest = 0
		nameWOfSecond = 0
		for i, info := range articles {
			textW, _ := canvas.MeasureString(info.Name + "(" + info.Other + ")")
			if i%2 == 0 && textW > nameWOfFiest {
				nameWOfFiest = textW
			} else if textW > nameWOfSecond {
				nameWOfSecond = textW
			}
		}
		valueW, _ = canvas.MeasureString("10000")
	}
	wallW := (980 - (10 + nameWOfFiest + 10 + valueW + 10) - (10 + nameWOfSecond + 10 + valueW + 10)) / 2
	backY := math.Max(10+int(titleH*1.6)+10+int(textH*2)*(math.Ceil(len(articles), 2)+1), 500)
	canvas = gg.NewContext(1000, backY)
	// 画底色
	canvas.DrawRectangle(0, 0, 1000, float64(backY))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 边框框
	canvas.DrawRectangle(0, 0, 1000, float64(backY))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	// 放字
	canvas.SetColor(color.Black)
	err = canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	canvas.DrawString("背包信息", 10, 10+titleH*1.2)
	canvas.DrawLine(10, titleH*1.6, titleW, titleH*1.6)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	textDy := 10 + titleH*1.7
	if err = canvas.ParseFontFace(fontdata, front); err != nil {
		return nil, err
	}
	canvas.SetColor(color.Black)
	numberOfFish := 0
	numberOfEquip := 0
	canvas.DrawStringAnchored("名称", wallW+20+nameWOfFiest/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("数量", wallW+20+nameWOfFiest+10+valueW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("名称", wallW+20+nameWOfFiest+10+valueW+10+10+nameWOfSecond/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("数量", wallW+20+nameWOfFiest+10+valueW+10+10+nameWOfSecond+10+valueW/2, textDy+textH/2, 0.5, 0.5)
	textDy += textH * 2
	for i, info := range articles {
		name := info.Name
		if info.Other != "" {
			if strings.Contains(info.Name, "竿") {
				numberOfEquip++
			}
			name += "(" + info.Other + ")"
		} else if strings.Contains(name, "鱼") {
			numberOfFish += info.Number
		}
		valueStr := strconv.Itoa(info.Number)
		if i%2 == 0 {
			if i != 0 {
				textDy += textH * 2
			}
			canvas.DrawStringAnchored(name, wallW+20+nameWOfFiest/2, textDy+textH/2, 0.5, 0.5)
			canvas.DrawStringAnchored(valueStr, wallW+20+nameWOfFiest+10+valueW/2, textDy+textH/2, 0.5, 0.5)
		} else {
			canvas.DrawStringAnchored(name, wallW+20+nameWOfFiest+10+valueW+10+10+nameWOfSecond/2, textDy+textH/2, 0.5, 0.5)
			canvas.DrawStringAnchored(valueStr, wallW+20+nameWOfFiest+10+valueW+10+10+nameWOfSecond+10+valueW/2, textDy+textH/2, 0.5, 0.5)
		}
	}
	if err = canvas.ParseFontFace(fontdata, 30); err != nil {
		return nil, err
	}
	textDy = 10
	text := "钱包余额: " + strconv.Itoa(wallet.GetWalletOf(uid))
	textW, textH := canvas.MeasureString(text)
	w, _ := canvas.MeasureString("维修大师[已激活]")
	if w > textW {
		textW = w
	}
	canvas.DrawStringAnchored(text, 980-textW/2, textDy+textH/2, 0.5, 0.5)
	textDy += textH * 1.5
	if numberOfFish > 100 {
		canvas.DrawStringAnchored("钓鱼佬[已激活]", 980-textW/2, textDy+textH/2, 0.5, 0.5)
		textDy += textH * 1.5
	}
	if numberOfEquip > 10 {
		canvas.DrawStringAnchored("维修大师[已激活]", 980-textW/2, textDy+textH/2, 0.5, 0.5)
	}
	return canvas.Image(), nil
}
