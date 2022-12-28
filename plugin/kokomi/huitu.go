package kokomi

import (
	"image"
	"image/color"

	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/img"
)

// AdjustOpacity 更改透明度
func AdjustOpacity(m image.Image, percentage float64) image.Image {
	bounds := m.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	newRgba := image.NewRGBA64(bounds)
	for i := 0; i < dx; i++ {
		for j := 0; j < dy; j++ {
			colorRgb := m.At(i, j)
			r, g, b, a := colorRgb.RGBA()
			opacity := uint16(float64(a) * percentage)
			v := newRgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: opacity})
			_r, _g, _b, _a := v.RGBA()
			newRgba.SetRGBA64(i, j, color.RGBA64{R: uint16(_r), G: uint16(_g), B: uint16(_b), A: uint16(_a)})
		}
	}
	return newRgba
}

// Yinying 绘制阴影
func Yinying(x int, y int, r float64) image.Image {
	//新建图层,实现阴影400*510
	zero := gg.NewContext(x, y)
	zero.SetRGBA255(0, 0, 0, 213)
	zero.DrawRoundedRectangle(0, 0, float64(x), float64(y), r)
	zero.Fill()
	//模糊
	//shadow := imaging.Blur(one.Image(), 16)
	bg := AdjustOpacity(zero.Image(), 0.6)
	return bg
}

// SetMark 绘制马赛克
func SetMark(pic image.Image) (pictrue []byte) {
	dst := img.Size(pic, 256*5, 256*5)
	b := dst.Im.Bounds()
	markSize := 32

	for y0fMarknum := 0; y0fMarknum <= math.Ceil(b.Max.Y, markSize); y0fMarknum++ {
		for x0fMarknum := 0; x0fMarknum <= math.Ceil(b.Max.X, markSize); x0fMarknum++ {
			a := dst.Im.At(x0fMarknum*markSize+markSize/2, y0fMarknum*markSize+markSize/2)
			cc := color.NRGBAModel.Convert(a).(color.NRGBA)
			for y := 0; y < markSize; y++ {
				for x := 0; x < markSize; x++ {
					x0fPic := x0fMarknum*markSize + x
					y0fPic := y0fMarknum*markSize + y
					dst.Im.Set(x0fPic, y0fPic, cc)
				}
			}
		}
	}
	pictrue, cl := writer.ToBytes(dst.Im)
	defer cl()
	return
}
