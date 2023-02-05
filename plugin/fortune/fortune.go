// Package fortune 每日运势
package fortune

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"image"
	"io"
	"os"
	"strconv"

	"github.com/Coloured-glaze/gg" // 注册了 jpg png gif
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/pool"
)

const (
	// 底图缓存位置
	images = "data/Fortune/"
	// 基础文件位置
	omikujson = "data/Fortune/text.json"
	// 字体文件位置
	font = "data/Font/sakura.ttf"
	// 生成图缓存位置
	cache = images + "cache/"
)

var (
	// 底图类型列表
	table = [...]string{"车万", "DC4", "爱因斯坦", "星空列车", "樱云之恋", "富婆妹", "李清歌", "公主连结", "原神", "明日方舟", "碧蓝航线", "碧蓝幻想", "战双", "阴阳师", "赛马娘", "东方归言录", "奇异恩典", "夏日口袋", "ASoul", "Hololive"}
	// 映射底图与 index
	index = make(map[string]uint8)
	// 签文
	omikujis []map[string]string
)

func init() {
	// 插件主体
	en := control.Register("fortune", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "每日运势",
		Help: "- 运势 | 抽签\n" +
			"- 设置底图[车万 | DC4 | 爱因斯坦 | 星空列车 | 樱云之恋 | 富婆妹 | 李清歌 | 公主连结 | 原神 | 明日方舟 | 碧蓝航线 | 碧蓝幻想 | 战双 | 阴阳师 | 赛马娘 | 东方归言录 | 奇异恩典 | 夏日口袋 | ASoul | Hololive]",
		PublicDataFolder: "Fortune",
	})
	_ = os.RemoveAll(cache)
	err := os.MkdirAll(cache, 0755)
	if err != nil {
		panic(err)
	}
	for i, s := range table {
		index[s] = uint8(i)
	}
	en.OnRegex(`^设置底图\s?(.*)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			i, ok := index[ctx.State["regex_matched"].([]string)[1]]
			if ok {
				c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
				if ok {
					err := c.SetData(gid, int64(i)&0xff)
					if err != nil {
						ctx.SendChain(message.Text("设置失败:", err))
						return
					}
					ctx.SendChain(message.Text("设置成功~"))
					return
				}
				ctx.SendChain(message.Text("设置失败: 找不到插件"))
				return
			}
			ctx.SendChain(message.Text("没有这个底图哦～"))
		})
	en.OnFullMatchGroup([]string{"运势", "抽签"}, fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			data, err := file.GetLazyData(omikujson, control.Md5File, false)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = json.Unmarshal(data, &omikujis)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			_, err = file.GetLazyData(font, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			return true
		},
	)).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取该群背景类型，默认车万
			kind := "车万"
			gid := ctx.Event.GroupID
			if gid <= 0 {
				// 个人用户设为负数
				gid = -ctx.Event.UserID
			}
			logrus.Debugln("[fortune]gid:", ctx.Event.GroupID, "uid:", ctx.Event.UserID)
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if ok {
				v := uint8(c.GetData(gid) & 0xff)
				if int(v) < len(table) {
					kind = table[v]
				}
			}
			// 检查背景图片是否存在
			zipfile := images + kind + ".zip"
			_, err := file.GetLazyData(zipfile, control.Md5File, false)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			// 随机获取背景
			background, index, err := randimage(zipfile, ctx)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			// 随机获取签文
			randtextindex := fcext.RandSenderPerDayN(ctx.Event.UserID, len(omikujis))
			title, text := omikujis[randtextindex]["title"], omikujis[randtextindex]["content"]
			digest := md5.Sum(helper.StringToBytes(zipfile + strconv.Itoa(index) + title + text))
			cachefile := cache + hex.EncodeToString(digest[:])

			err = pool.SendImageFromPool(cachefile, cachefile, func() error {
				f, err := os.Create(cachefile)
				if err != nil {
					return err
				}
				_, err = draw(background, title, text, f)
				_ = f.Close()
				return err
			}, ctxext.Send(ctx), ctxext.GetMessage(ctx))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		})
}

// @function randimage 随机选取zip内的文件
// @param path zip路径
// @param ctx *zero.Ctx
// @return 文件路径 & 错误信息
func randimage(path string, ctx *zero.Ctx) (im image.Image, index int, err error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	defer reader.Close()

	file := reader.File[fcext.RandSenderPerDayN(ctx.Event.UserID, len(reader.File))]
	f, err := file.Open()
	if err != nil {
		return
	}
	defer f.Close()

	im, _, err = image.Decode(f)
	return
}

// @function draw 绘制运势图
// @param background 背景图片路径
// @param seed 随机数种子
// @param title 签名
// @param text 签文
// @return 错误信息
func draw(back image.Image, title, txt string, f io.Writer) (int64, error) {
	canvas := gg.NewContext(back.Bounds().Size().Y, back.Bounds().Size().X)
	canvas.DrawImage(back, 0, 0)
	// 写标题
	canvas.SetRGB(1, 1, 1)
	if err := canvas.LoadFontFace(font, 45); err != nil {
		return -1, err
	}
	sw, _ := canvas.MeasureString(title)
	canvas.DrawString(title, 140-sw/2, 112)
	// 写正文
	canvas.SetRGB(0, 0, 0)
	if err := canvas.LoadFontFace(font, 23); err != nil {
		return -1, err
	}
	tw, th := canvas.MeasureString("测")
	tw, th = tw+10, th+10
	r := []rune(txt)
	xsum := rowsnum(len(r), 9)
	switch xsum {
	default:
		for i, o := range r {
			xnow := rowsnum(i+1, 9)
			ysum := math.Min(len(r)-(xnow-1)*9, 9)
			ynow := i%9 + 1
			canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(ysum, ynow, th)+320.0)
		}
	case 2:
		div := rowsnum(len(r), 2)
		for i, o := range r {
			xnow := rowsnum(i+1, div)
			ysum := math.Min(len(r)-(xnow-1)*div, div)
			ynow := i%div + 1
			switch xnow {
			case 1:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow, th)+320.0)
			case 2:
				canvas.DrawString(string(o), -offest(xsum, xnow, tw)+115, offest(9, ynow+(9-ysum), th)+320.0)
			}
		}
	}
	return writer.WriteTo(canvas.Image(), f)
}

func offest(total, now int, distance float64) float64 {
	if total%2 == 0 {
		return (float64(now-total/2) - 1) * distance
	}
	return (float64(now-total/2) - 1.5) * distance
}

func rowsnum(total, div int) int {
	temp := total / div
	if total%div != 0 {
		temp++
	}
	return temp
}
