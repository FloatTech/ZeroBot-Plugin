package txt2img

import (
	"bytes"
	"encoding/base64"
	"github.com/fogleman/gg"
	"github.com/mattn/go-runewidth"
	log "github.com/sirupsen/logrus"
	"image/jpeg"
	"strings"
)

const (
	whitespace = "\t\n\r\x0b\x0c"
	fontPath   = "data/font/simhei.ttf"
)

func Txt2img(text string, width, fontSize int) (base64Bytes []byte, err error) {
	buff := make([]string, 0)
	line := ""
	count := 0
	for _, v := range text {
		c := string(v)
		if strings.Contains(whitespace, c) {
			buff = append(buff, strings.TrimSpace(line))
			count = 0
			line = ""
			continue
		}
		if count <= width {
			line += c
			count = count + runewidth.StringWidth(c)
		} else {
			buff = append(buff, line)
			line = c
			count = runewidth.StringWidth(c)
		}
	}

	canvas := gg.NewContext((fontSize+3)*width/2, (len(buff)+2)*fontSize)
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()
	canvas.SetRGB(0, 0, 0)
	if err = canvas.LoadFontFace(fontPath, float64(fontSize)); err != nil {
		log.Println("err:", err)
	}
	for i, v := range buff {
		if v != "" {
			canvas.DrawString(v, float64(width/2), float64((i+2)*fontSize))
		}
	}
	// 转成 base64
	buffer := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buffer)
	var opt jpeg.Options
	opt.Quality = 70
	err = jpeg.Encode(encoder, canvas.Image(), &opt)
	if err != nil {
		return nil, err
	}
	encoder.Close()
	base64Bytes = buffer.Bytes()
	return
}
