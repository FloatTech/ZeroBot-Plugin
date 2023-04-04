// Package score 签到
package score

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/rendercard"
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

func drawScore18(a *scdata) (img image.Image, err error) {
	var fontdata []byte
	var fdwg sync.WaitGroup
	fdwg.Add(1)
	go func() {
		defer fdwg.Done()
		fontdata, _ = file.GetLazyData(text.GlowSansFontFile, control.Md5File, false)
	}()

	getAvatar, err := initPic(a.picfile, a.uid)
	if err != nil {
		return
	}

	var back, blurback image.Image
	var bx, by, sc float64
	var colors []color.RGBA
	var bwg sync.WaitGroup
	bwg.Add(1)
	back, err = gg.LoadImage(a.picfile)
	if err != nil {
		return
	}
	defer bwg.Done()
	blurback = imaging.Blur(back, 20)
	bx, by = float64(back.Bounds().Dx()), float64(back.Bounds().Dy())
	sc = 1280 / float64(bx)
	colors = gg.TakeColor(back, 3)

	canvas := gg.NewContext(1280, 1280*int(by)/int(bx))
	canvas.ScaleAbout(sc, sc, float64(canvas.W())/2, float64(canvas.H())/2)
	canvas.DrawImageAnchored(blurback, canvas.W()/2, canvas.H()/2, 0.5, 0.5)
	canvas.Identity()

	cw, ch := float64(canvas.W()), float64(canvas.H())

	scback := gg.NewContext(canvas.W(), canvas.H())
	scback.ScaleAbout(sc, sc, float64(canvas.W())/2, float64(canvas.H())/2)
	scback.DrawImageAnchored(back, canvas.W()/2, canvas.H()/2, 0.5, 0.5)
	scback.Identity()

	pureblack := gg.NewContext(canvas.W(), canvas.H())
	pureblack.SetRGBA255(0, 0, 0, 255)
	pureblack.Clear()

	shadow := gg.NewContext(canvas.W(), canvas.H())
	shadow.ScaleAbout(0.6, 0.6, float64(canvas.W())-float64(canvas.W())/3, float64(canvas.H())/2)
	shadow.DrawImageAnchored(pureblack.Image(), canvas.W()-canvas.W()/3, canvas.H()/2, 0.5, 0.5)
	shadow.Identity()

	canvas.DrawImage(imaging.Blur(shadow.Image(), 8), 0, 0)

	canvas.ScaleAbout(0.6, 0.6, float64(canvas.W())-float64(canvas.W())/3, float64(canvas.H())/2)
	canvas.DrawImageAnchored(rendercard.Fillet(scback.Image(), 12), canvas.W()-canvas.W()/3, canvas.H()/2, 0.5, 0.5)
	canvas.Identity()

	ava, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return
	}

	isc := (ch - ch*6/10) / 2 / 2 / 2 * 3 / float64(ava.Bounds().Dy())
	avatar := gg.NewContext(int((ch-ch*6/10)/2/2/2*3), int((ch-ch*6/10)/2/2/2*3))
	avatar.ScaleAbout(isc, isc, float64(avatar.W())/2, float64(avatar.H())/2)
	avatar.DrawImageAnchored(ava, avatar.W()/2, avatar.H()/2, 0.5, 0.5)
	avatar.Identity()
	fdwg.Wait()
	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/2/2)
	if err != nil {
		return
	}
	namew, _ := canvas.MeasureString(a.nickname)

	shadow2 := gg.NewContext(canvas.W(), canvas.H())
	shadow2.DrawRoundedRectangle((ch-ch*6/10)/2/2-float64(avatar.W())/2-float64(avatar.W())/40, (ch-ch*6/10)/2/2-float64(avatar.W())/2-float64(avatar.H())/40, float64(avatar.W())+float64(avatar.W())/40*2, float64(avatar.H())+float64(avatar.H())/40*2, 8)
	shadow2.SetColor(color.Black)
	shadow2.Fill()
	shadow2.DrawRoundedRectangle((ch-ch*6/10)/2/2, (ch-ch*6/10)/2/2-float64(avatar.H())/4, float64(avatar.W())/2+float64(avatar.W())/40*5+namew, float64(avatar.H())/2, 8)
	shadow2.Fill()

	canvas.DrawImage(imaging.Blur(shadow2.Image(), 8), 0, 0)

	canvas.DrawRoundedRectangle((ch-ch*6/10)/2/2-float64(avatar.W())/2-float64(avatar.W())/40, (ch-ch*6/10)/2/2-float64(avatar.W())/2-float64(avatar.H())/40, float64(avatar.W())+float64(avatar.W())/40*2, float64(avatar.H())+float64(avatar.H())/40*2, 8)
	canvas.SetColor(colors[0])
	canvas.Fill()
	canvas.DrawRoundedRectangle((ch-ch*6/10)/2/2, (ch-ch*6/10)/2/2-float64(avatar.H())/4, float64(avatar.W())/2+float64(avatar.W())/40*5+namew, float64(avatar.H())/2, 8)
	canvas.Fill()

	canvas.DrawImageAnchored(rendercard.Fillet(avatar.Image(), 8), int((ch-ch*6/10)/2/2), int((ch-ch*6/10)/2/2), 0.5, 0.5)

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored(a.nickname, (ch-ch*6/10)/2/2+float64(avatar.W())/2+float64(avatar.W())/40*2+2, (ch-ch*6/10)/2/2+2, 0, 0.5)
	canvas.SetColor(color.White)
	canvas.DrawStringAnchored(a.nickname, (ch-ch*6/10)/2/2+float64(avatar.W())/2+float64(avatar.W())/40*2, (ch-ch*6/10)/2/2, 0, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/2/3*2)
	if err != nil {
		return
	}

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored(time.Now().Format("2006/01/02"), cw-cw/6+2, (ch-ch*6/10)/2/4*3+2, 0.5, 0.5)

	canvas.SetColor(color.White)
	canvas.DrawStringAnchored(time.Now().Format("2006/01/02"), cw-cw/6, (ch-ch*6/10)/2/4*3, 0.5, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/2/2)
	if err != nil {
		return
	}
	nextrankScore := 0
	if a.rank < 10 {
		nextrankScore = rankArray[a.rank+1]
	} else {
		nextrankScore = SCOREMAX
	}
	nextLevelStyle := strconv.Itoa(a.level) + "/" + strconv.Itoa(nextrankScore)

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored("Level "+strconv.Itoa(a.rank), cw/3*2-cw*6/10/2+2, ch-(ch-ch*6/10)/2/4*3+2, 0, 0.5)
	canvas.DrawStringAnchored(nextLevelStyle, cw/3*2+cw*6/10/2+2, ch-(ch-ch*6/10)/2/4*3+2, 1, 0.5)
	canvas.SetColor(color.White)
	canvas.DrawStringAnchored("Level "+strconv.Itoa(a.rank), cw/3*2-cw*6/10/2, ch-(ch-ch*6/10)/2/4*3, 0, 0.5)
	canvas.DrawStringAnchored(nextLevelStyle, cw/3*2+cw*6/10/2, ch-(ch-ch*6/10)/2/4*3, 1, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/2/3)
	if err != nil {
		return
	}

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored("Create By ZeroBot-Plugin "+banner.Version, 0+4+2, ch+2, 0, -0.5)
	canvas.SetColor(color.White)
	canvas.DrawStringAnchored("Create By ZeroBot-Plugin "+banner.Version, 0+4, ch, 0, -0.5)

	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/5*3)
	if err != nil {
		return
	}

	tempfh := canvas.FontHeight()

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored(getHourWord(time.Now()), ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8+2, (ch-ch*6/10)/2+ch*6/10/4+2, 0, 0.5)
	canvas.SetColor(color.White)
	canvas.DrawStringAnchored(getHourWord(time.Now()), ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8, (ch-ch*6/10)/2+ch*6/10/4, 0, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/5)
	if err != nil {
		return
	}

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored("ATRI币 + "+strconv.Itoa(a.inc), ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8+2, (ch-ch*6/10)/2+ch*6/10/4+tempfh+2, 0, 0.5)
	canvas.DrawStringAnchored("签到天数 + 1", ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8+2, (ch-ch*6/10)/2+ch*6/10/4+tempfh+canvas.FontHeight()+2, 0, 1)

	canvas.SetColor(color.White)
	canvas.DrawStringAnchored("ATRI币 + "+strconv.Itoa(a.inc), ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8, (ch-ch*6/10)/2+ch*6/10/4+tempfh, 0, 0.5)
	canvas.DrawStringAnchored("签到天数 + 1", ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8, (ch-ch*6/10)/2+ch*6/10/4+tempfh+canvas.FontHeight(), 0, 1)

	err = canvas.ParseFontFace(fontdata, (ch-ch*6/10)/2/4)
	if err != nil {
		return
	}

	canvas.SetColor(color.Black)
	canvas.DrawStringAnchored("你有 "+strconv.Itoa(a.score)+" 枚ATRI币", ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8+2, (ch-ch*6/10)/2+ch*6/10/4*3+2, 0, 0.5)
	canvas.SetColor(color.White)
	canvas.DrawStringAnchored("你有 "+strconv.Itoa(a.score)+" 枚ATRI币", ((cw-cw*6/10)-(cw/3-cw*6/10/2))/8, (ch-ch*6/10)/2+ch*6/10/4*3, 0, 0.5)

	img = canvas.Image()
	return
}
