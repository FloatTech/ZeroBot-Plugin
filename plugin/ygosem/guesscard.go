// Package ygosem 基于ygosem的插件功能
package ygosem

import (
	"bytes"
	"image"
	"image/color"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	semurl = "https://www.ygo-sem.cn/"
	ua     = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"
)

func init() {
	engine := control.Register("guessygo", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏王猜卡游戏",
		Help:             "-猜卡游戏\n-(黑边|反色|马赛克|旋转|切图)猜卡游戏",
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
	engine.OnRegex("^(黑边|反色|马赛克|旋转|切图)?猜卡游戏$", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
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
		url := "https://www.ygo-sem.cn/Cards/Default.aspx"
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, "GET", url, ua, nil)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 获取卡牌数量
		listmax := regexp.MustCompile(`条 共:\s*(?s:(.*?))\s*条</span>`).FindAllStringSubmatch(helper.BytesToString(body), -1)
		if len(listmax) == 0 {
			ctx.SendChain(message.Text("数据存在错误: 无法获取当前卡池数量"))
			return
		}
		maxnumber, _ := strconv.Atoi(listmax[0][1])
		drawCard := strconv.Itoa(rand.Intn(maxnumber + 1))
		url = "https://www.ygo-sem.cn/Cards/S.aspx?q=" + drawCard
		// 获取卡片信息
		body, err = web.RequestDataWith(web.NewDefaultClient(), url, "GET", url, ua, nil)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 获取卡面信息
		cardData := getCarddata(helper.BytesToString(body))
		if cardData == (gameCardInfo{}) {
			ctx.SendChain(message.Text("数据存在错误: 无法获取卡片信息"))
			return
		}
		// 获取卡图连接
		picHref := regexp.MustCompile(`picsCN(/\d+/\d+).jpg`).FindAllStringSubmatch(helper.BytesToString(body), -1)
		if len(picHref) == 0 {
			ctx.SendChain(message.Text("数据存在错误: 无法获取卡图信息"))
			return
		}
		url = "https://www.ygo-sem.cn/yugioh/larg/" + picHref[0][1] + ".jpg"
		body, err = web.RequestDataWith(web.NewDefaultClient(), url, "GET", url, ua, nil)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 对卡图做处理
		pic, _, err := image.Decode(bytes.NewReader(body))
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		pictrue, cl := randPicture(pic, mode)
		defer cl()
		// 进行猜卡环节
		ctx.SendChain(message.Text("请回答下图的卡名\n以“我猜xxx”格式回答\n(xxx需包含卡名1/4以上)\n或发“提示”得提示;“取消”结束游戏"), message.ImageBytes(pictrue))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup,
			zero.RegexRule("^(我猜.*|提示|取消)"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		tick := time.NewTimer(105 * time.Second)
		over := time.NewTimer(120 * time.Second)
		wg := sync.WaitGroup{}
		var (
			messageStr  message.MessageSegment // 文本信息
			tickCount   = 0                    // 提示次数
			answerCount = 0                    // 问答次数
			win         bool                   // 是否结束游戏
		)
		for {
			select {
			case <-tick.C:
				ctx.SendChain(message.Text("还有15s作答时间"))
			case <-over.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("时间超时,游戏结束\n卡名是:\n", cardData.Name),
					message.ImageBytes(body)))
				return
			case c := <-recv:
				wg.Add(1)
				go func() {
					messageStr, answerCount, tickCount, win = gameMatch(c, ctx.Event.UserID, cardData, answerCount, tickCount)
					if win { // 游戏结束的话
						tick.Stop()
						over.Stop()
						ctx.SendChain(message.Reply(c.Event.MessageID), messageStr, message.ImageBytes(body))
					} else {
						tick.Reset(105 * time.Second)
						over.Reset(120 * time.Second)
						ctx.SendChain(message.Reply(c.Event.MessageID), messageStr)
					}
					wg.Done()
				}()
				wg.Wait()
			}
			if win {
				break
			}
		}
	})
}

// 随机选择
func randPicture(pic image.Image, mode int) ([]byte, func()) {
	dst := img.Size(pic, 256*5, 256*5)
	if mode == -1 {
		mode = rand.Intn(5)
	}
	switch mode {
	case 0:
		return setPicture(dst)
	case 1:
		return setBlur(dst)
	case 2:
		return doublePicture(dst)
	case 3:
		return setMark(dst)
	case 4:
		return cutPic(pic)
	default:
		return nil, nil
	}
}

// 获取黑边
func setPicture(dst *img.Factory) ([]byte, func()) {
	dst = dst.Invert().Grayscale()
	b := dst.Im.Bounds()
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			a := dst.Im.At(x, y)
			c := color.NRGBAModel.Convert(a).(color.NRGBA)
			if c.R > 127 || c.G > 127 || c.B > 127 {
				c.R = 255
				c.G = 255
				c.B = 255
			}
			dst.Im.Set(x, y, c)
		}
	}
	return writer.ToBytes(dst.Im)
}

// 旋转
func doublePicture(dst *img.Factory) ([]byte, func()) {
	b := dst.Im.Bounds()
	pic := dst.FlipH().FlipV()
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			a := pic.Im.At(x, y)
			c := color.NRGBAModel.Convert(a).(color.NRGBA)
			a1 := dst.Im.At(x, y)
			c1 := color.NRGBAModel.Convert(a1).(color.NRGBA)
			switch {
			case y > x && x < b.Max.X/2 && y < b.Max.Y/2:
				dst.Im.Set(x, y, c)
			case y < x && x > b.Max.X/2 && y > b.Max.Y/2:
				dst.Im.Set(x, y, c)
			case y > b.Max.Y-x && x < b.Max.X/2 && y > b.Max.Y/2:
				dst.Im.Set(x, y, c)
			case y < b.Max.Y-x && x > b.Max.X/2 && y < b.Max.Y/2:
				dst.Im.Set(x, y, c)
			default:
				dst.Im.Set(x, y, color.NRGBA{
					R: 255 - c1.R,
					G: 255 - c1.G,
					B: 255 - c1.B,
					A: 255,
				})
			}
		}
	}
	return writer.ToBytes(dst.Im)
}

// 反色
func setBlur(dst *img.Factory) ([]byte, func()) {
	b := dst.Im.Bounds()
	for y1 := b.Min.Y; y1 <= b.Max.Y; y1++ {
		for x1 := b.Min.X; x1 <= b.Max.X; x1++ {
			a := dst.Im.At(x1, y1)
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
				dst.Im.Set(x1, y1, c)
			}
		}
	}
	return writer.ToBytes(dst.Invert().Blur(10).Im)
}

// 马赛克
func setMark(dst *img.Factory) ([]byte, func()) {
	b := dst.Im.Bounds()
	markSize := 64

	for yOfMarknum := 0; yOfMarknum <= math.Ceil(b.Max.Y, markSize); yOfMarknum++ {
		for xOfMarknum := 0; xOfMarknum <= math.Ceil(b.Max.X, markSize); xOfMarknum++ {
			a := dst.Im.At(xOfMarknum*markSize+markSize/2, yOfMarknum*markSize+markSize/2)
			cc := color.NRGBAModel.Convert(a).(color.NRGBA)
			for y := 0; y < markSize; y++ {
				for x := 0; x < markSize; x++ {
					xOfPic := xOfMarknum*markSize + x
					yOfPic := yOfMarknum*markSize + y
					dst.Im.Set(xOfPic, yOfPic, cc)
				}
			}
		}
	}
	return writer.ToBytes(dst.Blur(3).Im)
}

// 随机切割
func cutPic(pic image.Image) ([]byte, func()) {
	indexOfx := rand.Intn(3)
	indexOfy := rand.Intn(3)
	indexOfx2 := rand.Intn(3)
	indexOfy2 := rand.Intn(3)
	dst := img.Size(pic, 256*5, 256*5)
	b := dst.Im.Bounds()
	bx := b.Max.X / 3
	by := b.Max.Y / 3
	returnpic := img.NewFactory(b.Max.X, b.Max.Y, color.NRGBA{255, 255, 255, 255})

	for yOfMarknum := b.Min.Y; yOfMarknum <= b.Max.Y; yOfMarknum++ {
		for xOfMarknum := b.Min.X; xOfMarknum <= b.Max.X; xOfMarknum++ {
			if xOfMarknum == bx || yOfMarknum == by || xOfMarknum == bx*2 || yOfMarknum == by*2 {
				//黑框
				returnpic.Im.Set(xOfMarknum, yOfMarknum, color.NRGBA{0, 0, 0, 0})
			}
			if xOfMarknum >= bx*indexOfx && xOfMarknum < bx*(indexOfx+1) {
				if yOfMarknum >= by*indexOfy && yOfMarknum < by*(indexOfy+1) {
					a := dst.Im.At(xOfMarknum, yOfMarknum)
					cc := color.NRGBAModel.Convert(a).(color.NRGBA)
					returnpic.Im.Set(xOfMarknum, yOfMarknum, cc)
				}
			}
			if xOfMarknum >= bx*indexOfx2 && xOfMarknum < bx*(indexOfx2+1) {
				if yOfMarknum >= by*indexOfy2 && yOfMarknum < by*(indexOfy2+1) {
					a := dst.Im.At(xOfMarknum, yOfMarknum)
					cc := color.NRGBAModel.Convert(a).(color.NRGBA)
					returnpic.Im.Set(xOfMarknum, yOfMarknum, cc)
				}
			}
		}
	}
	return writer.ToBytes(returnpic.Im)
}

// 数据匹配（结果信息，答题次数，提示次数，是否结束游戏）
func gameMatch(c *zero.Ctx, beginner int64, cardData gameCardInfo, answerCount, tickCount int) (message.MessageSegment, int, int, bool) {
	answer := c.Event.Message.String()
	switch answer {
	case "取消":
		if c.Event.UserID == beginner {
			return message.Text("游戏已取消\n卡名是:\n", cardData.Name), answerCount, tickCount, true
		}
		return message.Text("你无权限取消"), answerCount, tickCount, false
	case "提示":
		if tickCount > 3 {
			return message.Text("已经没有提示了哦"), answerCount, tickCount, false
		}
		tips := getTips(cardData, tickCount)
		tickCount++
		return message.Text(tips), answerCount, tickCount, false
	default:
		_, answer, _ := strings.Cut(answer, "我猜")
		name := []rune(cardData.Name)
		switch {
		case len([]rune(answer)) < math.Ceil(len(name), 4):
			return message.Text("请输入", math.Ceil(len(name), 4), "字以上"), answerCount, tickCount, false
		case strings.Contains(cardData.Name, answer):
			return message.Text("太棒了,你猜对了!\n卡名是:\n", cardData.Name), answerCount, tickCount, true
		}
		answerCount++
		if answerCount < 6 {
			return message.Text("答案不对哦,还有", 6-answerCount, "次机会,加油啊~"), answerCount, tickCount, false
		}
		tickCount++
		return message.Text("次数到了,很遗憾没能猜出来\n卡名是:\n", cardData.Name), answerCount, tickCount, true
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
