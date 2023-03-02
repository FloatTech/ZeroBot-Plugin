package gif

import (
	"errors"
	"image/color"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

// pa 爬
func pa(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "爬.png"
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	// 随机爬图序号
	rand := rand.Intn(92) + 1
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
	imgf, err := imgfactory.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgf.InsertUp(tou, 100, 100, 0, 400).Image())
}

// si 撕
func si(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "撕.png"
	tou, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	im1 := imgfactory.Rotate(tou, 20, 380, 380)
	im2 := imgfactory.Rotate(tou, -12, 380, 380)
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
	imgf, err := imgfactory.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgf.InsertBottom(im1.Image(), im1.W(), im1.H(), -3, 370).InsertBottom(im2.Image(), im2.W(), im2.H(), 653, 310).Image())
}

// flipV 上翻,下翻
func flipV(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "FlipV.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.FlipV().Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// flipH 左翻,右翻
func flipH(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "FlipH.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.FlipH().Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// invert 反色
func invert(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "Invert.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Invert().Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// blur 反色
func blur(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "Blur.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Blur(10).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// grayscale 灰度
func grayscale(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "Grayscale.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Grayscale().Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// invertAndGrayscale 负片
func invertAndGrayscale(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "InvertAndGrayscale.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Invert().Grayscale().Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// convolve3x3 浮雕
func convolve3x3(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "Convolve3x3.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := im.Relief().Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// rotate 旋转
func rotate(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Rotate.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	r, _ := strconv.ParseFloat(args[0], 64)
	imgnrgba := imgfactory.Rotate(im.Image(), r, 0, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// deformation 变形
func deformation(cc *context, args ...string) (string, error) {
	name := cc.usrdir + "Deformation.png"
	// 加载图片
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
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
	imgnrgba := imgfactory.Size(im.Image(), w, h).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// alike 你像个xxx一样
func alike(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("alike", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Anyasuki.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 82, 69)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Image(), 0, 0, 136, 21).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// marriage
func marriage(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("marriage", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Marriage.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 1080, 1080)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 0, 0).InsertUp(imgs[1].Image(), 0, 0, 800, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// anyasuki 阿尼亚喜欢
func anyasuki(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("anyasuki", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Size(face, 347, 267).Image(), 82, 53)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("always_like", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Size(face, 380, 380).Image(), 44, 74)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 56); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("decent_kiss", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "DecentKiss.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 589, 577)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 0, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// chinaFlag 国旗
func chinaFlag(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("china_flag", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "ChinaFlag.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 410, 410)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 0, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// dontTouch 不要靠近
func dontTouch(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("dont_touch", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "DontTouch.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 410, 410)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Image(), 148, 148, 46, 238).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// universal 万能表情 空白表情
func universal(cc *context, args ...string) (string, error) {
	_ = args
	name := cc.usrdir + "Universal.png"
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(500, 550)
	canvas.DrawImage(imgfactory.Size(face, 500, 500).Image(), 0, 0)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 40); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("interview", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Size(face, 124, 124).Image(), 100, 50)
	canvas.DrawImage(huaji, 376, 50)
	canvas.DrawImage(microphone, 300, 50)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 40); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("need", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Need.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 114, 114)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 327, 232).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// paint 这像画吗
func paint(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("paint", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Paint.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 117, 135)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 4, 0, 0).Image(), 0, 0, 95, 107).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// painter 小画家
func painter(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("painter", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Painter.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 240, 345)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 125, 91).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// perfect 完美
func perfect(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("perfect", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Perfect.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 310, 460)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUp(im.Image(), 0, 0, 313, 64).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// playGame 玩游戏
func playGame(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("play_game", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Rotate(face, 10, 225, 160).Image(), 161, 117)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 40); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("police", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Police.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 245, 245)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 224, 46).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// police1 警察
func police1(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("police", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Police1.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 60, 75)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[1].InsertBottom(imgfactory.Rotate(im.Image(), 16, 0, 0).Image(), 0, 0, 37, 291).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// prpr 舔 舔屏 prpr
func prpr(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("prpr", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Prpr.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 330, 330)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 8, 0, 0).Image(), 0, 0, 46, 264).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// safeSense 安全感
func safeSense(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("safe_sense", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Size(face, 215, 343).Image(), 215, 135)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("support", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Support.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 815, 815)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 23, 0, 0).Image(), 0, 0, -172, -17).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// thinkwhat 想什么
func thinkwhat(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("thinkwhat", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Thinkwhat.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 530, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// wallpaper 墙纸
func wallpaper(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("wallpaper", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Wallpaper.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 775, 496)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 260, 580).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// whyatme 为什么at我
func whyatme(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("whyatme", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Whyatme.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 265, 265)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 0, 0, 42, 13).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// makeFriend 交个朋友
func makeFriend(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("make_friend", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Size(face, 1000, 1000).Image(), 0, 0)
	canvas.DrawImage(imgfactory.Rotate(face, 9, 250, 250).Image(), 743, 845)
	canvas.DrawImage(imgfactory.Rotate(face, 9, 55, 55).Image(), 836, 722)
	canvas.DrawImage(back, 0, 0)
	canvas.SetColor(color.White)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 20); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("back_to_work", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "BackToWork.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 220, 310)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 25, 0, 0).Image(), 0, 0, 56, 32).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// coupon 兑换券
func coupon(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("coupon", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
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
	canvas.DrawImage(imgfactory.Size(face, 60, 60).Image(), 100, 163)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 30); err != nil {
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
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("distracted", 2, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 2)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Distracted.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 500, 500)
	if err != nil {
		return "", err
	}
	imgnrgba := im.InsertUp(imgs[0].Image(), 0, 0, 140, 320).InsertUp(imgs[1].Image(), 0, 0, 0, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// throw 扔
func throw(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("throw", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "Throw.png"
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertUpC(imgfactory.Rotate(face, float64(rand.Intn(360)), 143, 143).Image(), 0, 0, 86, 249).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// 远离
func yuanli(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("yuanli", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "yuanli.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 420, 420, 45, 90).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// 不是你老婆
func nowife(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("nowife", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "nowife.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 534, 493)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 400, 400, 112, 81).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// youer 你老婆
func youer(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("youer", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "youer.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	tou, err := cc.getLogo(120, 120)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(690, 690)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(tou, 350, 350).Image(), 55, 165)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 56); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "老婆真棒"
	}
	args[0] = "你的" + args[0]
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/3.0, 630)
	return "file:///" + name, canvas.SavePNG(name)
}

// xiaotiamshi 小天使
func xiaotianshi(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("xiaotianshi", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "xiaotianshi.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(522, 665)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 480, 480).Image(), 20, 80)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 35); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "我老婆"
	}
	args[0] = "请问你们看到" + args[0] + "了吗？"
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/10, 50)
	return "file:///" + name, canvas.SavePNG(name)
}

// 不要再看这些了
func neko(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("neko", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "neko.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 712, 949)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 0, 0, 0).Image(), 450, 450, 0, 170).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// 给我变
func bian(cc *context, args ...string) (string, error) {
	_ = args
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("bian", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "bian.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 640, 550)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(imgfactory.Rotate(im.Image(), 0, 0, 0).Image(), 380, 380, 225, -20).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// van 玩一下
func van(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("van", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "van.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := gg.LoadImage(cc.headimgsdir[0])
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(522, 665)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 480, 480).Image(), 20, 80)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 35); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = "RBQ"
	}
	args[0] = "请问你们看到" + args[0] + "了吗？"
	l, _ := canvas.MeasureString(args[0])
	if l > 830 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (830-l)/10, 50)
	return "file:///" + name, canvas.SavePNG(name)
}

// eihei 诶嘿
func eihei(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("eihei", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "eihei.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 690, 690)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 450, 450, 121, 162).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// fanfa 犯法
func fanfa(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("fanfa", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "fanfa.png"
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	m1 := imgfactory.Rotate(face, 45, 110, 110)
	imgnrgba := imgs[0].InsertUp(m1.Image(), 0, 0, 125, 360).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// huai 怀
func huai(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("huai", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "huai.png"
	im, err := imgfactory.LoadFirstFrame(cc.headimgsdir[0], 640, 640)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(im.Image(), 640, 640, 0, 0).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// haowan 好玩
func haowan(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("haowan", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	imgs, err := loadFirstFrames(c, 1)
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "haowan.png"
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	imgnrgba := imgs[0].InsertBottom(face, 90, 90, 321, 172).Image()
	return "file:///" + name, imgfactory.SavePNG2Path(name, imgnrgba)
}

// mengbi 蒙蔽
func mengbi(cc *context, args ...string) (string, error) {
	_ = args
	var wg sync.WaitGroup
	var m sync.Mutex
	var err error
	c := dlrange("mengbi", 1, &wg, func(e error) {
		m.Lock()
		err = e
		m.Unlock()
	})
	wg.Wait()
	if err != nil {
		return "", err
	}
	name := cc.usrdir + "mengbi.png"
	back, err := gg.LoadImage(c[0])
	if err != nil {
		return "", err
	}
	face, err := cc.getLogo(0, 0)
	if err != nil {
		return "", err
	}
	canvas := gg.NewContext(1080, 1080)
	canvas.DrawImage(back, 0, 0)
	canvas.DrawImage(imgfactory.Size(face, 100, 100).Image(), 392, 460)
	canvas.DrawImage(imgfactory.Size(face, 100, 100).Image(), 606, 443)
	canvas.SetColor(color.Black)
	data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return "", err
	}
	if err = canvas.ParseFontFace(data, 80); err != nil {
		return "", err
	}
	if args[0] == "" {
		args[0] = ""
	}
	l, _ := canvas.MeasureString(args[0])
	if l > 1080 {
		return "", errors.New("文字消息太长了")
	}
	canvas.DrawString(args[0], (1080-l)/2, 1000)
	return "file:///" + name, canvas.SavePNG(name)
}
