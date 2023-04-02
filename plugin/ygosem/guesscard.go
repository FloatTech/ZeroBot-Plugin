// Package ygosem 基于ygosem的插件功能
package ygosem

import (
	"image"
	"image/color"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	semurl = "https://www.ygo-sem.cn/"
	ua     = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"
)

type carddb struct {
	db *sql.Sqlite
	sync.RWMutex
}

var (
	mu        sync.RWMutex
	carddatas = &carddb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("guessygo", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "游戏王猜卡游戏",
		Help:              "-猜卡游戏\n-(黑边|反色|马赛克|旋转|切图)猜卡游戏",
		PrivateDataFolder: "ygosemdata",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("已经有正在进行的游戏..."),
				),
			)
		}),
	))
	cachePath = engine.DataFolder() + "pics/"
	getdb     = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		carddatas.db.DBPath = engine.DataFolder() + "carddata.db"
		err := carddatas.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		err = carddatas.db.Create("cards", &picInfos{})
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func init() {
	go func() {
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
	}()
	engine.OnRegex("^(黑边|反色|马赛克|旋转|切图)?猜卡游戏$", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("正在准备题目,请稍等"))
		mode := -1
		switch ctx.State["regex_matched"].([]string)[1] {
		case "黑边":
			mode = 0
		case "反色":
			mode = 1
		case "马赛克":
			mode = 2
		case "旋转":
			mode = 3
		case "切图":
			mode = 4
		}
		semdata, picFile, err := getSemData()
		if err == nil {
			err = carddatas.insert(picInfos{text: semdata, picFile: picFile})
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]", err))
			}
		} else {
			ctx.SendChain(message.Text("[ERROR]", err))
			semdata, picFile, err = carddatas.pick()
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]", err))
				return
			}
		}
		picFile = cachePath + picFile
		// 对卡图做处理
		pictrue, err := randPicture(picFile, mode)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 进行猜卡环节
		game := newGame(semdata)
		ctx.SendChain(message.Text("请回答下图的卡名\n以“我猜xxx”格式回答\n(xxx需包含卡名1/4以上)\n或发“提示”得提示;“取消”结束游戏"), message.ImageBytes(pictrue))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup,
			zero.RegexRule("^((我猜.+)|提示|取消)$"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		// defer cancel()
		tick := time.NewTimer(105 * time.Second)
		over := time.NewTimer(120 * time.Second)
		var (
			tickCount   = 0 // 提示次数
			answerCount = 0 // 问答次数
		)
		for {
			select {
			case <-tick.C:
				ctx.SendChain(message.Text("还有15s作答时间"))
			case <-over.C:
				cancel()
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("时间超时,游戏结束\n卡名是:\n", semdata.Name),
					message.Image("file:///"+file.BOTPATH+"/"+picFile)))
				return
			case c := <-recv:
				answer := c.Event.Message.String()
				if answer == "取消" {
					if c.Event.UserID == ctx.Event.UserID {
						cancel()
						tick.Stop()
						over.Stop()
						ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("游戏已取消\n卡名是:\n", semdata.Name),
							message.Image("file:///"+file.BOTPATH+"/"+picFile)))
						return
					}
					ctx.Send(message.ReplyWithMessage(c.Event.MessageID, message.Text("你无权限取消")))
					return
				}
				if answer == "提示" {
					if tickCount > 3 {
						tick.Reset(105 * time.Second)
						over.Reset(120 * time.Second)
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID, message.Text("已经没有提示了哦,加油啊")))
						continue
					}
					tickCount++
				} else {
					_, answer, _ = strings.Cut(answer, "我猜")
					answerCount++
				}
				messageStr, win := game(answer, tickCount-1, answerCount)
				if win {
					cancel()
					tick.Stop()
					over.Stop()
					ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
						message.Text(messageStr),
						message.Image("file:///"+file.BOTPATH+"/"+picFile)))
					return
				}
				if answerCount >= 6 {
					cancel()
					tick.Stop()
					over.Stop()
					ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
						message.Text("次数到了,很遗憾没能猜出来\n卡名是:\n"+semdata.Name),
						message.Image("file:///"+file.BOTPATH+"/"+picFile)))
					return
				}
				tick.Reset(105 * time.Second)
				over.Reset(120 * time.Second)
				ctx.Send(message.ReplyWithMessage(c.Event.MessageID, message.Text(messageStr)))
			}
		}
	})
}

// 随机选择
func randPicture(picFile string, mode int) ([]byte, error) {
	pic, err := gg.LoadImage(picFile)
	if err != nil {
		return nil, err
	}
	dst := imgfactory.Size(pic, 256*5, 256*5)
	if mode == -1 {
		mode = rand.Intn(5)
	}
	switch mode {
	case 0:
		return setPicture(dst)
	case 1:
		return setBlur(dst)
	case 2:
		return setMark(dst)
	case 3:
		return doublePicture(dst)
	case 4:
		return cutPic(pic)
	default:
		return nil, nil
	}
}

// 获取黑边
func setPicture(dst *imgfactory.Factory) ([]byte, error) {
	dst = dst.Invert().Grayscale()
	b := dst.Image().Bounds()
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			a := dst.Image().At(x, y)
			c := color.NRGBAModel.Convert(a).(color.NRGBA)
			if c.R > 127 || c.G > 127 || c.B > 127 {
				c.R = 255
				c.G = 255
				c.B = 255
			}
			dst.Image().Set(x, y, c)
		}
	}
	return imgfactory.ToBytes(dst.Image())
}

// 旋转
func doublePicture(dst *imgfactory.Factory) ([]byte, error) {
	b := dst.Image().Bounds()
	pic := dst.FlipH().FlipV()
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			a := pic.Image().At(x, y)
			c := color.NRGBAModel.Convert(a).(color.NRGBA)
			a1 := dst.Image().At(x, y)
			c1 := color.NRGBAModel.Convert(a1).(color.NRGBA)
			switch {
			case y > x && x < b.Max.X/2 && y < b.Max.Y/2:
				dst.Image().Set(x, y, c)
			case y < x && x > b.Max.X/2 && y > b.Max.Y/2:
				dst.Image().Set(x, y, c)
			case y > b.Max.Y-x && x < b.Max.X/2 && y > b.Max.Y/2:
				dst.Image().Set(x, y, c)
			case y < b.Max.Y-x && x > b.Max.X/2 && y < b.Max.Y/2:
				dst.Image().Set(x, y, c)
			default:
				dst.Image().Set(x, y, color.NRGBA{
					R: 255 - c1.R,
					G: 255 - c1.G,
					B: 255 - c1.B,
					A: 255,
				})
			}
		}
	}
	return imgfactory.ToBytes(dst.Image())
}

// 反色
func setBlur(dst *imgfactory.Factory) ([]byte, error) {
	b := dst.Image().Bounds()
	for y1 := b.Min.Y; y1 <= b.Max.Y; y1++ {
		for x1 := b.Min.X; x1 <= b.Max.X; x1++ {
			a := dst.Image().At(x1, y1)
			c := color.NRGBAModel.Convert(a).(color.NRGBA)
			if c.R > 128 || c.G > 128 || c.B > 128 {
				switch rand.Intn(6) {
				case 0: // 红
					c.R, c.G, c.B = uint8(rand.Intn(50)+180), uint8(rand.Intn(30)), uint8(rand.Intn(80)+40)
				case 1: // 橙
					c.R, c.G, c.B = uint8(rand.Intn(40)+210), uint8(rand.Intn(50)+70), uint8(rand.Intn(50)+20)
				case 2: // 黄
					c.R, c.G, c.B = uint8(rand.Intn(40)+210), uint8(rand.Intn(50)+170), uint8(rand.Intn(110)+40)
				case 3: // 绿
					c.R, c.G, c.B = uint8(rand.Intn(60)+80), uint8(rand.Intn(80)+140), uint8(rand.Intn(60)+80)
				case 4: // 蓝
					c.R, c.G, c.B = uint8(rand.Intn(60)+80), uint8(rand.Intn(50)+170), uint8(rand.Intn(50)+170)
				case 5: // 紫
					c.R, c.G, c.B = uint8(rand.Intn(60)+80), uint8(rand.Intn(60)+60), uint8(rand.Intn(50)+170)
				}
				dst.Image().Set(x1, y1, c)
			}
		}
	}
	return imgfactory.ToBytes(dst.Invert().Blur(10).Image())
}

// 马赛克
func setMark(dst *imgfactory.Factory) ([]byte, error) {
	b := dst.Image().Bounds()
	markSize := 64 * (1 + rand.Intn(2))

	for yOfMarknum := 0; yOfMarknum <= math.Ceil(b.Max.Y, markSize); yOfMarknum++ {
		for xOfMarknum := 0; xOfMarknum <= math.Ceil(b.Max.X, markSize); xOfMarknum++ {
			a := dst.Image().At(xOfMarknum*markSize+markSize/2, yOfMarknum*markSize+markSize/2)
			cc := color.NRGBAModel.Convert(a).(color.NRGBA)
			for y := 0; y < markSize; y++ {
				for x := 0; x < markSize; x++ {
					xOfPic := xOfMarknum*markSize + x
					yOfPic := yOfMarknum*markSize + y
					dst.Image().Set(xOfPic, yOfPic, cc)
				}
			}
		}
	}
	return imgfactory.ToBytes(dst.Blur(3).Image())
}

// 随机切割
func cutPic(pic image.Image) ([]byte, error) {
	indexOfx := rand.Intn(3)
	indexOfy := rand.Intn(3)
	indexOfx2 := rand.Intn(3)
	indexOfy2 := rand.Intn(3)
	dst := imgfactory.Size(pic, 256*5, 256*5)
	b := dst.Image()
	bx := b.Bounds().Max.X / 3
	by := b.Bounds().Max.Y / 3
	returnpic := imgfactory.NewFactoryBG(dst.W(), dst.H(), color.NRGBA{255, 255, 255, 255})

	for yOfMarknum := b.Bounds().Min.Y; yOfMarknum <= b.Bounds().Max.Y; yOfMarknum++ {
		for xOfMarknum := b.Bounds().Min.X; xOfMarknum <= b.Bounds().Max.X; xOfMarknum++ {
			if xOfMarknum == bx || yOfMarknum == by || xOfMarknum == bx*2 || yOfMarknum == by*2 {
				// 黑框
				returnpic.Image().Set(xOfMarknum, yOfMarknum, color.NRGBA{0, 0, 0, 0})
			}
			if xOfMarknum >= bx*indexOfx && xOfMarknum < bx*(indexOfx+1) {
				if yOfMarknum >= by*indexOfy && yOfMarknum < by*(indexOfy+1) {
					a := dst.Image().At(xOfMarknum, yOfMarknum)
					cc := color.NRGBAModel.Convert(a).(color.NRGBA)
					returnpic.Image().Set(xOfMarknum, yOfMarknum, cc)
				}
			}
			if xOfMarknum >= bx*indexOfx2 && xOfMarknum < bx*(indexOfx2+1) {
				if yOfMarknum >= by*indexOfy2 && yOfMarknum < by*(indexOfy2+1) {
					a := dst.Image().At(xOfMarknum, yOfMarknum)
					cc := color.NRGBAModel.Convert(a).(color.NRGBA)
					returnpic.Image().Set(xOfMarknum, yOfMarknum, cc)
				}
			}
		}
	}
	return imgfactory.ToBytes(returnpic.Image())
}

func newGame(cardData gameCardInfo) func(string, int, int) (string, bool) {
	return func(s string, stickCount, answerCount int) (message string, win bool) {
		switch s {
		case "提示":
			tips := getTips(cardData, stickCount)
			return tips, false
		default:
			name := []rune(cardData.Name)
			switch {
			case len([]rune(s)) < math.Ceil(len(name), 4):
				return "请输入" + strconv.Itoa(math.Ceil(len(name), 4)) + "字以上", false
			case strings.Contains(cardData.Name, s):
				return "太棒了,你猜对了!\n卡名是:\n" + cardData.Name, true
			}
		}
		return "答案不对哦,还有" + strconv.Itoa(6-answerCount) + "次回答机会,加油啊~", false
	}
}

// 拼接提示词
func getTips(cardData gameCardInfo, quitCount int) string {
	name := []rune(cardData.Name)
	switch quitCount {
	case 0:
		return "这是一张" + cardData.Type + ",卡名是" + strconv.Itoa(len(name)) + "字的"
	case 3:
		return "卡名含有: " + string(name[rand.Intn(len(name))])
	default:
		var textrand []string
		depict := strings.Split(cardData.Depict, "。")
		for _, value := range depict {
			if value != "" {
				list := strings.Split(value, "，")
				for _, value2 := range list {
					if value2 != "" {
						textrand = append(textrand, value2)
					}
				}
			}
		}
		if strings.Contains(cardData.Type, "怪兽") {
			text := []string{
				"这只怪兽的属性是" + cardData.Attr,
				"这只怪兽的种族是" + cardData.Race,
				"这只怪兽的等级/阶级/连接值是" + cardData.Level,
				"这只怪兽的效果/描述含有:\n" + textrand[rand.Intn(len(textrand))],
			}
			return text[rand.Intn(len(text))]
		}
		return textrand[rand.Intn(len(textrand))]
	}
}

type picInfos struct {
	text    gameCardInfo
	picFile string
}

func (sql *carddb) insert(dbInfo picInfos) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("cards", &picInfos{})
	if err == nil {
		return err
	}
	return sql.db.Insert("cards", &dbInfo)
}

func (sql *carddb) pick() (dbInfo gameCardInfo, picFile string, err error) {
	sql.RLock()
	defer sql.RUnlock()
	info := picInfos{}
	err = sql.db.Pick("cards", &info)
	if err != nil {
		return
	}
	return info.text, info.picFile, nil
}
