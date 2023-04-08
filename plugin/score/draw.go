// Package score 签到
package score

import (
	"bytes"
	"errors"
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

type scoredrawer func(a *scdata) (image.Image, error)

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

func drawScore17b2(a *scdata) (img image.Image, err error) {
	fontdata, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, false)
	if err != nil {
		return
	}

	getAvatar, err := initPic(a.picfile, a.uid)
	if err != nil {
		return
	}

	back, err := gg.LoadImage(a.picfile)
	if err != nil {
		return
	}

	bx, by := float64(back.Bounds().Dx()), float64(back.Bounds().Dy())

	sc := 1280 / bx

	colors := gg.TakeColor(back, 3)

	canvas := gg.NewContext(1280, 1280*int(by)/int(bx))

	cw, ch := float64(canvas.W()), float64(canvas.H())

	sch := ch * 6 / 10

	var blurback, scbackimg, backshadowimg, avatarimg, avatarbackimg, avatarshadowimg, whitetext, blacktext image.Image
	var wg sync.WaitGroup
	wg.Add(8)

	go func() {
		defer wg.Done()
		scback := gg.NewContext(canvas.W(), canvas.H())
		scback.ScaleAbout(sc, sc, cw/2, ch/2)
		scback.DrawImageAnchored(back, canvas.W()/2, canvas.H()/2, 0.5, 0.5)
		scback.Identity()

		go func() {
			defer wg.Done()
			blurback = imaging.Blur(scback.Image(), 20)
		}()

		scbackimg = rendercard.Fillet(scback.Image(), 12)
	}()

	go func() {
		defer wg.Done()
		pureblack := gg.NewContext(canvas.W(), canvas.H())
		pureblack.SetRGBA255(0, 0, 0, 255)
		pureblack.Clear()

		shadow := gg.NewContext(canvas.W(), canvas.H())
		shadow.ScaleAbout(0.6, 0.6, cw-cw/3, ch/2)
		shadow.DrawImageAnchored(pureblack.Image(), canvas.W()-canvas.W()/3, canvas.H()/2, 0.5, 0.5)
		shadow.Identity()

		backshadowimg = imaging.Blur(shadow.Image(), 8)
	}()

	aw, ah := (ch-sch)/2/2/2*3, (ch-sch)/2/2/2*3

	go func() {
		defer wg.Done()
		avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
		if err != nil {
			return
		}

		isc := (ch - sch) / 2 / 2 / 2 * 3 / float64(avatar.Bounds().Dy())

		scavatar := gg.NewContext(int(aw), int(ah))

		scavatar.ScaleAbout(isc, isc, aw/2, ah/2)
		scavatar.DrawImageAnchored(avatar, scavatar.W()/2, scavatar.H()/2, 0.5, 0.5)
		scavatar.Identity()

		avatarimg = rendercard.Fillet(scavatar.Image(), 8)
	}()

	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/2/2)
	if err != nil {
		return
	}
	namew, _ := canvas.MeasureString(a.nickname)

	go func() {
		defer wg.Done()
		avatarshadowimg = imaging.Blur(customrectangle(cw, ch, aw, ah, namew, color.Black), 8)
	}()

	go func() {
		defer wg.Done()
		avatarbackimg = customrectangle(cw, ch, aw, ah, namew, colors[0])
	}()

	go func() {
		defer wg.Done()
		whitetext, err = customtext(a, fontdata, cw, ch, aw, color.White)
		if err != nil {
			return
		}
	}()

	go func() {
		defer wg.Done()
		blacktext, err = customtext(a, fontdata, cw, ch, aw, color.Black)
		if err != nil {
			return
		}
	}()

	wg.Wait()
	if scbackimg == nil || backshadowimg == nil || avatarimg == nil || avatarbackimg == nil || avatarshadowimg == nil || whitetext == nil || blacktext == nil {
		err = errors.New("图片渲染失败")
		return
	}

	canvas.DrawImageAnchored(blurback, canvas.W()/2, canvas.H()/2, 0.5, 0.5)

	canvas.DrawImage(backshadowimg, 0, 0)

	canvas.ScaleAbout(0.6, 0.6, cw-cw/3, ch/2)
	canvas.DrawImageAnchored(scbackimg, canvas.W()-canvas.W()/3, canvas.H()/2, 0.5, 0.5)
	canvas.Identity()

	canvas.DrawImage(avatarshadowimg, 0, 0)
	canvas.DrawImage(avatarbackimg, 0, 0)
	canvas.DrawImageAnchored(avatarimg, int((ch-sch)/2/2), int((ch-sch)/2/2), 0.5, 0.5)

	canvas.DrawImage(blacktext, 2, 2)
	canvas.DrawImage(whitetext, 0, 0)

	img = canvas.Image()
	return
}

func customrectangle(cw, ch, aw, ah, namew float64, rtgcolor color.Color) (img image.Image) {
	canvas := gg.NewContext(int(cw), int(ch))
	sch := ch * 6 / 10
	canvas.DrawRoundedRectangle((ch-sch)/2/2-aw/2-aw/40, (ch-sch)/2/2-aw/2-ah/40, aw+aw/40*2, ah+ah/40*2, 8)
	canvas.SetColor(rtgcolor)
	canvas.Fill()
	canvas.DrawRoundedRectangle((ch-sch)/2/2, (ch-sch)/2/2-ah/4, aw/2+aw/40*5+namew, ah/2, 8)
	canvas.Fill()

	img = canvas.Image()
	return
}

func customtext(a *scdata, fontdata []byte, cw, ch, aw float64, textcolor color.Color) (img image.Image, err error) {
	canvas := gg.NewContext(int(cw), int(ch))
	canvas.SetColor(textcolor)
	scw, sch := cw*6/10, ch*6/10
	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/2/2)
	if err != nil {
		return
	}
	canvas.DrawStringAnchored(a.nickname, (ch-sch)/2/2+aw/2+aw/40*2, (ch-sch)/2/2, 0, 0.5)
	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/2/3*2)
	if err != nil {
		return
	}
	canvas.DrawStringAnchored(time.Now().Format("2006/01/02"), cw-cw/6, ch/2-sch/2-canvas.FontHeight(), 0.5, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/2/2)
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

	canvas.DrawStringAnchored("Level "+strconv.Itoa(a.rank), cw/3*2-scw/2, ch/2+sch/2+canvas.FontHeight(), 0, 0.5)
	canvas.DrawStringAnchored(nextLevelStyle, cw/3*2+scw/2, ch/2+sch/2+canvas.FontHeight(), 1, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/2/3)
	if err != nil {
		return
	}

	canvas.DrawStringAnchored("Create By ZeroBot-Plugin "+banner.Version, 0+4, ch, 0, -0.5)

	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/5*3)
	if err != nil {
		return
	}

	tempfh := canvas.FontHeight()

	canvas.DrawStringAnchored(getHourWord(time.Now()), ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4, 0, 0.5)

	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/5)
	if err != nil {
		return
	}

	canvas.DrawStringAnchored("ATRI币 + "+strconv.Itoa(a.inc), ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4+tempfh, 0, 0.5)
	canvas.DrawStringAnchored("EXP + 1", ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4+tempfh+canvas.FontHeight(), 0, 1)

	err = canvas.ParseFontFace(fontdata, (ch-sch)/2/4)
	if err != nil {
		return
	}

	canvas.DrawStringAnchored("你有 "+strconv.Itoa(a.score)+" 枚ATRI币", ((cw-scw)-(cw/3-scw/2))/8, (ch-sch)/2+sch/4*3, 0, 0.5)

	img = canvas.Image()
	return
}
