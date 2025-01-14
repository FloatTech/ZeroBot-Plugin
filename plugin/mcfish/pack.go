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
	engine.OnFullMatch("查看钓鱼规则", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		msg := "一款钓鱼模拟器\n----------指令----------\n" +
			"- 钓鱼看板/钓鱼商店\n- 购买xxx\n- 购买xxx [数量]\n- 出售xxx\n- 出售xxx [数量]\n- 出售所有垃圾\n" +
			"- 钓鱼背包\n- 装备[xx竿|三叉戟|美西螈]\n- 附魔[诱钓|海之眷顾]\n- 修复鱼竿\n- 合成[xx竿|三叉戟]\n- 消除[绑定|宝藏]诅咒\n- 消除[绑定|宝藏]诅咒 [数量]\n" +
			"- 进行钓鱼\n- 进行n次钓鱼\n- " +
			"当前装备概率明细\n" +
			"规则V" + version + ":\n" +
			"1.每日的商店价格是波动的!!如何最大化收益自己考虑一下喔\n" +
			"2.装备信息:\n-> 木竿 : 耐久上限:30 均价:100 上钩概率:0.7%\n-> 铁竿 : 耐久上限:50 均价:300 上钩概率:0.2%\n-> 金竿 : 耐久上限:70 均价700 上钩概率:0.06%\n" +
			"-> 钻石竿 : 耐久上限:100 均价1500 上钩概率:0.03%\n-> 下界合金竿 : 耐久上限:150 均价3100 上钩概率:0.01%\n-> 三叉戟 : 可使1次钓鱼视为3次钓鱼. 耐久上限:300 均价4000 只能合成、修复和交易\n" +
			"3.附魔书信息:\n-> 诱钓 : 减少上钩时间. 均价:1000, 上钩概率:0.25%\n-> 海之眷顾 : 增加宝藏上钩概率. 均价:2500, 上钩概率:0.10%\n" +
			"4.稀有物品:\n-> 唱片 : 出售物品时使用该物品使价格翻倍. 均价:3000, 上钩概率:0.01%\n" +
			"-> 美西螈 : 可装备,获得隐形[钓鱼佬]buff,并让钓到除鱼竿和美西螈外的物品数量变成5,无耐久上限.不可修复/附魔,每次钓鱼消耗3条鱼. 均价:3000, 上钩概率:0.01%\n" +
			"-> 海豚 : 使空竿概率变成垃圾概率. 均价:1000, 上钩概率:0.19%\n" +
			"-> 宝藏诅咒 : 无法交易,每一层就会增加购买时10%价格和减少出售时10%价格(超过10层会变为倒贴钱). 上钩概率:0.25%\n-> 净化书 : 用于消除宝藏诅咒. 均价:5000, 上钩概率:0.19%\n" +
			"5.鱼类信息:\n-> 鳕鱼 : 均价:10 上钩概率:0.69%\n-> 鲑鱼 : 均价:50 上钩概率:0.2%\n-> 热带鱼 : 均价:100 上钩概率:0.06%\n-> 河豚 : 均价:300 上钩概率:0.03%\n-> 鹦鹉螺 : 均价:500 上钩概率:0.01%\n-> 墨鱼 : 均价:500 上钩概率:0.01%\n" +
			"6.垃圾:\n-> 均价:10 上钩概率:30%\n" +
			"7.物品BUFF:\n-> 钓鱼佬 : 当背包名字含有'鱼'的物品数量超过100时激活,钓到物品概率提高至90%\n-> 修复大师 : 当背包鱼竿数量超过10时激活,修复物品时耐久百分百继承\n" +
			"8.合成:\n-> 铁竿 : 3x木竿\n-> 金竿 : 3x铁竿\n-> 钻石竿 : 3x金竿\n-> 下界合金竿 : 3x钻石竿\n-> 三叉戟 : 3x下界合金竿\n注:合成成功率90%(包括梭哈),合成鱼竿的附魔等级=（附魔等级合/合成鱼竿数量）\n" +
			"9.杂项:\n-> 无装备的情况下,每人最多可以购买3次100块钱的鱼竿\n-> 默认状态钓鱼上钩概率为60%(理论值!!!)\n-> 附魔的鱼竿会因附魔变得昂贵,每个附魔最高3级\n-> 三叉戟不算鱼竿,修复时可直接满耐久\n" +
			"-> 鱼竿数量大于50的不能买东西;\n     鱼竿数量大于30的不能钓鱼;\n     每购/售10次鱼竿获得1层宝藏诅咒;\n     每购买20次物品将获得3次价格减半福利;\n     每钓鱼75次获得1本净化书;\n" +
			"     每天可交易鱼竿10个，购买物品30件（垃圾除外）."

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
