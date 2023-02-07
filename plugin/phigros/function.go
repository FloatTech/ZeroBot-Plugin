package phigros

import (
	"hash/crc64"
	"image/color"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
)

func renderb19(plname, allrks, chal, chalnum, uid string, list []result) error {
	canvas := gg.NewContext(2360, 4780)
	canvas.SetRGB255(0, 255, 0)
	canvas.Clear()

	drawfile, _ := os.ReadDir(filepath + Illustration)

	imgs, err := imgfactory.LoadFirstFrame(filepath+Illustration+drawfile[rand.Intn(len(drawfile))].Name(), 2048, 1080)
	if err != nil {
		return err
	}

	blured := imgs.Blur(30)

	var a float64 = 75

	canvas.DrawImage(imgfactory.Size(blured.Image(), 9064, 4780).Image(), -3352, 0)

	draw4(canvas, a, 0, 166, 1324, 410)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	draw4(canvas, a, 1318, 192, 1200, 350)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	draw4(canvas, a, 1320, 164, 6, 414)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	draw4(canvas, a, 534, 4342, 1312, 342)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	draw4(canvas, a, 530, 4344, 6, 346)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	draw4(canvas, a, 1842, 4344, 6, 346)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()

	font, err := gg.LoadFontFace(filepath+Font, 60)
	if err != nil {
		return err
	}
	canvas.SetFontFace(font)
	fw6, fh6 := canvas.MeasureString("Create By ZeroBot-Plugin")
	fw7, _ := canvas.MeasureString("UI Designer: eastown")
	fw8, _ := canvas.MeasureString("*Phigros B19 Picture*")
	canvas.DrawString("Create By ZeroBot-Plugin", 494+(1312-fw6)/2, 4342+((332-fh6*3)/4)+fh6-15)
	canvas.DrawString("UI Designer: eastown", 494+(1312-fw7)/2, 4342+((332-fh6*3)/2)+(fh6*2)-15)
	canvas.DrawString("*Phigros B19 Picture*", 494+(1312-fw8)/2, 4342+((((332-fh6*3)/4)*3)+(fh6*3))-15)

	logo, err := gg.LoadPNG(filepath + Icon)
	if err != nil {
		return err
	}
	canvas.DrawImage(imgfactory.Size(logo, 290, 290).Image(), 50, 216)

	font, err = gg.LoadFontFace(filepath+Font, 90)
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
	_, fh9 := canvas.MeasureString("Player: ")
	canvas.DrawString("Player: "+plname, 1434, 192+(338-(fh9*3))/4+fh9-14)
	canvas.DrawString("RankingScore: "+allrks, 1434, 192+((338-(fh9*3))/2)+(fh9*2)-14)
	canvas.DrawString("ChallengeMode: ", 1434, 192+(((338-(fh9*3))/4)*3)+(fh9*3)-14)
	if chal != "" {
		chall, err := gg.LoadPNG(filepath + Challengemode + chal + ".png")
		if err != nil {
			return err
		}
		canvas.DrawImage(imgfactory.Size(chall, 208, 100).Image(), 1848, 404)
		chalnumlen, _ := canvas.MeasureString(chalnum)
		canvas.DrawString(chalnum, 1882+(chalnumlen/2), 192+(((338-(fh9*3))/4)*3)+(fh9*3)-14)
	}

	var x, y float64 = 188, 682
	var i int64
	var xj, yj float64 = 1090, 160

	err = mix(canvas, i, a, x, y, list[i])
	if err != nil {
		return err
	}
	i++
	x += xj
	y += yj
	for ; i < 22; i++ {
		if i%2 == 0 {
			err := mix(canvas, i, a, x, y, list[i])
			if err != nil {
				return err
			}

			x += xj
			y += yj
		} else {
			err := mix(canvas, i, a, x, y, list[i])
			if err != nil {
				return err
			}

			x -= xj
			y += yj
		}
	}
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

func mix(canvas *gg.Context, i int64, a, x, y float64, list result) (err error) {
	// 画排名背景
	draw4(canvas, a, x, y, 70, 44)
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
		canvas.DrawString("Phi", (x-10)+((70-fw2)/2), y+32)
	} else {
		fw2, _ = canvas.MeasureString("#" + strconv.FormatInt(i, 10))
		canvas.DrawString("#"+strconv.FormatInt(i, 10), (x-10)+((70-fw2)/2), y+32)
	}

	// 画分数背景
	draw4(canvas, a, x+408, y+12, 518, 218)
	canvas.SetRGBA255(0, 0, 0, 160)
	canvas.Fill()

	var rankim *imgfactory.Factory
	// 画rank图标
	if list.Rank != "" {
		rankim, err = imgfactory.LoadFirstFrame(filepath+Rank+list.Rank+".png", 110, 110)
		if err != nil {
			return
		}
		canvas.DrawImage(rankim.Image(), int(x)+412, int(y)+88)
	}

	// 画分数线
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.DrawRectangle(x+536, y+142, 326, 2)
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
		canvas.DrawString(scorestr, (x+532)+((326-fw5)/2), y+116)
	} else {
		fw5, _ = canvas.MeasureString("0000000")
		canvas.DrawString("0000000", (x+532)+((326-fw5)/2), y+116)
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
		canvas.DrawString(strconv.FormatFloat(list.Acc, 'f', 2, 64)+"%", (x+536)+((336-fw)/2), y+196)
	} else {
		fw, _ = canvas.MeasureString("00.00%")
		canvas.DrawString("00.00%", (x+536)+((336-fw)/2), y+196)
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
		canvas.DrawString(list.Songname, (x+408)+((518-fw1)/2), y+58)
	} else {
		fw1, _ = canvas.MeasureString(" ")
		canvas.DrawString(" ", (x+408)+((518-fw1)/2), y+58)
	}

	// 画图片
	draw4(canvas, a, x+68, y, 348, 238)
	canvas.SetRGBA255(0, 0, 255, 0)
	canvas.Fill()
	var imgs *imgfactory.Factory
	if list.Songname != "" {
		imgs, err = imgfactory.LoadFirstFrame(filepath+Illustration+list.Songname+".png", 2048, 1080)
		if err != nil {
			return
		}
		cutted := cut4img(imgs, a)
		canvas.DrawImage(imgfactory.Size(cutted.Image(), 436, 230).Image(), int(x)+6, int(y))
	}

	// 画定数背景
	draw4(canvas, a, x-36, y+139, 138, 94)
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
		canvas.DrawString(list.Diff+" "+strconv.FormatFloat(list.Diffnum, 'f', 1, 64), (x-44)+((138-fw3)/2), y+174)
	} else {
		fw3, _ := canvas.MeasureString("SP ?")
		canvas.DrawString("SP ?", (x-44)+((138-fw3)/2), y+174)
	}

	font, err = gg.LoadFontFace(filepath+Font, 44)
	if err != nil {
		return
	}
	canvas.SetFontFace(font)
	canvas.SetRGBA255(255, 255, 255, 255)
	if list.Rksm != 0 {
		fw4, _ := canvas.MeasureString(strconv.FormatFloat(list.Rksm, 'f', 2, 64))
		canvas.DrawString(strconv.FormatFloat(list.Rksm, 'f', 2, 64), (x-50)+((138-fw4)/2), y+216)
	} else {
		fw4, _ := canvas.MeasureString("0.00")
		canvas.DrawString("0.00", (x-50)+((138-fw4)/2), y+216)
	}

	// 画边缘
	draw4(canvas, a, x+926, y+10, 6, 222)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Fill()
	return nil
}

// 将矩形图片裁切为平行四边形 angle 为角度
func cut4img(imgs *imgfactory.Factory, angle float64) *imgfactory.Factory {
	db := imgs.Image().Bounds()
	dst := imgs
	maxy := db.Max.Y
	maxx := db.Max.X
	sax := (float64(maxy) * (math.Cos(angle * math.Pi / 180.0)))
	ax := sax
	for autoadd := 1; autoadd < maxy; autoadd++ {
		for ; ax > 0; ax-- {
			dst.Image().Set(int(ax), autoadd, color.NRGBA{0, 0, 0, 0})
			dst.Image().Set(maxx+int(-ax), maxy-autoadd, color.NRGBA{0, 0, 0, 0})
		}
		ax = (float64(maxy-autoadd) * (math.Cos(angle * math.Pi / 180.0)))
	}
	return dst
}

func rksc(accc, diff float64) float64 {
	return ((100.0*(accc/100.0) - 55.0) / 45.0) * ((100.0*(accc/100.0) - 55.0) / 45.0) * diff
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
