// Package text 文字转图片
package text

import (
	"image"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/imgfactory"
)

// 加载数据库
func init() {
	_ = os.MkdirAll(FontPath, 0755)
}

// RenderToBase64 文字转base64
func RenderToBase64(text, font string, width, fontSize int) (base64Bytes []byte, err error) {
	im, err := Render(text, font, width, fontSize)
	if err != nil {
		log.Println("[txt2img]", err)
		return nil, err
	}
	base64Bytes, err = imgfactory.ToBase64(im)
	if err != nil {
		log.Println("[txt2img]", err)
		return nil, err
	}
	return
}

// Render 文字转图片 width 是图片宽度
func Render(text, font string, width, fontSize int) (txtPic image.Image, err error) {
	data, err := file.GetLazyData(font, "data/control/stor.spb", true)
	if err != nil {
		return
	}

	return imgfactory.RenderTextWith(text, data, width, fontSize)
}
