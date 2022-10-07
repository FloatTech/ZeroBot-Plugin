package phigros

import (
	"hash/crc64"
	"image/color"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/zbputils/img"
)

func renderb19(plname, allrks, chal, chalnum, uid string, list []result) error {
	canvas := gg.NewContext(2360, 4780)
	canvas.SetRGB255(0, 255, 0)
	canvas.Clear()

	drawfile, _ := os.ReadDir(filepath + Illustration)

	imgs, err := img.LoadFirstFrame(filepath+Illustration+drawfile[rand.Intn(len(drawfile))].Name(), 2048, 1080)
	if err != nil {
		return err
	}

	blured := imgs.Blur(30)

	canvas.DrawImage(img.Size(blured.Im, 9064, 4780).Im, -3352, 0)

	draw4(canvas, a, 0, 166, 1324, 410)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	draw4(canvas, a, 1318, 192, 1200, 350)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	draw4(canvas, a, 1320, 164, 6, 414)
	canvas.SetColor(color.White)
	canvas.Fill()

	logo, err := gg.LoadPNG(filepath + Icon)
	if err != nil {
		return err
	}
	canvas.DrawImage(img.Size(logo, 290, 290).Im, 50, 216)

	font, err := gg.LoadFontFace(filepath+Font, 90)
	if err != nil {
		return err
	}
	canvas.SetFontFace(font)
	canvas.DrawString("Phigros", 422, 336)
	canvas.DrawString("RankingScore查询", 422, 462)

	font, err = gg.LoadFontFace(filepath+Font, 54)
	if err != nil {
		return err
	}
	canvas.SetFontFace(font)
	canvas.DrawString(pl+plname, 1434, 300)
	canvas.DrawString(rkss+allrks, 1434, 380)
	canvas.DrawString(cm, 1434, 460)
	if chal != "" {
		chall, err := gg.LoadPNG(filepath + Challengemode + chal + ".png")
		if err != nil {
			return err
		}
		canvas.DrawImage(img.Size(chall, 208, 100).Im, 1848, 392)
		chalnumlen, _ := canvas.MeasureString(chalnum)
		canvas.DrawString(chalnum, 1882+(chalnumlen/2), 460)
	}

	var i int64
	var xj, yj float64 = 1090, 160

	err = mix(canvas, i, list[i])
	if err != nil {
		return err
	}
	x, x1, x2, x3, x4, x5, x6, x7, x8, x9, x10, x11, x12, x13 = x+xj, x1+xj, x2+xj, x3+xj, x4+xj, x5+int(xj), x6+xj, x7+xj, x8+xj, x9+xj, x10+int(xj), x11+xj, x12+xj, x13+xj
	y, y1, y2, y3, y4, y5, y6, y7, y8, y9, y10, y11, y12, y13 = y+yj, y1+yj, y2+yj, y3+yj, y4+yj, y5+int(yj), y6+yj, y7+yj, y8+yj, y9+yj, y10+int(yj), y11+yj, y12+yj, y13+yj
	i++
	for ; i < 22; i++ {
		if i%2 == 0 {
			err := mix(canvas, i, list[i])
			if err != nil {
				return err
			}

			x, x1, x2, x3, x4, x5, x6, x7, x8, x9, x10, x11, x12, x13 = x+xj, x1+xj, x2+xj, x3+xj, x4+xj, x5+int(xj), x6+xj, x7+xj, x8+xj, x9+xj, x10+int(xj), x11+xj, x12+xj, x13+xj
			y, y1, y2, y3, y4, y5, y6, y7, y8, y9, y10, y11, y12, y13 = y+yj, y1+yj, y2+yj, y3+yj, y4+yj, y5+int(yj), y6+yj, y7+yj, y8+yj, y9+yj, y10+int(yj), y11+yj, y12+yj, y13+yj
		} else {
			err := mix(canvas, i, list[i])
			if err != nil {
				return err
			}

			x, x1, x2, x3, x4, x5, x6, x7, x8, x9, x10, x11, x12, x13 = x-xj, x1-xj, x2-xj, x3-xj, x4-xj, x5-int(xj), x6-xj, x7-xj, x8-xj, x9-xj, x10-int(xj), x11-xj, x12-xj, x13-xj
			y, y1, y2, y3, y4, y5, y6, y7, y8, y9, y10, y11, y12, y13 = y+yj, y1+yj, y2+yj, y3+yj, y4+yj, y5+int(yj), y6+yj, y7+yj, y8+yj, y9+yj, y10+int(yj), y11+yj, y12+yj, y13+yj
		}
	}

	// x, x1, x2, x3, x4, x5, x6, x7, x8, x9, x10, x11, x12, x13 = x-xj, x1-xj, x2-xj, x3-xj, x4-xj, x5-int(xj), x6-xj, x7-xj, x8-xj, x9-xj, x10-int(xj), x11-xj, x12-xj, x13-xj
	y, y1, y2, y3, y4, y5, y6, y7, y8, y9, y10, y11, y12, y13 = y-(yj*20), y1-(yj*20), y2-(yj*20), y3-(yj*20), y4-(yj*20), y5-int(yj*20), y6-(yj*20), y7-(yj*20), y8-(yj*20), y9-(yj*20), y10-int(yj*20), y11-(yj*20), y12-(yj*20), y13-(yj*20)
	_ = os.Mkdir(filepath+uid, 0644)
	return canvas.SavePNG(filepath + uid + "/output.png")
}

// 绘制平行四边形 angle 角度 x, y 坐标 w 宽度 l 斜边长
func draw4(canvas *gg.Context, angle, x, y, w, l float64) {
	// 左上角为原点
	x0, y0 := x, y
	// 右上角
	x1, y1 := x+w, y
	// 右下角
	x2 := x1 - (l * (math.Cos(angle * math.Pi / 180.0)))
	y2 := y1 + (l * (math.Sin(angle * math.Pi / 180.0)))
	// 左下角
	x3, y3 := x2-w, y2
	canvas.NewSubPath()
	canvas.MoveTo(x0, y0)
	canvas.LineTo(x1, y1)
	canvas.LineTo(x2, y2)
	canvas.LineTo(x3, y3)
	canvas.ClosePath()
}

func mix(canvas *gg.Context, i int64, list result) (err error) {
	// 画排名背景
	draw4(canvas, a, x, y, w, h)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	// 画排名
	font, err := gg.LoadFontFace(filepath+Font, 30)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(0, 0, 0, 255)
	var fw2 float64
	if i == 0 {
		fw2, _ = canvas.MeasureString("Phi")
		canvas.DrawString("Phi", x6+((w-fw2)/2), y6)
	} else {
		fw2, _ = canvas.MeasureString("#" + strconv.FormatInt(i, 10))
		canvas.DrawString("#"+strconv.FormatInt(i, 10), x6+((w-fw2)/2), y6)
	}

	// 画分数背景
	draw4(canvas, a, x3, y3, w3, h3)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	var rankim *img.Factory
	// 画rank图标
	if list.Rank != "" {
		rankim, err = img.LoadFirstFrame(filepath+Rank+list.Rank+".png", 110, 110)
		if err != nil {
			return
		}
		canvas.DrawImage(rankim.Im, x10, y10)
	}

	// 画分数线
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.DrawRectangle(x7, y7, w7, h7)
	canvas.Fill()

	// 画分数
	font, err = gg.LoadFontFace(filepath+Font, 50)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(255, 255, 255, 255)
	var fw5 float64
	scorestr := strconv.FormatInt(list.Score, 10)
	if len(scorestr) < 7 {
		for i := len(scorestr); i < 7; i++ {
			scorestr = "0" + scorestr
		}
	}
	if list.Score != 0 {
		fw5, _ = canvas.MeasureString(scorestr)
		canvas.DrawString(scorestr, x11+((w7-fw5)/2), y11)
	} else {
		fw5, _ = canvas.MeasureString("0000000")
		canvas.DrawString("0000000", x11+((w7-fw5)/2), y11)
	}

	// 画acc
	font, err = gg.LoadFontFace(filepath+Font, 44)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(255, 255, 255, 255)
	var fw float64
	if list.Acc != 0 {
		fw, _ = canvas.MeasureString(strconv.FormatFloat(list.Acc, 'f', 2, 64) + "%")
		canvas.DrawString(strconv.FormatFloat(list.Acc, 'f', 2, 64)+"%", x13+((w3-fw)/2), y13)
	} else {
		fw, _ = canvas.MeasureString("00.00%")
		canvas.DrawString("00.00%", x13+((w3-fw)/2), y13)
	}

	// 画曲名
	font, err = gg.LoadFontFace(filepath+Font, 32)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(255, 255, 255, 255)
	var fw1 float64
	if list.Songname != "" {
		fw1, _ = canvas.MeasureString(list.Songname)
		canvas.DrawString(list.Songname, x12+((w3-fw1)/2), y12)
	} else {
		fw1, _ = canvas.MeasureString(" ")
		canvas.DrawString(" ", x12+((w3-fw1)/2), y12)
	}

	// 画图片
	draw4(canvas, a, x1, y1, w1, h1)
	canvas.SetRGBA255(0, 0, 255, 0)
	canvas.Fill()
	var imgs *img.Factory
	if list.Songname != "" {
		imgs, err = img.LoadFirstFrame(filepath+Illustration+list.Songname+".png", 2048, 1080)
		if err != nil {
			return
		}
		cutted := cut4img(imgs, a)
		canvas.DrawImage(img.Size(cutted.Im, 436, 230).Im, x5, y5)
	}

	// 画定数背景
	draw4(canvas, a, x2, y2, w2, h2)
	switch list.Diff {
	case "AT":
		canvas.SetRGBA255(56, 56, 56, 255)
	case "IN":
		canvas.SetRGBA255(190, 45, 35, 255)
	case "HD":
		canvas.SetRGBA255(3, 115, 190, 255)
	case "EZ":
		canvas.SetRGBA255(15, 180, 145, 255)
	default:
		canvas.SetRGBA255(56, 56, 56, 255)
	}
	canvas.Fill()

	// 画定数
	font, err = gg.LoadFontFace(filepath+Font, 30)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(255, 255, 255, 255)
	if list.Diff != "" {
		fw3, _ := canvas.MeasureString(list.Diff + " " + strconv.FormatFloat(list.Diffnum, 'f', 1, 64))
		canvas.DrawString(list.Diff+" "+strconv.FormatFloat(list.Diffnum, 'f', 1, 64), x8+((w2-fw3)/2), y8)
	} else {
		fw3, _ := canvas.MeasureString("SP ?")
		canvas.DrawString("SP ?", x8+((w2-fw3)/2), y8)
	}

	font, err = gg.LoadFontFace(filepath+Font, 44)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(255, 255, 255, 255)
	if list.Rksm != 0 {
		fw4, _ := canvas.MeasureString(strconv.FormatFloat(list.Rksm, 'f', 2, 64))
		canvas.DrawString(strconv.FormatFloat(list.Rksm, 'f', 2, 64), x9+((w2-fw4)/2), y9)
	} else {
		fw4, _ := canvas.MeasureString("0.00")
		canvas.DrawString("0.00", x9+((w2-fw4)/2), y9)
	}

	// 画边缘
	draw4(canvas, a, x4, y4, w4, h4)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()
	return nil
}

// 将矩形图片裁切为平行四边形 angle 为角度
func cut4img(imgs *img.Factory, angle float64) *img.Factory {
	db := imgs.Im.Bounds()
	dst := imgs
	maxy := db.Max.Y
	maxx := db.Max.X
	sax := (float64(maxy) * (math.Cos(angle * math.Pi / 180.0)))
	ax := sax
	for autoadd := 1; autoadd < maxy; autoadd++ {
		for ; ax > 0; ax-- {
			dst.Im.Set(int(ax), autoadd, color.NRGBA{0, 0, 0, 0})
			dst.Im.Set(maxx+int(-ax), maxy-autoadd, color.NRGBA{0, 0, 0, 0})
		}
		ax = (float64(maxy-autoadd) * (math.Cos(angle * math.Pi / 180.0)))
	}
	return dst
}

func rksc(accc, diff float64) float64 {
	return ((100.0*(accc/100.0) - 55.0) / 45.0) * ((100.0*(accc/100.0) - 55.0) / 45.0) * diff
}

func arks(sco float64) float64 {
	return sco / 20
}

func idof(songname, diff string) int64 {
	return int64(crc64.Checksum(binary.StringToBytes(songname+diff), crc64.MakeTable(crc64.ISO)))
}

func checkrank(score int64) string {
	if score == 1000000 {
		return "phi"
	}
	if score >= 960000 {
		return "v"
	}
	if score >= 920000 {
		return "s"
	}
	if score >= 880000 {
		return "a"
	}
	if score >= 820000 {
		return "b"
	}
	if score >= 700000 {
		return "c"
	}
	return "f"
}
