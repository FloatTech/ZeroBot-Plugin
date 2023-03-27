// Package score 签到
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
	data, err := file.GetLazyData(text.MaokenFontFile, control.Md5File, true)
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
	canvas.DrawStringAnchored("ATRI币 + "+strconv.Itoa(a.inc), 40, float64(imgDY-90), 0, 0)
	canvas.DrawStringAnchored("当前ATRI币："+strconv.Itoa(a.score), 40, float64(imgDY-60), 0, 0)
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
