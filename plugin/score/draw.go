// Package score 签到，答题得分
package score

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/disintegration/imaging"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
)

func drawScore16(a *scdata) (image.Image, error) {
	// 绘图
	getAvatar, err := initPic(a.picfile, a.uid)
	if err != nil {
		return nil, err
	}
	back, err := gg.LoadImage(a.picfile)
	if err != nil {
		return nil, err
	}
	// 避免图片过大，最大 1280*720
	back = imgfactory.Limit(back, 1280, 720)
	imgDX := back.Bounds().Dx()
	imgDY := back.Bounds().Dy()
	canvas := gg.NewContext(imgDX, imgDY)
	// draw Aero Style
	aeroStyle := gg.NewContext(imgDX-202, imgDY-202)
	aeroStyle.DrawImage(imaging.Blur(back, 2.5), -100, -100)
	// aero draw image.
	aeroStyle.DrawRoundedRectangle(0, 0, float64(imgDX-200), float64(imgDY-200), 16)
	// SideLine
	aeroStyle.SetLineWidth(3)
	aeroStyle.SetRGBA255(255, 255, 255, 100)
	aeroStyle.StrokePreserve()
	aeroStyle.SetRGBA255(255, 255, 255, 140)
	// fill
	aeroStyle.Fill()
	// draw background
	canvas.DrawImage(back, 0, 0)
	// Aero style combine
	canvas.DrawImage(aeroStyle.Image(), 100, 100)
	canvas.Fill()
	hourWord := getHourWord(time.Now())
	avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return nil, err
	}
	avatarf := imgfactory.Size(avatar, 200, 200)
	canvas.DrawImage(avatarf.Circle(0).Image(), 120, 120)
	// draw info(name,coin,etc)
	canvas.SetRGB255(0, 0, 0)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, 50); err != nil {
		return nil, err
	}
	// draw head
	canvas.DrawStringWrapped(a.nickname, 350, 180, 0.5, 0.5, 0.5, 0.5, gg.AlignLeft)
	canvas.Fill()
	// main draw
	data, err = file.GetLazyData(text.FontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored(hourWord, 350, 280, 0, 0)
	canvas.DrawStringAnchored("ATRI币 + "+strconv.Itoa(a.inc), 350, 350, 0, 0)
	canvas.DrawStringAnchored("当前ATRI币："+strconv.Itoa(a.score), 350, 400, 0, 0)
	canvas.DrawStringAnchored("LEVEL: "+strconv.Itoa(getrank(a.level)), 350, 450, 0, 0)
	// draw Info(Time,etc.)
	getTime := time.Now().Format("2006-01-02 15:04:05")
	getTimeLengthWidth, getTimeLengthHight := canvas.MeasureString(getTime)
	canvas.DrawStringAnchored(getTime, float64(imgDX)-100-20-getTimeLengthWidth/2, float64(imgDY)-100-getTimeLengthHight, 0.5, 0.5) // time
	var nextrankScore int
	if a.rank < 10 {
		nextrankScore = rankArray[a.rank+1]
	} else {
		nextrankScore = SCOREMAX
	}
	nextLevelStyle := strconv.Itoa(a.level) + "/" + strconv.Itoa(nextrankScore)
	getLevelLength, _ := canvas.MeasureString(nextLevelStyle)
	canvas.DrawStringAnchored(nextLevelStyle, 100+getLevelLength, float64(imgDY)-100-getTimeLengthHight, 0.5, 0.5) // time
	canvas.Fill()
	canvas.SetRGB255(255, 255, 255)
	if err = canvas.ParseFontFace(data, 20); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored("Created By Zerobot-Plugin "+banner.Version, float64(imgDX)/2, float64(imgDY)-20, 0.5, 0.5) // zbp
	canvas.SetRGB255(0, 0, 0)
	canvas.DrawStringAnchored("Created By Zerobot-Plugin "+banner.Version, float64(imgDX)/2-3, float64(imgDY)-19, 0.5, 0.5) // zbp
	canvas.SetRGB255(255, 255, 255)
	// Gradient
	grad := gg.NewLinearGradient(20, 320, 400, 20)
	grad.AddColorStop(0, color.RGBA{G: 255, A: 255})
	grad.AddColorStop(1, color.RGBA{B: 255, A: 255})
	grad.AddColorStop(0.5, color.RGBA{R: 255, A: 255})
	canvas.SetStrokeStyle(grad)
	canvas.SetLineWidth(4)
	// level array with rectangle work.
	gradLineLength := float64(imgDX-120) - 120
	renderLine := (float64(a.level) / float64(nextrankScore)) * gradLineLength
	canvas.MoveTo(120, float64(imgDY)-102)
	canvas.LineTo(120+renderLine, float64(imgDY)-102)
	canvas.ClosePath()
	canvas.Stroke()
	return canvas.Image(), nil
}

func drawScore15(a *scdata) (image.Image, error) {
	// 绘图
	_, err := initPic(a.picfile, a.uid)
	if err != nil {
		return nil, err
	}
	back, err := gg.LoadImage(a.picfile)
	if err != nil {
		return nil, err
	}
	// 避免图片过大，最大 1280*720
	back = imgfactory.Limit(back, 1280, 720)
	canvas := gg.NewContext(back.Bounds().Size().X, int(float64(back.Bounds().Size().Y)*1.7))
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.DrawImage(back, 0, 0)
	monthWord := time.Now().Format("01/02")
	hourWord := getHourWord(time.Now())
	_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, float64(back.Bounds().Size().X)*0.1); err != nil {
		return nil, err
	}
	canvas.SetRGB(0, 0, 0)
	canvas.DrawString(hourWord, float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.2)
	canvas.DrawString(monthWord, float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*1.2)
	_, err = file.GetLazyData(text.FontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.LoadFontFace(text.FontFile, float64(back.Bounds().Size().X)*0.04); err != nil {
		return nil, err
	}
	canvas.DrawString(a.nickname+fmt.Sprintf(" ATRI币+%d", a.inc), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.3)
	canvas.DrawString("当前ATRI币:"+strconv.FormatInt(int64(a.score), 10), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.4)
	canvas.DrawString("LEVEL:"+strconv.FormatInt(int64(a.rank), 10), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.5)
	canvas.DrawRectangle(float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.55, float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*0.1)
	canvas.SetRGB255(150, 150, 150)
	canvas.Fill()
	var nextrankScore int
	if a.rank < 10 {
		nextrankScore = rankArray[a.rank+1]
	} else {
		nextrankScore = SCOREMAX
	}
	canvas.SetRGB255(0, 0, 0)
	canvas.DrawRectangle(float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.55, float64(back.Bounds().Size().X)*0.6*float64(a.level)/float64(nextrankScore), float64(back.Bounds().Size().Y)*0.1)
	canvas.SetRGB255(102, 102, 102)
	canvas.Fill()
	canvas.DrawString(fmt.Sprintf("%d/%d", a.level, nextrankScore), float64(back.Bounds().Size().X)*0.75, float64(back.Bounds().Size().Y)*1.62)
	return canvas.Image(), nil
}

func drawScore17(a *scdata) (image.Image, error) {
	getAvatar, err := initPic(a.picfile, a.uid)
	if err != nil {
		return nil, err
	}
	back, err := gg.LoadImage(a.picfile)
	if err != nil {
		return nil, err
	}
	// 避免图片过大，最大 1280*720
	back = imgfactory.Limit(back, 1280, 720)
	imgDX := back.Bounds().Dx()
	imgDY := back.Bounds().Dy()
	canvas := gg.NewContext(imgDX, imgDY)

	// draw background
	canvas.DrawImage(back, 0, 0)

	// Create smaller Aero Style boxes
	createAeroBox := func(x, y, width, height float64) {
		aeroStyle := gg.NewContext(int(width), int(height))
		aeroStyle.DrawRoundedRectangle(0, 0, width, height, 8)
		aeroStyle.SetLineWidth(2)
		aeroStyle.SetRGBA255(255, 255, 255, 100)
		aeroStyle.StrokePreserve()
		aeroStyle.SetRGBA255(255, 255, 255, 140)
		aeroStyle.Fill()
		canvas.DrawImage(aeroStyle.Image(), int(x), int(y))
	}

	// draw aero boxes for text
	createAeroBox(20, float64(imgDY-120), 280, 100)               // left bottom
	createAeroBox(float64(imgDX-272), float64(imgDY-60), 252, 40) // right bottom

	// draw info(name, coin, etc)
	hourWord := getHourWord(time.Now())
	canvas.SetRGB255(0, 0, 0)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.ParseFontFace(data, 24); err != nil {
		return nil, err
	}
	getNameLengthWidth, _ := canvas.MeasureString(a.nickname)
	// draw aero box
	if getNameLengthWidth > 140 {
		createAeroBox(20, 40, 140+getNameLengthWidth, 100) // left top
	} else {
		createAeroBox(20, 40, 280, 100) // left top
	}

	// draw avatar
	avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return nil, err
	}
	avatarf := imgfactory.Size(avatar, 100, 100)
	canvas.DrawImage(avatarf.Circle(0).Image(), 30, 20)

	canvas.DrawString(a.nickname, 140, 80)
	canvas.DrawStringAnchored(hourWord, 140, 120, 0, 0)

	if err = canvas.ParseFontFace(data, 20); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored("金币 + "+strconv.Itoa(a.inc), 40, float64(imgDY-90), 0, 0)
	canvas.DrawStringAnchored("当前金币："+strconv.Itoa(a.score), 40, float64(imgDY-60), 0, 0)
	canvas.DrawStringAnchored("LEVEL: "+strconv.Itoa(getrank(a.level)), 40, float64(imgDY-30), 0, 0)

	// Draw Info(Time, etc.)
	getTime := time.Now().Format("2006-01-02 15:04:05")
	canvas.DrawStringAnchored(getTime, float64(imgDX)-146, float64(imgDY)-40, 0.5, 0.5) // time
	var nextrankScore int
	if a.rank < 10 {
		nextrankScore = rankArray[a.rank+1]
	} else {
		nextrankScore = SCOREMAX
	}
	nextLevelStyle := strconv.Itoa(a.level) + "/" + strconv.Itoa(nextrankScore)
	canvas.DrawStringAnchored(nextLevelStyle, 190, float64(imgDY-30), 0, 0) // time

	// Draw Zerobot-Plugin information
	canvas.SetRGB255(255, 255, 255)
	if err = canvas.ParseFontFace(data, 20); err != nil {
		return nil, err
	}
	canvas.DrawStringAnchored("Created By Zerobot-Plugin "+banner.Version, float64(imgDX)/2, float64(imgDY)-20, 0.5, 0.5) // zbp
	canvas.SetRGB255(0, 0, 0)
	canvas.DrawStringAnchored("Created By Zerobot-Plugin "+banner.Version, float64(imgDX)/2-3, float64(imgDY)-19, 0.5, 0.5) // zbp
	canvas.SetRGB255(255, 255, 255)
	return canvas.Image(), nil
}

func drawScore18(a *scdata) (image.Image, error) {
	back, err := gg.LoadImage(a.picfile)
	if err != nil {
		return nil, err
	}
	imgDX := back.Bounds().Dx()
	imgDY := back.Bounds().Dy()
	backDX := 1500

	imgDW := backDX - 100
	scale := float64(imgDW) / float64(imgDX)
	imgDH := int(float64(imgDY) * scale)
	back = imgfactory.Size(back, imgDW, imgDH).Image()

	backDY := imgDH + 500
	canvas := gg.NewContext(backDX, backDY)
	// 放置毛玻璃背景
	backBlurW := float64(imgDW) * (float64(backDY) / float64(imgDH))
	canvas.DrawImageAnchored(imaging.Blur(imgfactory.Size(back, int(backBlurW), backDY).Image(), 8), backDX/2, backDY/2, 0.5, 0.5)
	canvas.DrawRectangle(1, 1, float64(backDX), float64(backDY))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(255, 255, 255, 100)
	canvas.StrokePreserve()
	canvas.SetRGBA255(255, 255, 255, 140)
	canvas.Fill()
	// 信息框
	canvas.DrawRoundedRectangle(20, 20, 1500-20-20, 450-20, (450-20)/5)
	canvas.SetLineWidth(6)
	canvas.SetDash(20.0, 10.0, 0)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Stroke()
	// 放置头像
	getAvatar, err := initPic(a.picfile, a.uid)
	if err != nil {
		return nil, err
	}
	avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return nil, err
	}
	avatarf := imgfactory.Size(avatar, 300, 300)
	canvas.DrawCircle(50+float64(avatarf.W())/2, 50+float64(avatarf.H())/2, float64(avatarf.W())/2+2)
	canvas.SetLineWidth(3)
	canvas.SetDash()
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Stroke()
	canvas.DrawImage(avatarf.Circle(0).Image(), 50, 50)
	// 放置昵称
	canvas.SetRGB(0, 0, 0)
	fontSize := 150.0
	_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return nil, err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
		return nil, err
	}
	nameW, nameH := canvas.MeasureString(a.nickname)
	// 昵称范围
	textH := 300.0
	textW := float64(backDX) * 2 / 3
	// 如果文字超过长度了，比列缩小字体
	if nameW > textW {
		scale := 2 * nameH / textH
		fontSize = fontSize * scale
		if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
			return nil, err
		}
		_, nameH := canvas.MeasureString(a.nickname)
		// 昵称分段
		name := []rune(a.nickname)
		names := make([]string, 0, 4)
		// 如果一半都没到界面边界就分两行
		wordw, _ := canvas.MeasureString(string(name[:len(name)/2]))
		if wordw < textW*3/4 {
			names = append(names, string(name[:len(name)/2+1]))
			names = append(names, string(name[len(name)/2+1:]))
		} else {
			nameLength := 0.0
			lastIndex := 0
			for i, word := range name {
				wordw, _ = canvas.MeasureString(string(word))
				nameLength += wordw
				if nameLength > textW*3/4 || i == len(name)-1 {
					names = append(names, string(name[lastIndex:i+1]))
					lastIndex = i + 1
					nameLength = 0
				}
			}
			// 超过两行就重新配置一下字体大小
			scale = float64(len(names)) * nameH / textH
			fontSize = fontSize * scale
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				return nil, err
			}
		}
		fmt.Println(scale)
		for i, nameSplit := range names {
			canvas.DrawStringAnchored(nameSplit, float64(backDX)/2+25, 25+(200+70*scale)*float64(i+1)/float64(len(names))-nameH/2, 0.5, 0.5)
		}
	} else {
		canvas.DrawStringAnchored(a.nickname, float64(backDX)/2+25, 200-nameH/2, 0.5, 0.5)
	}

	// level
	if err = canvas.LoadFontFace(text.BoldFontFile, 72); err != nil {
		return nil, err
	}
	level := a.level
	levelX := float64(backDX) * 4 / 5
	canvas.DrawRoundedRectangle(levelX, 50, 200, 200, 200/5)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 100)
	canvas.StrokePreserve()
	canvas.SetRGBA255(255, 255, 255, 100)
	canvas.Fill()
	canvas.DrawRoundedRectangle(levelX, 50, 200, 100, 200/5)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 100)
	canvas.StrokePreserve()
	canvas.SetRGBA255(255, 255, 255, 100)
	canvas.Fill()
	canvas.SetRGBA255(0, 0, 0, 255)
	//canvas.DrawStringAnchored(levelrank[level], levelX+100, 50+50, 0.5, 0.5)
	canvas.DrawStringAnchored(fmt.Sprintf("LV%d", level), levelX+100, 50+100+50, 0.5, 0.5)

	if add == 0 {
		canvas.DrawString(fmt.Sprintf("已连签 %d 天    总资产: %d", userinfo.Continuous, a.score), 350, 370)
	} else {
		canvas.DrawString(fmt.Sprintf("连签 %d 天 总资产( +%d ) : %d", userinfo.Continuous, add+level*5, score), 350, 370)
	}
	// 绘制等级进度条
	if err = canvas.LoadFontFace(text.BoldFontFile, 50); err != nil {
		return nil, err
	}
	_, textH = canvas.MeasureString("/")
	switch {
	case userinfo.Level < scoreMax && add == 0:
		canvas.DrawStringAnchored(fmt.Sprintf("%d/%d", userinfo.Level, nextLevelScore), float64(backDX)/2, 455-textH, 0.5, 0.5)
	case userinfo.Level < scoreMax:
		canvas.DrawStringAnchored(fmt.Sprintf("(%d+%d)/%d", userinfo.Level-add, add, nextLevelScore), float64(backDX)/2, 455-textH, 0.5, 0.5)
	default:
		canvas.DrawStringAnchored("Max/Max", float64(backDX)/2, 455-textH, 0.5, 0.5)

	}
	// 创建彩虹条
	grad := gg.NewLinearGradient(0, 500, 1500, 300)
	grad.AddColorStop(0, color.RGBA{G: 255, A: 255})
	grad.AddColorStop(0.25, color.RGBA{B: 255, A: 255})
	grad.AddColorStop(0.5, color.RGBA{R: 255, A: 255})
	grad.AddColorStop(0.75, color.RGBA{B: 255, A: 255})
	grad.AddColorStop(1, color.RGBA{G: 255, A: 255})
	canvas.SetStrokeStyle(grad)
	canvas.SetLineWidth(7)
	// 设置长度
	gradMax := 1300.0
	LevelLength := gradMax * (float64(userinfo.Level) / float64(nextLevelScore))
	canvas.MoveTo((float64(backDX)-LevelLength)/2, 450)
	canvas.LineTo((float64(backDX)+LevelLength)/2, 450)
	canvas.ClosePath()
	canvas.Stroke()
	// 放置图片
	canvas.DrawImageAnchored(back, backDX/2, imgDH/2+475, 0.5, 0.5)
	// 生成图片
	return canvas.Image(), nil
}
