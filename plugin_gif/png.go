package gif

import (
	"image"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/img"
)

// 爬
func (cc *context) pa() string {
	name := cc.usrdir + `爬.png`
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0).Circle(0).Im
	// 随机爬图序号
	rand := rand.Intn(60) + 1
	dc := img.LoadFirstFrame(dlblock(`pa/`+strconv.Itoa(rand)+`.png`), 0, 0).
		InsertBottom(tou, 100, 100, 0, 400).Im
	_ = img.SavePng(dc, name)
	return "file:///" + name
}

// 撕
func (cc *context) si() string {
	name := cc.usrdir + `撕.png`
	tou := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0).Im
	im1 := img.Rotate(tou, 20, 380, 380)
	im2 := img.Rotate(tou, -12, 380, 380)
	dc := img.LoadFirstFrame(dlblock(`si/0.png`), 0, 0).
		InsertBottom(im1.Im, im1.W, im1.H, -3, 370).
		InsertBottom(im2.Im, im2.W, im2.H, 653, 310).Im
	_ = img.SavePng(dc, name)
	return "file:///" + name
}

// 简单
func (cc *context) other(value ...string) string {
	name := cc.usrdir + value[0] + `.png`
	// 加载图片
	im := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	var a *image.NRGBA

	switch value[0] {
	case "上翻", "下翻":
		a = im.FlipV().Im
	case "左翻", "右翻":
		a = im.FlipH().Im
	case "反色":
		a = im.Invert().Im
	case "灰度":
		a = im.Grayscale().Im
	case "负片":
		a = im.Invert().Grayscale().Im
	case "浮雕":
		a = im.Convolve3x3().Im
	case "打码":
		a = im.Blur(10).Im
	case "旋转":
		r, _ := strconv.ParseFloat(value[1], 64)
		a = img.Rotate(im.Im, r, 0, 0).Im
	case "变形":
		w, _ := strconv.Atoi(value[1])
		h, _ := strconv.Atoi(value[2])
		a = img.Size(im.Im, w, h).Im
	}

	_ = img.SavePng(a, name)
	return "file:///" + name
}
