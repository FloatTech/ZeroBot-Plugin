// Package coc coc插件
package coc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/rendercard"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnPrefixGroup([]string{".coc", "。coc", ".COC"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + DefaultJsonFile
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Text("你群还没有布置coc,请相关人员后台布局coc.(详情看用法)"))
			return
		}
		var (
			cocInfo cocJSON
			err     error
		)
		if file.IsNotExist(engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".json") {
			cocInfo, err = loadPanel(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			cocInfo.ID = uid
			baseMsg := strings.Split(ctx.State["args"].(string), "/")
			for _, msgInfo := range baseMsg {
				msgValue := strings.Split(msgInfo, "#")
				if msgValue[0] == "描述" {
					cocInfo.Other = append(cocInfo.Other, msgValue[1])
					continue
				}
				for i, info := range cocInfo.BaseInfo {
					if msgValue[0] == info.Name {
						munberValue, err := strconv.Atoi(msgValue[1])
						if err != nil {
							cocInfo.BaseInfo[i].Value = msgValue[1]
						} else {
							cocInfo.BaseInfo[i].Value = munberValue
						}
					}
				}
			}
			for i, info := range cocInfo.Attribute {
				max := info.MaxValue - info.MinValue
				negative := -1
				if info.MinValue < 0 {
					negative = 1
				}
				value := 0
				for i := 0; i < 3; i++ {
					value += rand.Intn(6) + 1
				}
				value = max*value*5/100 + negative*info.MinValue
				cocInfo.Attribute[i].Value = value
			}
			err = savePanel(cocInfo, gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		} else {
			cocInfo, err = loadPanel(gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		}
		pic, err := drawImage(cocInfo)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.ImageBytes(pic))
	})
}

func drawImage(userInfo cocJSON) (imagePicByte []byte, err error) {
	var (
		wg             sync.WaitGroup
		userIDBlock    image.Image
		infoBlock      image.Image
		atrrBlock      image.Image
		otherBlock     image.Image
		halfSkillBlock image.Image
	)
	wg.Add(1)
	// 绘制ID
	go func() {
		defer wg.Done()
		userIDBlock, err = userInfo.drawIDBlock()
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制基本信息
	go func() {
		defer wg.Done()
		infoBlock, err = userInfo.drawInfoBlock()
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制属性框
	go func() {
		defer wg.Done()
		atrrBlock, err = userInfo.drawAttrBlock()
		if err != nil {
			return
		}
	}()
	wg.Add(1)
	// 绘制技能框
	go func() {
		defer wg.Done()
		otherBlock, halfSkillBlock, err = userInfo.drawOtherBlock()
		if err != nil {
			return
		}
	}()
	wg.Wait()
	if userIDBlock == nil || infoBlock == nil || atrrBlock == nil || otherBlock == nil || halfSkillBlock == nil {
		return
	}
	// 计算图片高度
	choose := 0
	picDY := 1000
	backDX := 1020
	backDY := 1020
	switch {
	case 25+userIDBlock.Bounds().Dy()+25+infoBlock.Bounds().Dy()+atrrBlock.Bounds().Dy()+50 > 1000:
		choose = 1
		picDY = 25 + userIDBlock.Bounds().Dy() + 25 + infoBlock.Bounds().Dy() + atrrBlock.Bounds().Dy() + 50
		backDY = 10 + picDY + 50 + otherBlock.Bounds().Dy() + 10
	case 25+userIDBlock.Bounds().Dy()+25+infoBlock.Bounds().Dy()+atrrBlock.Bounds().Dy()+25+halfSkillBlock.Bounds().Dy()+50 > 1000:
		choose = 2
		backDY = 10 + picDY + 10 + otherBlock.Bounds().Dy() + 10
	}
	canvas := gg.NewContext(backDX, backDY)

	// 画底色
	canvas.DrawRectangle(0, 0, float64(backDX), float64(backDY))
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	// 放置头像
	getAvatar, err := web.GetData("http://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(userInfo.ID, 10) + "&s=640")
	if err != nil {
		return
	}
	avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return
	}
	avatarf := imgfactory.Size(avatar, picDY, picDY)
	canvas.DrawImageAnchored(avatarf.Blur(10).Image(), 10+picDY/2, 10+picDY/2, 0.5, 0.5)
	canvas.DrawImageAnchored(avatarf.Clip(picDY/2, picDY, 0, 0).Image(), 10+picDY/4, 10+picDY/2, 0.5, 0.5)

	// 头像框
	canvas.DrawRectangle(10, 10, float64(picDY), float64(picDY))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()

	// 放入信息
	tempDY := 20
	canvas.DrawImageAnchored(userIDBlock, backDX-10-20-userIDBlock.Bounds().Dx()/2, tempDY+userIDBlock.Bounds().Dy()/2, 0.5, 0.5)
	tempDY = 35 + userIDBlock.Bounds().Dy() + 25
	canvas.DrawImageAnchored(infoBlock, backDX-10-20-infoBlock.Bounds().Dx()/2, tempDY+infoBlock.Bounds().Dy()/2, 0.5, 0.5)
	tempDY = 35 + userIDBlock.Bounds().Dy() + 25 + infoBlock.Bounds().Dy()
	canvas.DrawImageAnchored(atrrBlock, backDX-10-20-atrrBlock.Bounds().Dx()/2, tempDY+atrrBlock.Bounds().Dy()/2, 0.5, 0.5)
	if choose == 0 {
		tempDY = 35 + userIDBlock.Bounds().Dy() + 25 + infoBlock.Bounds().Dy() + atrrBlock.Bounds().Dy() + 10
		canvas.DrawImageAnchored(halfSkillBlock, backDX-10-20-halfSkillBlock.Bounds().Dx()/2, tempDY+halfSkillBlock.Bounds().Dy()/2, 0.5, 0.5)
	} else {
		canvas.DrawImageAnchored(otherBlock, 10+otherBlock.Bounds().Dx()/2, 10+picDY+10+otherBlock.Bounds().Dy()/2, 0.5, 0.5)
	}

	return imgfactory.ToBytes(canvas.Image())
}

// 绘制ID区域
func (userInfo *cocJSON) drawIDBlock() (image.Image, error) {
	fontSize := 25.0
	userIDstr := "编号:" + strconv.FormatInt(userInfo.ID, 10)
	canvas := gg.NewContext(1, 1)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	textW, textH := canvas.MeasureString(userIDstr)
	canvas = gg.NewContext(int(textW*1.2), int(textH)*2+15)
	// 放字
	canvas.SetColor(color.Black)
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored(userIDstr, textW*0.6, textH, 0.5, 0.5)
	// 画下划线
	canvas.DrawLine(0, textH*2, textW*1.2, textH*2)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	return canvas.Image(), nil
}

// 绘制基本信息区域
func (userInfo *cocJSON) drawInfoBlock() (image.Image, error) {
	fontSize := 50.0
	raw := len(userInfo.BaseInfo)
	canvas := gg.NewContext(1, 1)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("高度")
	canvas = gg.NewContext(500-40, 50+int(textH*2)*raw)
	// 画底色
	canvas.DrawRectangle(0, 0, 500-40, textH*2*float64(raw)+50)
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 放字
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	for i, info := range userInfo.BaseInfo {
		str := fmt.Sprintf("%v : %v", info.Name, info.Value)
		textW, _ := canvas.MeasureString(str)
		textDY := 50.0 + textH*2*float64(i)
		canvas.SetColor(color.Black)
		canvas.DrawStringAnchored(str, 25+textW/2, textDY+textH/2, 0.5, 0.5)
		textDY += textH * 1.5
		// 画下划线
		canvas.DrawLine(25, textDY, 500-40-25, textDY)
		canvas.SetLineWidth(3)
		canvas.SetRGBA255(0, 0, 0, 255)
		canvas.Stroke()
	}
	return canvas.Image(), nil
}

// 绘制属性信息区域
func (userInfo *cocJSON) drawAttrBlock() (image.Image, error) {
	fontSize := 25.0
	raw := len(userInfo.Attribute)
	canvas := gg.NewContext(1, 1)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	_, textH := canvas.MeasureString("高度")
	offset := 0.0
	median := 0.0
	for _, info := range userInfo.Attribute {
		textW, _ := canvas.MeasureString(strconv.Itoa(info.MaxValue))
		if offset < textW {
			offset = textW
		}
	}
	for _, info := range userInfo.Attribute {
		textW, _ := canvas.MeasureString(info.Name)
		if median < textW {
			median = textW
		}
	}
	median = (500 - 40 - median - 10 - 300 - 10 - offset) / 2
	canvas = gg.NewContext(500-40, 50+int(textH*2)*raw)
	// 画底色
	canvas.DrawRectangle(0, 0, 500-40, textH*2*float64(raw)+50)
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 放字
	if err = canvas.ParseFontFace(data, fontSize); err != nil {
		return nil, err
	}
	for i, info := range userInfo.Attribute {
		nameW, _ := canvas.MeasureString(info.Name)
		valueStr := strconv.Itoa(info.Value)
		valueW, _ := canvas.MeasureString(valueStr)
		textDY := 25.0 + textH*2*float64(i)
		canvas.SetColor(color.Black)
		// 名称
		canvas.DrawStringAnchored(info.Name, 500-20-median-offset-10-300-10-nameW/2, textDY+textH/2, 0.5, 0.5)
		// 数值
		canvas.DrawStringAnchored(valueStr, 500-20-median-offset/2-valueW/2, textDY+textH/2, 0.5, 0.5)
		// 画属性条
		canvas.DrawRectangle(170-median-offset, textDY, 300, textH*1.2)
		canvas.SetRGB255(150, 150, 150)
		canvas.Fill()
		canvas.SetRGB255(0, 0, 0)
		canvas.DrawRectangle(170-median-offset, textDY, 300*(math.Abs(float64(info.Value))/math.Max(float64(info.MaxValue), float64(info.MaxValue)-float64(info.MinValue))), textH*1.2)
		canvas.SetRGB255(102, 102, 102)
		canvas.Fill()
	}
	return canvas.Image(), nil
}

// 绘制其他信息区域
func (userInfo *cocJSON) drawOtherBlock() (pic, halfPic image.Image, err error) {
	fontSize := 35.0
	glowsd, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	newplugininfo, err := rendercard.Truncate(glowsd, userInfo.Other, 500-60, 50)
	if err != nil {
		return
	}
	canvas := gg.NewContext(1, 1)
	if err = canvas.ParseFontFace(glowsd, fontSize); err != nil {
		return
	}
	_, textH := canvas.MeasureString("高度")
	raw := 50 + int(textH*2)*len(newplugininfo)
	if raw < 200 {
		raw = 200
	}
	canvas = gg.NewContext(500-40, raw)
	// 画底色
	canvas.DrawRectangle(0, 0, 500-40, float64(raw))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 放字
	if err = canvas.ParseFontFace(glowsd, fontSize); err != nil {
		return
	}
	canvas.SetColor(color.Black)
	for i, info := range newplugininfo {
		if info == "" {
			info = "暂无信息"
		}
		textW, _ := canvas.MeasureString(info)
		textDY := 25.0 + textH*2*float64(i)
		canvas.DrawStringAnchored(info, 10+textW/2, textDY+textH/2, 0.5, 0.5)
	}
	halfPic = canvas.Image()
	// 绘制整个图
	newplugininfo, err = rendercard.Truncate(glowsd, userInfo.Other, 900, 50)
	if err != nil {
		return
	}
	raw = 50 + int(textH*2)*len(newplugininfo)
	if raw < 200 {
		raw = 200
	}
	canvas = gg.NewContext(1, 1)
	if err = canvas.ParseFontFace(glowsd, fontSize); err != nil {
		return
	}
	_, textH = canvas.MeasureString("高度")
	canvas = gg.NewContext(1000, raw)
	// 画底色
	canvas.DrawRectangle(0, 0, 1000, float64(raw))
	canvas.SetRGBA255(255, 255, 255, 150)
	canvas.Fill()
	// 放字
	if err = canvas.ParseFontFace(glowsd, fontSize); err != nil {
		return
	}
	canvas.SetColor(color.Black)
	for i, info := range newplugininfo {
		if info == "" {
			info = "暂无信息"
		}
		textW, _ := canvas.MeasureString(info)
		textDY := 25.0 + textH*2*float64(i)
		canvas.DrawStringAnchored(info, 10+textW/2, textDY+textH/2, 0.5, 0.5)
	}
	// 边框框
	canvas.DrawRectangle(0, 0, 1000, float64(raw))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.Stroke()
	pic = canvas.Image()
	return
}
