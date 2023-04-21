// Package ygosem 基于ygosem的插件功能
package ygosem

import (
	"errors"
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

type punish struct {
	GroupID  int64 // 群ID
	LastTime int64 // 时间
	Value    int   // 惩罚值
}

var (
	carddatas = &carddb{
		db: &sql.Sqlite{},
	}
	types = map[string]func(*imgfactory.Factory) ([]byte, error){
		"黑边":  setBack,
		"反色":  setBlur,
		"马赛克": setMark,
		"旋转":  doublePicture,
		"切图":  cutPic,
	}
	engine = control.Register("guessygo", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "游戏王猜卡游戏",
		Help:              "-猜卡游戏\n-(黑边|反色|马赛克|旋转|切图)猜卡游戏\n-----------------------\n惩罚值:\n当惩罚值达到30将关闭猜卡游戏功能30分钟",
		PrivateDataFolder: "ygosem",
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
		err = carddatas.db.Create("cards", &gameCardInfo{})
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		err = carddatas.db.Create("punish", &punish{})
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
	engine.OnRegex("^(黑边|反色|马赛克|旋转|切图)?猜卡游戏$", zero.OnlyGroup, getdb, func(ctx *zero.Ctx) bool {
		subTime, ok := carddatas.checkGroup(ctx.Event.GroupID)
		if !ok {
			ctx.SendChain(message.Text("处于惩罚期间,", strconv.FormatFloat(30-subTime, 'f', 0, 64), "分钟后解除"))
			return false
		}
		return true
	}).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("正在准备题目,请稍等"))
		var (
			length     int
			semdata    gameCardInfo
			picFile    string
			err        error
			answerName string
		)
		semdata, picFile, err = getSemData()
		if err != nil {
			semdata, err = carddatas.pick()
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]", err))
				return
			}
			picFile = semdata.PicFile
		}
		answerName = semdata.Name
		length = math.Ceil(len([]rune(answerName)), 4)
		picFile = cachePath + picFile
		// 对卡图做处理
		pictrue, err := randPicture(picFile, ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]", err))
			return
		}
		// 进行猜卡环节
		ctx.SendChain(message.Text("请回答下图的卡名\n以“我猜xxx”格式回答\n(xxx需包含卡名1/4以上)\n或发“提示”得提示;“取消”结束游戏"), message.ImageBytes(pictrue))
		recv, cancel := zero.NewFutureEvent("message", 1, true, zero.RegexRule("^((我猜.+)|提示|取消)$"), zero.OnlyGroup, zero.CheckGroup(ctx.Event.GroupID)).Repeat()
		defer cancel()
		tick := time.NewTimer(105 * time.Second)
		over := time.NewTimer(120 * time.Second)
		var (
			worry       = 1 //错误次数
			tickCount   = 0 // 提示次数
			answerCount = 0 // 问答次数
		)
		for {
			select {
			case <-tick.C:
				ctx.SendChain(message.Text("还有15s作答时间"))
			case <-over.C:
				err := carddatas.loadpunish(ctx.Event.GroupID, worry)
				if err == nil {
					err = errors.New("惩罚值+" + strconv.Itoa(worry))
				}
				msgID := ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("时间超时,游戏结束\n卡名是:\n", answerName, "\n"),
					message.Image("file:///"+file.BOTPATH+"/"+picFile),
					message.Text("\n", err)))
				if msgID.ID() == 0 {
					ctx.SendChain(message.Text("图片发送失败,可能被风控\n答案是:", answerName))
				}
				return
			case c := <-recv:
				msgID := c.Event.MessageID
				answer := c.Event.Message.String()
				_, after, ok := strings.Cut(answer, "我猜")
				if ok {
					if len([]rune(after)) < length {
						ctx.Send(message.ReplyWithMessage(msgID, message.Text("请输入", length, "字以上")))
						continue
					}
					answer = after
				}
				switch {
				case answer == "取消":
					if c.Event.UserID != ctx.Event.UserID {
						ctx.Send(message.ReplyWithMessage(msgID, message.Text("你无权限取消")))
						continue
					}
					tick.Stop()
					over.Stop()
					worry = math.Max(worry, 6)
					err := carddatas.loadpunish(ctx.Event.GroupID, worry)
					if err == nil {
						err = errors.New("惩罚值+" + strconv.Itoa(worry))
					}
					msgID := ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("游戏已取消\n卡名是:\n", answerName, "\n"),
						message.Image("file:///"+file.BOTPATH+"/"+picFile),
						message.Text("\n", err)))
					if msgID.ID() == 0 {
						ctx.SendChain(message.Text("图片发送失败,可能被风控\n答案是:", answerName))
					}
					return
				case answer == "提示" && tickCount > 3:
					tick.Reset(105 * time.Second)
					over.Reset(120 * time.Second)
					ctx.Send(message.ReplyWithMessage(msgID, message.Text("已经没有提示了哦,加油啊")))
					continue
				case answer != "提示" && answerCount > 5:
					tick.Stop()
					over.Stop()
					err := carddatas.loadpunish(ctx.Event.GroupID, worry)
					if err == nil {
						err = errors.New("惩罚值+" + strconv.Itoa(worry))
					}
					msgID := ctx.Send(message.ReplyWithMessage(msgID,
						message.Text("次数到了,很遗憾没能猜出来\n卡名是:\n", answerName, "\n"),
						message.Image("file:///"+file.BOTPATH+"/"+picFile),
						message.Text("\n", err)))
					if msgID.ID() == 0 {
						ctx.SendChain(message.Text("图片发送失败,可能被风控\n答案是:", answerName))
					}
					return
				case answer == "提示":
					worry++
					tickCount++
					tick.Reset(105 * time.Second)
					over.Reset(120 * time.Second)
					tips := getTips(semdata, tickCount)
					ctx.Send(message.ReplyWithMessage(msgID, message.Text(tips)))
				case strings.Contains(answerName, answer):
					tick.Stop()
					over.Stop()
					err := carddatas.loadpunish(ctx.Event.GroupID, worry)
					if err == nil {
						err = errors.New("惩罚值+" + strconv.Itoa(worry))
					}
					msgID := ctx.Send(message.ReplyWithMessage(msgID,
						message.Text("太棒了,你猜对了!\n卡名是:\n", answerName, "\n"),
						message.Image("file:///"+file.BOTPATH+"/"+picFile),
						message.Text("\n", err)))
					if msgID.ID() == 0 {
						ctx.SendChain(message.Text("图片发送失败,可能被风控\n答案是:", answerName))
					}
					return
				default:
					worry++
					answerCount++
					tick.Reset(105 * time.Second)
					over.Reset(120 * time.Second)
					ctx.Send(message.ReplyWithMessage(msgID, message.Text("答案不对哦,还有"+strconv.Itoa(6-answerCount)+"次回答机会,加油啊~")))
				}
			}
		}
	})
}

// 随机选择
func randPicture(picFile string, mode string) ([]byte, error) {
	pic, err := gg.LoadImage(picFile)
	if err != nil {
		return nil, err
	}
	dst := imgfactory.Size(pic, 256*5, 256*5)
	dstfunc, ok := types[mode]
	if !ok {
		var modes []string
		for mane := range types {
			modes = append(modes, mane)
		}
		dstfunc = types[modes[rand.Intn(len(modes))]]
	}
	return dstfunc(dst)
}

// 获取黑边
func setBack(dst *imgfactory.Factory) ([]byte, error) {
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
	markSize := 40 + 16*(rand.Intn(3))

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
func cutPic(dst *imgfactory.Factory) ([]byte, error) {
	indexOfx := rand.Intn(3)
	indexOfy := rand.Intn(3)
	indexOfx2 := rand.Intn(3)
	indexOfy2 := rand.Intn(3)
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

func (sql *carddb) insert(dbInfo gameCardInfo) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("cards", &gameCardInfo{})
	if err == nil {
		return sql.db.Insert("cards", &dbInfo)
	}
	return err
}

func (sql *carddb) pick() (dbInfo gameCardInfo, err error) {
	sql.RLock()
	defer sql.RUnlock()
	err = sql.db.Create("cards", &gameCardInfo{})
	if err == nil {
		err = sql.db.Pick("cards", &dbInfo)
	}
	return
}

func (sql *carddb) loadpunish(gid int64, i int) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("punish", &punish{})
	if err == nil {
		var groupInfo punish
		_ = sql.db.Find("punish", &groupInfo, "where GroupID = "+strconv.FormatInt(gid, 10))
		groupInfo.GroupID = gid
		groupInfo.LastTime = time.Now().Unix()
		groupInfo.Value += i
		return sql.db.Insert("punish", &groupInfo)
	}
	return err
}

func (sql *carddb) checkGroup(gid int64) (float64, bool) {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("punish", &punish{})
	if err != nil {
		return 0, true
	}
	var groupInfo punish
	_ = sql.db.Find("punish", &groupInfo, "where GroupID = "+strconv.FormatInt(gid, 10))
	if groupInfo.LastTime > 0 {
		subTime := time.Since(time.Unix(groupInfo.LastTime, 0)).Minutes()
		if subTime >= 30 {
			groupInfo.LastTime = time.Now().Unix()
			groupInfo.Value = 0
			_ = sql.db.Insert("punish", &groupInfo)
			return subTime, true
		} else if groupInfo.Value >= 30 {
			return subTime, false
		}
	}
	return 0, true
}
