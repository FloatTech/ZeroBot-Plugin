package gif

import (
	"errors"
	"image/color"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/fogleman/gg"
)

// pa 爬
func pa(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "爬.png"
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	// 随机爬图序号
	rand := rand.Intn(60) + 1
	if file.IsNotExist(datapath + "materials/pa") {
		err = os.MkdirAll(datapath+"materials/pa", 0755)
		if err != nil {
			return "", err
		}
	}
	f, err := dlblock("pa/" + strconv.Itoa(rand) + ".png")
	if err != nil {
		return "", err
	}
	imgf, err := img.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	return "file:///" + name, writer.SavePNG2Path(name, imgf.InsertBottom(tou, 100, 100, 0, 400).Im)
}

// si 撕
func si(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "撕.png"
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	im1 := img.Rotate(tou, 20, 380, 380)
	im2 := img.Rotate(tou, -12, 380, 380)
	if file.IsNotExist(datapath + "materials/si") {
		err = os.MkdirAll(datapath+"materials/si", 0755)
		if err != nil {
			return "", err
		}
	}
	f, err := dlblock("si/0.png")
	if err != nil {
		return "", err
	}
	imgf, err := img.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	return "file:///" + name, writer.SavePNG2Path(name, imgf.InsertBottom(im1.Im, im1.W, im1.H, -3, 370).InsertBottom(im2.Im, im2.W, im2.H, 653, 310).Im)
}

// flipV 上翻,下翻
func flipV(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "FlipV.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.FlipV().Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// flipH 左翻,右翻
func flipH(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "FlipH.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.FlipH().Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// invert 反色
func invert(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Invert.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Invert().Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// blur 反色
func blur(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Blur.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Blur(10).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// grayscale 灰度
func grayscale(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Grayscale.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Grayscale().Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// invertAndGrayscale 负片
func invertAndGrayscale(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "InvertAndGrayscale.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Invert().Grayscale().Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// convolve3x3 浮雕
func convolve3x3(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Convolve3x3.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Convolve3x3().Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// rotate 旋转,带参数暂时不用
func rotate(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Rotate.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	r, _ := strconv.ParseFloat(args[0], 64)
	imgnrgba := img.Rotate(im.Im, r, 0, 0).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// deformation 变形,带参数暂时不用
func deformation(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Deformation.png"
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	w, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	h, err := strconv.Atoi(args[1])
	if err != nil {
		return "", err
	}
	imgnrgba := img.Size(im.Im, w, h).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// alike 你像个xxx一样
func alike(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("alike", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Anyasuki.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 82, 69)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Im, 0, 0, 136, 21).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// marriage
func marriage(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("marriage", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Marriage.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 1080, 1080)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Im, 0, 0, 0, 0).InsertUp(imgs[1].Im, 0, 0, 800, 0).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// anyasuki 阿尼亚喜欢
func anyasuki(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("anyasuki", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "Anyasuki.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(475, 540)
	canvas.DrawImage(img.Size(face, 347, 267).Im, 82, 53)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 30); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "阿尼亚喜欢这个"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 500 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (500-l)/2.0, 535)
	return "file:///" + name, canvas.SavePNG(name)
}

// alwaysLike 我永远喜欢
func alwaysLike(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("always_like", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "AlwaysLike.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(830, 599)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(img.Size(face, 341, 341).Im, 44, 74)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 56); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "你们"
	}
	args[0] = "我永远喜欢" + args[0]
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/2.0, 559)
	return "file:///" + name, canvas.SavePNG(name)
}

// decentKiss 像样的亲亲
func decentKiss(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("decent_kiss", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "DecentKiss.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 589, 577)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Im, 0, 0, 0, 0).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// chinaFlag 国旗
func chinaFlag(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("china_flag", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "ChinaFlag.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 410, 410)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Im, 0, 0, 0, 0).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// dontTouch 不要靠近
func dontTouch(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("dont_touch", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "DontTouch.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 410, 410)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Im, 148, 148, 46, 238).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// universal 万能表情 空白表情
func universal(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Universal.png"
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(500, 550)
	canvas.DrawImage(img.Size(face, 500, 500).Im, 0, 0)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 40); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "在此处添加文字"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 500 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (500-l)/2.0, 545)
	return "file:///" + name, canvas.SavePNG(name)
}

// interview 采访
func interview(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("interview", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "Interview.png"
	huaji, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	microphone, err := gg.LoadImage(c[1])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(600, 300)
	canvas.DrawImage(img.Size(face, 124, 124).Im, 100, 50)
	canvas.DrawImage(huaji, 376, 50)
	canvas.DrawImage(microphone, 300, 50)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 40); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "采访大佬经验"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 600 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (600-l)/2.0, 270)
	return "file:///" + name, canvas.SavePNG(name)
}

// need 需要 你可能需要
func need(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("need", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Need.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 114, 114)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Im, 0, 0, 327, 232).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// paint 这像画吗
func paint(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("paint", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Paint.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 117, 135)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(img.Rotate(im.Im, 4, 0, 0).Im, 0, 0, 95, 107).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// painter 小画家
func painter(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("painter", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Painter.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 240, 345)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Im, 0, 0, 125, 91).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// perfect 完美
func perfect(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("perfect", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Perfect.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 310, 460)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Im, 0, 0, 313, 64).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// playGame 玩游戏
func playGame(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("play_game", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "PlayGame.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(526, 503)
	canvas.DrawImage(img.Rotate(face, 10, 225, 160).Im, 161, 117)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 40); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "来玩休闲游戏啊"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 526 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (526-l)/2.0, 483)
	return "file:///" + name, canvas.SavePNG(name)
}

// police 出警
func police(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("police", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Police.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 245, 245)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Im, 0, 0, 224, 46).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// police1 警察
func police1(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("police", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Police1.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 60, 75)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[1].InsertBottom(img.Rotate(im.Im, 16, 0, 0).Im, 0, 0, 37, 291).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// prpr 舔 舔屏 prpr
func prpr(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("prpr", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Prpr.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 330, 330)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(img.Rotate(im.Im, 8, 0, 0).Im, 0, 0, 46, 264).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// safeSense 安全感
func safeSense(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("safe_sense", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "SafeSense.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(430, 478)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(img.Size(face, 215, 343).Im, 215, 135)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 30); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "你给我的安全感远不如他的万分之一"
	}

	l, _ := canvas.MeasureString(args[0])
	if l > 860 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0][:len(args[0])/2], (430-l/2)/2.0, 40)
	canvas.DrawString(args[0][len(args[0])/2:], (430-l/2)/2.0, 80)
	return "file:///" + name, canvas.SavePNG(name)
}

// support 精神支柱
func support(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("support", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Support.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 815, 815)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(img.Rotate(im.Im, 23, 0, 0).Im, 0, 0, -172, -17).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// thinkwhat 想什么
func thinkwhat(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("thinkwhat", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Thinkwhat.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Im, 0, 0, 530, 0).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// wallpaper 墙纸
func wallpaper(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("wallpaper", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Wallpaper.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 775, 496)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Im, 0, 0, 260, 580).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// whyatme 为什么at我
func whyatme(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("whyatme", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Whyatme.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 265, 265)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Im, 0, 0, 42, 13).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// makeFriend 交个朋友
func makeFriend(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("make_friend", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "MakeFriend.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(1000, 1000)
	canvas.DrawImage(img.Size(face, 1000, 1000).Im, 0, 0)
	canvas.DrawImage(img.Rotate(face, 9, 250, 250).Im, 743, 845)
	canvas.DrawImage(img.Rotate(face, 9, 55, 55).Im, 836, 722)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.White)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 20); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "我"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 230 {
		return "", errors.New("文字消息太长了")
	}
	canvas.Rotate(gg.Radians(-9))
	canvas.DrawString(args[0], 595, 819)
	return "file:///" + name, canvas.SavePNG(name)
}

// backToWork 打工人, 继续干活
func backToWork(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("back_to_work", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "BackToWork.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 220, 310)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(img.Rotate(im.Im, 25, 0, 0).Im, 0, 0, 56, 32).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// coupon 兑换券
func coupon(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("coupon", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	name := cc.usrdir + "Coupon.png"
	if args[0] == "" {
		args[0] = "群主陪睡券"
	}
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(500, 355)
	canvas.DrawImage(back, 0, 0)
	canvas.Rotate(gg.Radians(-22))
	canvas.DrawImage(img.Size(face, 60, 60).Im, 100, 163)
	canvas.SetColor(color.Black)
	_, err = file.GetLazyData(text.BoldFontFile, true)
	if err != nil {
		return "", err
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, 30); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "陪睡券"
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 270 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawStringAnchored(args[0], 135, 255, 0.5, 0.5)
	canvas.DrawStringAnchored("（永久有效）", 135, 295, 0.5, 0.5)
	return "file:///" + name, canvas.SavePNG(name)
}

// distracted 注意力涣散
func distracted(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("distracted", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Distracted.png"
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 500, 500)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Im, 0, 0, 140, 320).InsertUp(imgs[1].Im, 0, 0, 0, 0).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}

// throw 扔
func throw(cc *context, args ...string) (string, error) {
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("throw", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	if err != nil {
		return "", err
	}
	wg.Wait()
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Throw.png"
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUpC(img.Rotate(face, float64(rand.Intn(360)), 143, 143).Im, 0, 0, 86, 249).Im
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}
