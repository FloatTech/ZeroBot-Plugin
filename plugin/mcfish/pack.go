// Package mcfish 钓鱼模拟器
package mcfish

import (
	"errors"
	"image"
	"image/color"
	"strconv"
	"sync"

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
		pic, err := drawPackImage(equipInfo, articles)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR at pack.go.3]:", err))
			return
		}
		ctx.SendChain(message.ImageBytes(pic))
	})
}

func drawPackImage(equipInfo equip, articles []article) (imagePicByte []byte, err error) {
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
			packBlock, err = drawArticleInfoBlock(articles, fontdata)
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
	backDY := 10 + equipBlock.Bounds().Dy() + 10 + packBlock.Bounds().Dy()
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

	equipPic, err := imgfactory.Load(engine.DataFolder() + equipInfo.Equip + ".png")
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
	valueW, _ := canvas.MeasureString(durable)
	barW := 1000 - textDx - textW - 10 - valueW - 10

	canvas.DrawStringAnchored("装备耐久", textDx+textW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawRectangle(textDx+textW+5, textDy, barW, textH*1.2)
	canvas.SetRGB255(150, 150, 150)
	canvas.Fill()
	canvas.SetRGB255(0, 0, 0)
	durableW := barW * float64(equipInfo.Durable) / float64(equipAttribute[equipInfo.Equip])
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
func drawArticleInfoBlock(articles []article, fontdata []byte) (image.Image, error) {
	canvas := gg.NewContext(1, 1)
	err := canvas.ParseFontFace(fontdata, 100)
	if err != nil {
		return nil, err
	}
	titleW, titleH := canvas.MeasureString("背包信息")
	err = canvas.ParseFontFace(fontdata, 50)
	if err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("高度")

	nameW := 0.0
	valueW := 0.0
	for _, info := range articles {
		textW, _ := canvas.MeasureString(info.Name + "(" + info.Other + ")")
		if nameW < textW {
			nameW = textW
		}
		textW, _ = canvas.MeasureString(strconv.Itoa(info.Number))
		if valueW < textW {
			valueW = textW
		}
	}

	bolckW := int(10 + nameW + 10 + valueW + 10)
	wallW := (1000 - bolckW*2 - 20) / 2
	backY := 10 + int(titleH*1.6) + 10 + int(textH*2)*(math.Ceil(len(articles), 2)+1)
	canvas = gg.NewContext(1000, math.Max(backY, 300))
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
	if err = canvas.ParseFontFace(fontdata, 50); err != nil {
		return nil, err
	}
	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored("名称", float64(wallW)+10+nameW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("数量", float64(wallW)+10+nameW+10+valueW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("名称", float64(wallW)+float64(bolckW)+20+nameW/2, textDy+textH/2, 0.5, 0.5)
	canvas.DrawStringAnchored("数量", float64(wallW)+float64(bolckW)+20+nameW+10+valueW/2, textDy+textH/2, 0.5, 0.5)
	cell := 0
	textDy += textH * 2
	for _, info := range articles {
		name := info.Name
		if info.Other != "" {
			name += "(" + info.Other + ")"
		}
		infoNameW, _ := canvas.MeasureString(name)
		valueStr := strconv.Itoa(info.Number)
		infoValueW, _ := canvas.MeasureString(valueStr)
		if cell == 2 {
			cell = 0
			textDy += textH * 2
		}
		canvas.DrawStringAnchored(name, float64(wallW)+float64((10+bolckW)*cell)+10+infoNameW/2, textDy+textH/2, 0.5, 0.5)
		canvas.DrawStringAnchored(valueStr, float64(wallW)+float64((10+bolckW)*cell)+10+nameW+10+infoValueW/2, textDy+textH/2, 0.5, 0.5)
		cell++
	}
	return canvas.Image(), nil
}
