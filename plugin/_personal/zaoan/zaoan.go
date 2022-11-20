// Package draw 画图测试
package draw

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"time"

	// "os"
	"github.com/Coloured-glaze/gg"
	// "github.com/FloatTech/floatbox/file"

	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	fontFile = "data/zaoan/regular.ttf" // 日期字体
	tu       = "https://iw233.cn/api.php?sort=pc"
	yan      = "https://v1.hitokoto.cn/?c=k&encode=text"
)

func init() {
	en := control.Register("zaoan", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "早安/晚安图",
		Help: "- 记录在\"6 30 * * *\"触发的指令\n" +
			"   - /早安||晚安",
	})
	en.OnFullMatchGroup([]string{"/早安", "/晚安"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now()
			hour := now.Hour() // 小时
			var txtPath string
			var yingwen string
			switch {
			case hour <= 11: // 早安
				txtPath = "data/zaoan/zao.jpg"
				yingwen = "Ciallo～(∠・ω< )⌒★"
			case hour <= 18: // 午安
				txtPath = "data/zaoan/wu.jpg"
				yingwen = "Good afternoon"
			default: // 晚安
				txtPath = "data/zaoan/wan.jpg"
				yingwen = "Good night"
			}
			img1, err := gg.LoadImage(txtPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			img2, err := gg.LoadImage("data/zaoan/an.jpg") // 安
			if err != nil {
				fmt.Println(err)
				return
			}
			/****************图片****************/
			pic, err := web.GetData(tu)
			if err != nil {
				ctx.SendChain(message.Text("错误：获取插图失败", err))
				return
			}
			dst, _, err := image.Decode(bytes.NewReader(pic))
			if err != nil {
				ctx.SendChain(message.Text("错误：获取插图失败", err))
				return
			}
			/****************一言****************/
			yi, err := web.GetData(yan)
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			yiyan := helper.BytesToString(yi)
			/***************画图**************/
			xOfPic, yOfPic, err := gg.GetImgWH(pic)
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			var s float64 = 1080
			sx := math.Ceil(s / float64(xOfPic)) // 计算缩放倍率,向上取整
			// 获取文字放大后宽度和高度
			dc := gg.NewContext(1080, yOfPic*int(sx)) // 画布大小
			if err := dc.LoadFontFace(fontFile, 50*sx); err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			wOfDay, hOfDay := dc.MeasureString("焯")
			// 开始画图
			dc = gg.NewContext(1080, 2*int(math.Ceil(hOfDay))+yOfPic*int(sx)+450) // 画布大小
			dc.SetRGB(1, 1, 1)
			dc.Clear()                                                               // 白背景
			dc.SetRGB(0, 0, 0)                                                       // 换黑色
			dc.DrawString(now.Format("2006/01/02"), 100*sx, 100*sx-hOfDay)           // 日期
			dc.DrawString(now.Weekday().String(), (300+10*wOfDay)*sx, 100*sx-hOfDay) // 星期,放在日期300像素后面(10)
			dc.Scale(1, sx)                                                          // 使画笔按倍率缩放
			dc.DrawImage(dst, 0, 100*(1+int(sx)))                                    // 贴图（会受上述缩放倍率影响）
			dc.Scale(1/sx, 1/sx)
			dc.DrawImage(img1, 400*int(sx), 100*(1+int(sx))+yOfPic*int(sx)+90)   // 早
			dc.DrawImage(img2, 400*int(sx), 100*(1+int(sx))+yOfPic*int(sx)+200)  // 安
			dc.DrawString(yingwen, 500*sx, 100*(1+sx)+float64(yOfPic)*sx-hOfDay) // 英文字符串
			dc.DrawStringWrapped(yiyan, 400*sx, 200+float64(yOfPic)*sx+340, 0.5, 0.5, 500, 1.5, gg.AlignLeft)
			ff, cl := writer.ToBytes(dc.Image())
			ctx.SendChain(message.ImageBytes(ff))
			cl()
		})
}
