// Package wife 抽老婆
package wife

import (
	"errors"
	"image/color"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	zbmath "github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/imgfactory"
)

var (
	sizeList = []int{0, 3, 5, 8}
	enguess  = control.Register("wifegame", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "- 猜老婆",
		Brief:            "从老婆库猜老婆",
	}).ApplySingle(ctxext.NewGroupSingle("已经有正在进行的游戏..."))
)

func init() {
	// _ = os.MkdirAll(engine.DataFolder()+"wives", 0755)
	enguess.OnFullMatch("猜老婆").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		var err error
		class := 3

		fileName, err := lottery()
		if err != nil {
			ctx.SendChain(message.Text("[猜老婆]error:\n", err))
			return
		}

		work, name := card2name(fileName)
		picFile := file.BOTPATH + "/" + engine.DataFolder() + "wives/" + fileName
		pic, err := os.ReadFile(picFile)
		if err != nil {
			ctx.SendChain(message.Text("[猜老婆]error:\n", err))
			return
		}
		img, err := gg.LoadImage(picFile)
		if err != nil {
			ctx.SendChain(message.Text("[猜老婆]error:\n", err))
			return
		}
		dst := imgfactory.Size(img, img.Bounds().Dx(), img.Bounds().Dy())
		q, err := mosaic(dst, class)
		if err != nil {
			ctx.SendChain(
				message.Reply(ctx.Event.MessageID),
				message.Text("[猜老婆]图片生成失败:\n", err),
			)
			return
		}
		if id := ctx.SendChain(
			message.ImageBytes(q),
		); id.ID() != 0 {
			ctx.SendChain(message.Text("请回答该二次元角色名字\n以“xxx酱”格式回答"))
		}
		var next *zero.FutureEvent
		if ctx.Event.GroupID == 0 {
			next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(·)?.+酱$`), ctx.CheckSession())
		} else {
			next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(`(·)?.+酱$`), zero.CheckGroup(ctx.Event.GroupID))
		}
		recv, cancel := next.Repeat()
		defer cancel()
		tick := time.NewTimer(105 * time.Second)
		after := time.NewTimer(120 * time.Second)
		for {
			select {
			case <-tick.C:
				ctx.SendChain(message.Text("[猜老婆]你还有15s作答时间"))
			case <-after.C:
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.ImageBytes(pic),
						message.Text("[猜老婆]倒计时结束，游戏结束...\n角色是:\n", name, "\n出自《", work, "》\n"),
					),
				)
				return
			case c := <-recv:
				tick.Reset(105 * time.Second)
				after.Reset(120 * time.Second)
				msg := c.Event.Message.String()
				msg, _, _ = strings.Cut(msg, "酱")
				class--
				if strings.Contains(name, msg) {
					if msgID := ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
						message.Text("太棒了,你猜对了!\n角色是:\n", name, "\n出自《", work, "》\n"),
						message.ImageBytes(pic))); msgID.ID() == 0 {
						ctx.SendChain(message.Text("太棒了,你猜对了!\n图片发送失败,可能被风控\n角色是:\n", name, "\n出自《", work, "》"))
					}
					return
				}
				if class < 1 {
					if msgID := ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
						message.Text("很遗憾,次数到了,游戏结束!\n角色是:\n", name, "\n出自《", work, "》\n"),
						message.ImageBytes(pic))); msgID.ID() == 0 {
						ctx.SendChain(message.Text("很遗憾,次数到了,游戏结束!\n图片发送失败,可能被风控\n角色是:\n", name, "\n出自《", work, "》"))
					}
					return
				}
				q, err = mosaic(dst, class)
				if err != nil {
					ctx.SendChain(
						message.Text("回答错误,你还有", class, "次机会\n请继续作答\n(提示：", work, ")"),
					)
					continue
				}
				ctx.SendChain(
					message.Text("回答错误,你还有", class, "次机会\n请继续作答(难度降低)\n"),
					message.ImageBytes(q),
				)
				continue
			}
		}
	})
}

// 从本地图库随机抽取，规避网络问题
func lottery() (fileName string, err error) {
	path := engine.DataFolder() + "wives" + "/"
	if file.IsNotExist(path) {
		err = errors.New("图库文件夹不存在,请先发送“抽老婆”扩展图库")
		return
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}
	// 如果本地列表为空
	if len(files) == 0 {
		err = errors.New("本地数据为0,请先发送“抽老婆”扩展图库")
		return
	}
	fileName = randPicture(files, 10)
	if fileName == "" {
		err = errors.New("抽取图库轮空了,请重试")
	}
	return
}

func randPicture(files []fs.DirEntry, indexMax int) (fileName string) {
	if len(files) > 1 {
		picture := files[rand.Intn(len(files))]
		// 如果是文件夹就递归
		if picture.IsDir() {
			indexMax--
			if indexMax <= 0 {
				return
			}
			fileName = randPicture(files, indexMax)
		} else {
			fileName = picture.Name()
		}
	} else {
		music := files[0]
		if !music.IsDir() {
			fileName = files[0].Name()
		}
	}
	return
}

// 马赛克生成
func mosaic(dst *imgfactory.Factory, level int) ([]byte, error) {
	b := dst.Image().Bounds()
	p := imgfactory.NewFactoryBG(dst.W(), dst.H(), color.NRGBA{255, 255, 255, 255})
	markSize := zbmath.Max(b.Max.X, b.Max.Y) * sizeList[level] / 200

	for yOfMarknum := 0; yOfMarknum <= zbmath.Ceil(b.Max.Y, markSize); yOfMarknum++ {
		for xOfMarknum := 0; xOfMarknum <= zbmath.Ceil(b.Max.X, markSize); xOfMarknum++ {
			a := dst.Image().At(xOfMarknum*markSize+markSize/2, yOfMarknum*markSize+markSize/2)
			cc := color.NRGBAModel.Convert(a).(color.NRGBA)
			for y := 0; y < markSize; y++ {
				for x := 0; x < markSize; x++ {
					xOfPic := xOfMarknum*markSize + x
					yOfPic := yOfMarknum*markSize + y
					p.Image().Set(xOfPic, yOfPic, cc)
				}
			}
		}
	}
	return imgfactory.ToBytes(p.Blur(3).Image())
}
