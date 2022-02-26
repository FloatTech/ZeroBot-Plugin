package gif

import (
	"errors"
	"image"
	"math/rand"
	"os"
	"strconv"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/writer"
)

// A爬
func (cc *context) A爬() (string, error) {
	name := cc.usrdir + `爬.png`
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
	f, err := dlblock(`pa/` + strconv.Itoa(rand) + `.png`)
	if err != nil {
		return "", err
	}
	imgf, err := img.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	return "file:///" + name, writer.SavePNG2Path(name, imgf.InsertBottom(tou, 100, 100, 0, 400).Im)
}

// A撕
func (cc *context) A撕() (string, error) {
	name := cc.usrdir + `撕.png`
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
	f, err := dlblock(`si/0.png`)
	if err != nil {
		return "", err
	}
	imgf, err := img.LoadFirstFrame(f, 0, 0)
	if err != nil {
		return "", err
	}
	return "file:///" + name, writer.SavePNG2Path(name, imgf.InsertBottom(im1.Im, im1.W, im1.H, -3, 370).InsertBottom(im2.Im, im2.W, im2.H, 653, 310).Im)
}

// 简单
func (cc *context) other(value ...string) (string, error) {
	name := cc.usrdir + value[0] + `.png`
	// 加载图片
	im, err := img.LoadFirstFrame(cc.headimgsdir[0], 0, 0)
	if err != nil {
		return "", err
	}
	var imgnrgba *image.NRGBA
	switch value[0] {
	case "上翻", "下翻":
		imgnrgba = im.FlipV().Im
	case "左翻", "右翻":
		imgnrgba = im.FlipH().Im
	case "反色":
		imgnrgba = im.Invert().Im
	case "灰度":
		imgnrgba = im.Grayscale().Im
	case "负片":
		imgnrgba = im.Invert().Grayscale().Im
	case "浮雕":
		imgnrgba = im.Convolve3x3().Im
	case "打码":
		imgnrgba = im.Blur(10).Im
	case "旋转":
		r, _ := strconv.ParseFloat(value[1], 64)
		imgnrgba = img.Rotate(im.Im, r, 0, 0).Im
	case "变形":
		w, err := strconv.Atoi(value[1])
		if err != nil {
			return "", err
		}
		h, err := strconv.Atoi(value[2])
		if err != nil {
			return "", err
		}
		imgnrgba = img.Size(im.Im, w, h).Im
	default:
		return "", errors.New("no such method")
	}
	return "file:///" + name, writer.SavePNG2Path(name, imgnrgba)
}
