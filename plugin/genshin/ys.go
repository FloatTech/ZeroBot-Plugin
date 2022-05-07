// Package genshin 原神抽卡
package genshin

import (
	"archive/zip"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/FloatTech/zbputils/process"
	"github.com/golang/freetype"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type zipfilestructure map[string][]*zip.File

var (
	totl                   uint64 // 累计抽奖次数
	filetree               = make(zipfilestructure, 32)
	starN3, starN4, starN5 *zip.File
	namereg                = regexp.MustCompile(`_(.*)\.png`)
)

func init() {
	engine := control.Register("genshin", &control.Options{
		DisableOnDefault: false,
		Help:             "原神抽卡\n- 原神十连\n- 切换原神卡池",
		PublicDataFolder: "Genshin",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnFullMatch("切换原神卡池").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			c, ok := control.Lookup("genshin")
			if !ok {
				ctx.SendChain(message.Text("找不到服务!"))
				return
			}
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			store := (storage)(c.GetData(gid))
			if store.setmode(!store.is5starsmode()) {
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Text("切换到五星卡池~"))
			} else {
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Text("切换到普通卡池~"))
			}
			err := c.SetData(gid, int64(store))
			if err != nil {
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Text("ERROR:", err))
			}
		})

	engine.OnFullMatch("原神十连", ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			zipfile := engine.DataFolder() + "Genshin.zip"
			_, err := engine.GetLazyData("Genshin.zip", false)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = parsezip(zipfile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			c, ok := control.Lookup("genshin")
			if !ok {
				ctx.SendChain(message.Text("找不到服务!"))
				return
			}
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			store := (storage)(c.GetData(gid))
			img, str, mode, err := randnums(10, store)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			b, cl := writer.ToBytes(img)
			if mode {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("恭喜你抽到了: \n", str), message.ImageBytes(b)))
			} else {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("十连成功~"), message.ImageBytes(b)))
			}
			cl()
		})
}

func randnums(nums int, store storage) (rgba *image.RGBA, str string, replyMode bool, err error) {
	var (
		fours, fives                  = make([]*zip.File, 0, 10), make([]*zip.File, 0, 10)                           // 抽到 四, 五星角色
		threeArms, fourArms, fiveArms = make([]*zip.File, 0, 10), make([]*zip.File, 0, 10), make([]*zip.File, 0, 10) // 抽到 三 , 四, 五星武器
		fourN, fiveN                  = 0, 0                                                                         // 抽到 四, 五星角色的数量
		bgs                           = make([]*zip.File, 0, 10)                                                     // 背景图片名
		threeN2, fourN2, fiveN2       = 0, 0, 0                                                                      // 抽到 三 , 四, 五星武器的数量
		hero, stars                   = make([]*zip.File, 0, 10), make([]*zip.File, 0, 10)                           // 角色武器名, 储存星级图标

		cicon                   = make([]*zip.File, 0, 10)                                                            // 元素图标
		fivebg, fourbg, threebg = filetree["five_bg.jpg"][0], filetree["four_bg.jpg"][0], filetree["three_bg.jpg"][0] // 背景图片名
		fivelen                 = len(filetree["five"])
		five2len                = len(filetree["five2"])
		threelen                = len(filetree["Three"])
		fourlen                 = len(filetree["four"])
		four2len                = len(filetree["four2"])
	)

	if totl%9 == 0 { // 累计9次加入一个五星
		switch rand.Intn(2) {
		case 0:
			fiveN++
			fives = append(fives, filetree["five"][rand.Intn(fivelen)])
		case 1:
			fiveN2++
			fiveArms = append(fiveArms, filetree["five2"][rand.Intn(five2len)])
		}
		nums--
	}

	if store.is5starsmode() { // 5星模式
		for i := 0; i < nums; i++ {
			switch rand.Intn(2) {
			case 0:
				fiveN++
				fives = append(fives, filetree["five"][rand.Intn(fivelen)])
			case 1:
				fiveN2++
				fiveArms = append(fiveArms, filetree["five2"][rand.Intn(five2len)])
			}
		}
	} else { // 默认模式
		for i := 0; i < nums; i++ {
			a := rand.Intn(1000) // 抽卡几率 三星80% 四星17% 五星3%
			switch {
			case a >= 0 && a <= 800:
				threeN2++
				threeArms = append(threeArms, filetree["Three"][rand.Intn(threelen)])
			case a > 800 && a <= 885:
				fourN++
				fours = append(fours, filetree["four"][rand.Intn(fourlen)]) // 随机角色
			case a > 885 && a <= 970:
				fourN2++
				fourArms = append(fourArms, filetree["four2"][rand.Intn(four2len)]) // 随机武器
			case a > 970 && a <= 985:
				fiveN++
				fives = append(fives, filetree["five"][rand.Intn(fivelen)])
			default:
				fiveN2++
				fiveArms = append(fiveArms, filetree["five2"][rand.Intn(five2len)])
			}
		}
		if fourN+fourN2 == 0 && threeN2 > 0 { // 没有四星时自动加入
			threeN2--
			threeArms = threeArms[:len(threeArms)-1]
			switch rand.Intn(2) {
			case 0:
				fourN++
				fours = append(fours, filetree["four"][rand.Intn(fourlen)]) // 随机角色
			case 1:
				fourN2++
				fourArms = append(fourArms, filetree["four2"][rand.Intn(four2len)]) // 随机武器
			}
		}
		_ = atomic.AddUint64(&totl, 1)
	}

	icon := func(f *zip.File) *zip.File {
		name := f.Name
		name = name[strings.LastIndex(name, "/")+1:strings.Index(name, "_")] + ".png"
		logrus.Debugln("[genshin]get named file", name)
		return filetree[name][0]
	}

	he := func(cnt int, id int, f *zip.File, bg *zip.File) {
		var hen *[]*zip.File
		for i := 0; i < cnt; i++ {
			switch id {
			case 1:
				hen = &threeArms
			case 2:
				hen = &fourArms
			case 3:
				hen = &fours
			case 4:
				hen = &fiveArms
			case 5:
				hen = &fives
			}
			bgs = append(bgs, bg) // 加入颜色背景
			hero = append(hero, (*hen)[i])
			stars = append(stars, f)               // 加入星级图标
			cicon = append(cicon, icon((*hen)[i])) // 加入元素图标
		}
	}

	if fiveN > 0 { // 按顺序加入
		he(fiveN, 5, starN5, fivebg) // 五星角色
		str += reply(fives, 1, str)
		replyMode = true
	}
	if fourN > 0 {
		he(fourN, 3, starN4, fourbg) // 四星角色
	}
	if fiveN2 > 0 {
		he(fiveN2, 4, starN5, fivebg) // 五星武器
		str += reply(fiveArms, 2, str)
		replyMode = true
	}
	if fourN2 > 0 {
		he(fourN2, 2, starN4, fourbg) // 四星武器
	}
	if threeN2 > 0 {
		he(threeN2, 1, starN3, threebg) // 三星武器
	}

	var c1, c2, c3 uint8 = 50, 50, 50 // 背景颜色

	img00, err := filetree["bg0.jpg"][0].Open() // 打开背景图片
	if err != nil {
		return
	}

	rectangle := image.Rect(0, 0, 1920, 1080) // 图片宽度, 图片高度
	rgba = image.NewRGBA(rectangle)
	draw.Draw(rgba, rgba.Bounds(), image.NewUniform(color.RGBA{c1, c2, c3, 255}), image.Point{}, draw.Over)
	context := freetype.NewContext() // 创建一个新的上下文
	context.SetDPI(72)               // 每英寸 dpi
	context.SetClip(rgba.Bounds())
	context.SetDst(rgba)

	defer img00.Close()
	img0, err := jpeg.Decode(img00) // 读取一个本地图像
	if err != nil {
		return
	}

	offset := image.Pt(0, 0) // 图片在背景上的位置
	draw.Draw(rgba, img0.Bounds().Add(offset), img0, image.Point{}, draw.Over)

	w1, h1 := 230, 0
	for i := 0; i < len(hero); i++ {
		if i > 0 {
			w1 += 146 // 图片宽度
		}

		imgs, err := bgs[i].Open() // 取出背景图片
		if err != nil {
			return nil, "", false, err
		}
		defer imgs.Close()

		img, _ := jpeg.Decode(imgs)
		offset := image.Pt(w1, h1)
		draw.Draw(rgba, img.Bounds().Add(offset), img, image.Point{}, draw.Over)

		imgs1, err := hero[i].Open() // 取出图片名
		if err != nil {
			return nil, "", false, err
		}
		defer imgs1.Close()

		img1, _ := png.Decode(imgs1)
		offset1 := image.Pt(w1, h1)
		draw.Draw(rgba, img1.Bounds().Add(offset1), img1, image.Point{}, draw.Over)

		imgs2, err := stars[i].Open() // 取出星级图标
		if err != nil {
			return nil, "", false, err
		}
		defer imgs2.Close()

		img2, _ := png.Decode(imgs2)
		offset2 := image.Pt(w1, h1)
		draw.Draw(rgba, img2.Bounds().Add(offset2), img2, image.Point{}, draw.Over)

		imgs3, err := cicon[i].Open() // 取出类型图标
		if err != nil {
			return nil, "", false, err
		}
		defer imgs3.Close()

		img3, _ := png.Decode(imgs3)
		offset3 := image.Pt(w1, h1)
		draw.Draw(rgba, img3.Bounds().Add(offset3), img3, image.Point{}, draw.Over)
	}
	imgs4, err := filetree["Reply.png"][0].Open() // "分享" 图标
	if err != nil {
		return nil, "", false, err
	}
	defer imgs4.Close()
	img4, err := png.Decode(imgs4)
	if err != nil {
		return nil, "", false, err
	}
	offset4 := image.Pt(1270, 945) // 宽, 高
	draw.Draw(rgba, img4.Bounds().Add(offset4), img4, image.Point{}, draw.Over)
	return
}

func parsezip(zipFile string) error {
	zipReader, err := zip.OpenReader(zipFile) // will not close
	if err != nil {
		return err
	}
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			filetree[f.Name] = make([]*zip.File, 0, 32)
			continue
		}
		f.Name = f.Name[8:]
		i := strings.LastIndex(f.Name, "/")
		if i < 0 {
			filetree[f.Name] = []*zip.File{f}
			logrus.Debugln("[genshin]insert file", f.Name)
			continue
		}
		folder := f.Name[:i]
		if folder != "" {
			filetree[folder] = append(filetree[folder], f)
			logrus.Debugln("[genshin]insert file into", folder)
			if folder == "gacha" {
				switch f.Name[i+1:] {
				case "ThreeStar.png":
					starN3 = f
				case "FourStar.png":
					starN4 = f
				case "FiveStar.png":
					starN5 = f
				}
			}
		}
	}
	return nil
}

// 取出角色武器名
func reply(z []*zip.File, num int, nameStr string) string {
	var tmp strings.Builder
	tmp.Grow(128)
	switch {
	case num == 1:
		tmp.WriteString("★五星角色★\n")
	case num == 2 && len(nameStr) > 0:
		tmp.WriteString("\n★五星武器★\n")
	default:
		tmp.WriteString("★五星武器★\n")
	}
	for i := range z {
		tmp.WriteString(namereg.FindStringSubmatch(z[i].Name)[1] + " * ")
	}
	return tmp.String()
}
