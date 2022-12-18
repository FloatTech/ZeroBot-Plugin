// Package gamesystem ...
package gamesystem

import (
	"bytes"
	"image"
	"image/color"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
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
	url = "https://www.ygo-sem.cn/"
	ua  = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"
)

type gameCardInfo struct {
	Name   string //卡名
	Type   string //种类
	Race   string //种族
	Attr   string //属性
	Level  string //等级
	Depict string //效果
}

func init() {
	engine := control.Register("guessygo", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏王猜卡游戏",
		Help:             "-猜卡游戏",
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
	engine.OnFullMatch("猜卡游戏").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("正在准备题目,请稍等"))
		url := "https://www.ygo-sem.cn/Cards/Default.aspx"
		// 请求html页面
		body, err := web.RequestDataWith(web.NewDefaultClient(), url, "GET", url, ua)
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
		body, err = web.RequestDataWith(web.NewDefaultClient(), url, "GET", url, ua)
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
		body, err = web.RequestDataWith(web.NewDefaultClient(), url, "GET", url, ua)
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
		dst := img.Size(pic, 256*5, 256*5).Grayscale()
		b := dst.Im.Bounds()
		for y1 := b.Min.Y; y1 <= b.Max.Y; y1++ {
			for x1 := b.Min.X; x1 <= b.Max.X; x1++ {
				a := dst.Im.At(x1, y1)
				c := color.NRGBAModel.Convert(a).(color.NRGBA)
				if c.R > 50 && c.G > 50 && c.B > 50 {
					c.R = 255
					c.G = 255
					c.B = 255
				}
				dst.Im.Set(x1, y1, c)
			}
		}
		pictrue, cl := writer.ToBytes(dst.Im)
		ctx.SendChain(message.Text("请回答下图的卡名\n以“我猜xxx”格式回答\n(xxx需包含卡名1/4以上)\n或发“提示”得提示;“取消”结束游戏"), message.ImageBytes(pictrue))
		cl()
		// 进行猜歌环节
		var quitCount = 0   // 音频数量
		var answerCount = 0 // 问答次数
		name := []rune(cardData.Name)
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup,
			zero.RegexRule("^(我猜.*|提示|取消)"), zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		tick := time.NewTimer(105 * time.Second)
		after := time.NewTimer(120 * time.Second)
		for {
			select {
			case <-tick.C:
				ctx.SendChain(message.Text("还有15s作答时间"))
			case <-after.C:
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("时间超时,游戏结束\n卡名是:\n", cardData.Name),
					message.ImageBytes(body)))
				return
			case c := <-recv:
				tick.Reset(105 * time.Second)
				after.Reset(120 * time.Second)
				answer := c.Event.Message.String()
				switch answer {
				case "取消":
					if c.Event.UserID == ctx.Event.UserID {
						tick.Stop()
						after.Stop()
						ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("游戏已取消\n卡名是:\n", cardData.Name),
							message.ImageBytes(body)))
						return
					}
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.Text("你无权限取消"),
						),
					)
				case "提示":
					if quitCount > 3 {
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("已经没有提示了哦"),
							),
						)
						continue
					}
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.Text(getTips(cardData, quitCount)),
						),
					)
					quitCount++
				default:
					_, answer, _ := strings.Cut(answer, "我猜")
					if len([]rune(answer)) < math.Ceil(len(name), 4) {
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("请输入", math.Ceil(len(name), 4), "字以上"),
							),
						)
						continue
					}
					if strings.Contains(cardData.Name, answer) {
						tick.Stop()
						after.Stop()
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("太棒了,你猜对了!\n卡名是:\n", cardData.Name),
							message.ImageBytes(body)))
						return
					}
					answerCount++
					switch {
					case answerCount < 6:
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("答案不对哦,加油啊~"),
							),
						)
					default:
						tick.Stop()
						after.Stop()
						ctx.Send(message.ReplyWithMessage(c.Event.MessageID,
							message.Text("次数到了,很遗憾没能猜出来\n卡名是:\n", cardData.Name),
							message.ImageBytes(body)))
						return
					}
				}
			}
		}
	})
}

// 获取卡面信息
func getCarddata(body string) (cardData gameCardInfo) {
	// 获取卡名
	cardName := regexp.MustCompile(`<b>中文名</b> </span>&nbsp;<span class="item_box_value">\s*(.*)</span>\s*</div>`).FindAllStringSubmatch(body, -1)
	if len(cardName) == 0 {
		return
	}
	cardData.Name = cardName[0][1]
	// 种类
	cardType := regexp.MustCompile(`<b>卡片种类</b> </span>&nbsp;<span class="item_box_value" id="dCnType">\s*(.*?)\s*</span>\s*<span`).FindAllStringSubmatch(body, -1)
	cardData.Type = cardType[0][1]
	if strings.Contains(cardType[0][1], "怪兽") {
		// 种族
		cardRace := regexp.MustCompile(`<span id="dCnRace" class="item_box_value">\s*(.*)\s*</span>\s*<span id="dEnRace"`).FindAllStringSubmatch(body, -1)
		cardData.Race = cardRace[0][1]
		// 属性
		cardAttr := regexp.MustCompile(`<b>属性</b> </span>&nbsp;<span class="item_box_value" id="attr">\s*(.*)\s*</span>`).FindAllStringSubmatch(body, -1)
		cardData.Attr = cardAttr[0][1]
		/*星数*/
		switch {
		case strings.Contains(cardType[0][1], "连接"):
			cardLevel := regexp.MustCompile(`<span class="item_box_value">(LINK.*)</span>`).FindAllStringSubmatch(body, -1)
			cardData.Level = cardLevel[0][1]
		default:
			cardLevel := regexp.MustCompile(`<b>星数/阶级</b> </span><span class=\"item_box_value\">\s*(.*)\s*</span>`).FindAllStringSubmatch(body, -1)
			cardData.Level = cardLevel[0][1]
		}
	}
	/*效果*/
	cardDepict := regexp.MustCompile(`<div class="item_box_text" id="cardDepict">\s*(?s:(.*?))\s*</div>`).FindAllStringSubmatch(body, -1)
	cardData.Depict = cardDepict[0][1]
	return
}

// 拼接提示词
func getTips(cardData gameCardInfo, quitCount int) string {
	name := []rune(cardData.Name)
	depict := []rune(cardData.Depict)
	index := rand.Intn(len(depict) / 2)
	switch quitCount {
	case 0:
		return "这是一张" + cardData.Type + ",卡名是" + strconv.Itoa(len(name)) + "字的"
	case 3:
		return "卡名含有: " + string(name[rand.Intn(len(name))])
	default:
		if strings.Contains(cardData.Type, "怪兽") {
			textrand := []string{
				"这只怪兽的属性是" + cardData.Attr,
				"这只怪兽的种族是" + cardData.Race,
				"这只怪兽是" + cardData.Type,
				"这只怪兽的等级/阶级/连接值是" + cardData.Level,
				"这只怪兽的效果/描述含有:\n" + string(depict[5+index:15+index]),
			}
			return textrand[rand.Intn(len(textrand))]
		} else {
			depict := strings.Split(cardData.Depict, "。")
			var textrand []string
			for _, value := range depict {
				textrand = append(textrand, strings.Split(value, "，")...)
			}
			return textrand[rand.Intn(len(textrand))]
		}
	}
}
