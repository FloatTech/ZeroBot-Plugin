// Package handou 猜成语
package handou

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/imgfactory"
	"github.com/sirupsen/logrus"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type idiomJson struct {
	Word         string   `json:"word"`         // 成语
	Chars        []string `json:"chars"`        // 成语
	Pinyin       []string `json:"pinyin"`       // 拼音
	Baobian      string   `json:"baobian"`      // 褒贬义
	Explanation  string   `json:"explanation"`  // 解释
	Derivation   string   `json:"derivation"`   // 词源
	Example      string   `json:"example"`      // 例句
	Abbreviation string   `json:"abbreviation"` // 结构
	Synonyms     []string `json:"synonyms"`     // 近义词
}

const (
	kong        = rune(' ')
	pinFontSize = 45.0
	hanFontSize = 150.0
)

const (
	match = iota
	exist
	notexist
	blockmatch
	blockexist
)

var colors = [...]color.RGBA{
	{0, 153, 0, 255},
	{255, 128, 0, 255},
	{123, 123, 123, 255},
	{125, 166, 108, 255},
	{199, 183, 96, 255},
}

var (
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "猜成语",
		Help: "- 个人猜成语\n" +
			"- 团队猜成语\n",
		PublicDataFolder: "Handou",
	}).ApplySingle(ctxext.NewGroupSingle("已经有正在进行的游戏..."))
	userHabitsFile = file.BOTPATH + "/" + en.DataFolder() + "userHabits.json"
	idiomFilePath  = file.BOTPATH + "/" + en.DataFolder() + "idiom.json"
	initialized    = fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			idiomFile, err := en.GetLazyData("idiom.json", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 下载字典时发生错误.\n", err))
				return false
			}
			err = json.Unmarshal(idiomFile, &idiomInfoMap)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 解析字典时发生错误.\n", err))
				return false
			}
			habitsIdiomKeys = make([]string, 0, len(idiomInfoMap))
			for k := range idiomInfoMap {
				habitsIdiomKeys = append(habitsIdiomKeys, k)
			}
			// 构建用户习惯库（全局高频N-gram）
			err = initUserHabits()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 构建用户习惯库时发生错误.\n", err))
				return false
			}
			// 下载字体
			data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 加载字体时发生错误.\n", err))
				return false
			}
			pinyinFont = data
			return true
		},
	)

	pinyinFont      []byte
	idiomInfoMap    = make(map[string]idiomJson)
	habitsIdiomKeys = make([]string, 0)

	errHadGuessed      = errors.New("had guessed")
	errLengthNotEnough = errors.New("length not enough")
	errUnknownWord     = errors.New("unknown word")
	errTimesRunOut     = errors.New("times run out")
)

func init() {
	en.OnRegex(`^猜成语热门(汉字|成语)$`, zero.OnlyGroup, initialized).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		if ctx.State["regex_matched"].([]string)[1] == "汉字" {
			topChars := getTopCharacters(10)
			ctx.SendChain(message.Text("热门汉字：\n", strings.Join(topChars, "\n")))
		} else {
			topIdioms := getTopIdioms(10)
			ctx.SendChain(message.Text("热门成语：\n", strings.Join(topIdioms, "\n")))
		}
	})
	en.OnRegex(`^(个人|团队)猜成语$`, zero.OnlyGroup, initialized).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		target := poolIdiom()
		idiomData := idiomInfoMap[target]
		game := newHandouGame(idiomData)
		_, img, _ := game("")
		anser := anserOutString(idiomData)
		worldLength := len(idiomData.Chars)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.ImageBytes(img),
				message.Text("你有", 7, "次机会猜出", worldLength, "字成语\n首字拼音为：", idiomData.Pinyin[0]),
			),
		)
		var next *zero.FutureEvent
		if ctx.State["regex_matched"].([]string)[1] == "个人" {
			next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(fmt.Sprintf(`^([\p{Han}，,]){%d}$`, worldLength)),
				zero.OnlyGroup, ctx.CheckSession())
		} else {
			next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(fmt.Sprintf(`^([\p{Han}，,]){%d}$`, worldLength)),
				zero.OnlyGroup, zero.CheckGroup(ctx.Event.GroupID))
		}
		var err error
		var win bool
		recv, cancel := next.Repeat()
		defer cancel()
		tick := time.NewTimer(105 * time.Second)
		after := time.NewTimer(120 * time.Second)
		for {
			select {
			case <-tick.C:
				ctx.SendChain(message.Text("猜成语，你还有15s作答时间"))
			case <-after.C:
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("猜成语超时，游戏结束...\n答案是: ", anser),
					),
				)
				return
			case c := <-recv:
				tick.Reset(105 * time.Second)
				after.Reset(120 * time.Second)
				err = updateHabits(c.Event.Message.String())
				if err != nil {
					logrus.Warn("更新用户习惯库时发生错误: ", err)
				}
				win, img, err = game(c.Event.Message.String())
				switch {
				case win:
					tick.Stop()
					after.Stop()
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.ImageBytes(img),
							message.Text("太棒了，你猜出来了！\n答案是: ", anser),
						),
					)
					return
				case err == errTimesRunOut:
					tick.Stop()
					after.Stop()
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.ImageBytes(img),
							message.Text("游戏结束...\n答案是: ", anser),
						),
					)
					return
				case err == errLengthNotEnough:
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.Text("成语长度错误"),
						),
					)
				case err == errHadGuessed:
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.Text("该成语已经猜过了"),
						),
					)
				case err == errUnknownWord:
					ctx.Send(
						message.ReplyWithMessage(c.Event.MessageID,
							message.Text("你确定存在这样的成语吗？"),
						),
					)
				default:
					if img != nil {
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.ImageBytes(img),
							),
						)
					} else {
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("回答错误。"),
							),
						)

					}
				}
			}
		}
	})
}

func poolIdiom() string {
	prioritizedData := prioritizeData(habitsIdiomKeys)
	if len(prioritizedData) > 0 {
		return prioritizedData[rand.Intn(len(prioritizedData))]
	}
	// 如果没有优先级数据，则随机选择一个成语
	keys := make([]string, 0, len(idiomInfoMap))
	for k := range idiomInfoMap {
		keys = append(keys, k)
	}
	return keys[rand.Intn(len(keys))]
}

func newHandouGame(target idiomJson) func(string) (bool, []byte, error) {
	var (
		class  = len(target.Chars)
		words  = target.Word
		chars  = target.Chars
		pinyin = target.Pinyin

		tickTruePinyin  = make([]string, class)
		tickExistChars  = make([]string, class)
		tickExistPinyin = make([]string, 0, class)

		record = make([]string, 0, 7)
	)
	// 初始化 tick, 第一个是已知的拼音
	for i := range class {
		if i == 0 {
			tickTruePinyin[i] = pinyin[0]
		} else {
			tickTruePinyin[i] = ""
		}
		tickExistChars[i] = "?"
	}

	return func(s string) (win bool, data []byte, err error) {
		answer := []rune(s)
		var answerData idiomJson

		if s != "" {
			if words == s {
				win = true
			}

			if len(answer) != len(chars) {
				err = errLengthNotEnough
				return
			}
			if slices.Contains(record, s) {
				err = errHadGuessed
				return
			}

			answerInfo, ok := idiomInfoMap[s]
			if !ok {
				newIdiom, err1 := geiAPIdata(s)
				if err1 != nil {
					logrus.Warningln("通过API获取成语信息时发生错误: ", err1)
					err = errUnknownWord
					return
				}
				logrus.Warningln("通过API获取成语信息: ", newIdiom.Word)
				if newIdiom.Word != "" {
					idiomInfoMap[newIdiom.Word] = *newIdiom
					go func() { _ = saveIdiomJson() }()
				}
				if newIdiom.Word != s {
					err = errUnknownWord
					return
				}
				answerData = *newIdiom
			} else {
				answerData = answerInfo
			}
			if len(record) >= 6 || win {
				// 结束了显示答案
				tickTruePinyin = target.Pinyin
				tickExistChars = target.Chars
			} else {
				// 处理汉字匹配逻辑
				for i := range class {
					char := answerData.Chars[i]
					if char == chars[i] {
						tickExistChars[i] = char
					} else {
						tickExistChars[i] = "?"
					}
				}

				// 确保 tickExistPinyin 有足够的长度
				if len(tickExistPinyin) < class {
					for i := len(tickExistPinyin); i < class; i++ {
						tickExistPinyin = append(tickExistPinyin, "")
					}
				}

				// 处理拼音匹配逻辑
				minPinyinLen := min(len(pinyin), len(answerData.Pinyin))
				for i := range minPinyinLen {
					pyChar := pinyin[i]
					answerPinyinChar := []rune(pyChar)
					tickTruePinyinChar := make([]rune, len(answerPinyinChar))
					tickExistPinyinChar := []rune(tickExistPinyin[i])

					if tickTruePinyin[i] != "" {
						copy(tickTruePinyinChar, []rune(tickTruePinyin[i]))
					} else {
						for k := range answerPinyinChar {
							tickTruePinyinChar[k] = kong
						}
					}

					PinyinChar := answerData.Pinyin[i]
					for j, c := range []rune(PinyinChar) {
						if c == kong {
							continue
						}
						switch {
						case j < len(answerPinyinChar) && c == answerPinyinChar[j]:
							tickTruePinyinChar[j] = c
						case slices.Contains(answerPinyinChar, c):
							// 如果字符存在但位置不对，添加到 tickExistPinyinChar
							if !slices.Contains(tickExistPinyinChar, c) {
								tickExistPinyinChar = append(tickExistPinyinChar, c)
							}
						default:
							if j < len(tickTruePinyinChar) {
								tickTruePinyinChar[j] = kong
							}
						}
					}

					// 处理提示逻辑，将非匹配位置设为下划线
					matchIndex := -1
					for j, v := range tickTruePinyinChar {
						if v != kong && v != '_' {
							matchIndex = j
						}
					}
					for j := range tickTruePinyinChar {
						if j > matchIndex {
							break
						}
						if tickTruePinyinChar[j] == kong {
							tickTruePinyinChar[j] = '_'
						}
					}
					// 更新提示拼音
					tickTruePinyin[i] = string(tickTruePinyinChar)
					tickExistPinyin[i] = string(tickExistPinyinChar)
				}
				if len(record) == 2 {
					tickTruePinyin[0] = pinyin[0]
					tickExistChars[0] = chars[0]
				}
			}
		}

		// 准备绘制数据
		existPinyin := make([]string, 0, class)
		for _, v := range tickExistPinyin {
			if v != "" {
				v = "?" + v
			}
			existPinyin = append(existPinyin, v)
		}
		tickIdiom := idiomJson{
			Chars:  tickExistChars,
			Pinyin: tickTruePinyin,
		}

		// 确保所有切片长度一致
		if len(tickIdiom.Chars) < class {
			// 如果答案字符数不足，用问号填充
			for i := len(tickIdiom.Chars); i < class; i++ {
				tickIdiom.Chars = append(tickIdiom.Chars, "?")
			}
		}
		if len(tickIdiom.Pinyin) < class {
			// 如果答案拼音数不足，用空字符串填充
			for i := len(tickIdiom.Pinyin); i < class; i++ {
				tickIdiom.Pinyin = append(tickIdiom.Pinyin, "")
			}
		}

		if s == "" {
			answerData = tickIdiom
		}

		var (
			tickImage   image.Image
			answerImage image.Image
			imgHistery  = make([]image.Image, 0)
			hisH        = 0
			wg          = &sync.WaitGroup{}
		)
		wg.Add(2)

		go func() {
			defer wg.Done()
			tickImage = drawHanBlock(hanFontSize/2, pinFontSize/2, tickIdiom, target, existPinyin...)
		}()
		go func() {
			defer wg.Done()
			answerImage = drawHanBlock(hanFontSize, pinFontSize, answerData, target)
		}()
		if len(record) > 0 {
			wg.Add(len(record))
			for i, v := range record {
				imgHistery = append(imgHistery, nil)
				go func(i int, v string) {
					defer wg.Done()
					idiom, ok := idiomInfoMap[v]
					if !ok {
						return
					}
					hisImage := drawHanBlock(hanFontSize/3, pinFontSize/3, idiom, target)
					imgHistery[i] = hisImage
					if i == 0 {
						hisH = hisImage.Bounds().Dy()
					}
				}(i, v)
			}
		}
		wg.Wait()

		// 记录猜过的成语
		if s != "" && !win {
			record = append(record, s)
		}

		if tickImage == nil || answerImage == nil {
			return
		}

		tickW, tickH := tickImage.Bounds().Dx(), tickImage.Bounds().Dy()
		answerW, answerH := answerImage.Bounds().Dx(), answerImage.Bounds().Dy()

		ctx := gg.NewContext(1, 1)
		_ = ctx.ParseFontFace(pinyinFont, pinFontSize/2)
		wordH, _ := ctx.MeasureString("M")

		ctxWidth := max(tickW, answerW)
		ctxHeight := tickH + answerH + int(wordH) + hisH*(len(imgHistery)+1)/2

		ctx = gg.NewContext(ctxWidth, ctxHeight)
		ctx.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Clear()

		ctx.SetColor(color.RGBA{0, 0, 0, 255})
		_ = ctx.ParseFontFace(pinyinFont, hanFontSize/2)
		ctx.DrawStringAnchored("题目:", float64(ctxWidth-tickW)/4, float64(tickH)/2, 0.5, 0.5)

		ctx.DrawImageAnchored(tickImage, ctxWidth/2, tickH/2, 0.5, 0.5)
		ctx.DrawImageAnchored(answerImage, ctxWidth/2, tickH+int(wordH)+answerH/2, 0.5, 0.5)

		k := 0
		for i, v := range imgHistery {
			if v == nil {
				continue
			}
			x := ctxWidth / 4
			y := tickH + int(wordH) + answerH + hisH*k

			if i%2 == 1 {
				x = ctxWidth * 3 / 4
				y = tickH + int(wordH) + answerH + hisH*k
				k++
			}
			ctx.DrawImageAnchored(v, x, y+hisH/2, 0.5, 0.5)
		}

		data, err = imgfactory.ToBytes(ctx.Image())
		if len(record) >= cap(record) {
			err = errTimesRunOut
			return
		}

		return
	}
}

// drawHanBlock 绘制汉字方块，支持多行显示（6字以上时分成两行）
func drawHanBlock(hanFontSize, pinFontSize float64, idiom, target idiomJson, exitPinyin ...string) image.Image {
	class := len(target.Chars)

	// 确保切片长度一致
	if len(idiom.Chars) < class {
		temp := make([]string, class)
		copy(temp, idiom.Chars)
		for i := len(idiom.Chars); i < class; i++ {
			temp[i] = "?"
		}
		idiom.Chars = temp
	}
	if len(idiom.Pinyin) < class {
		temp := make([]string, class)
		copy(temp, idiom.Pinyin)
		for i := len(idiom.Pinyin); i < class; i++ {
			temp[i] = ""
		}
		idiom.Pinyin = temp
	}

	chars := idiom.Chars
	pinyin := idiom.Pinyin

	// 确定行数和每行字数
	rows := 1
	charsPerRow := class
	if class > 6 {
		rows = 2
		charsPerRow = (class + 1) / 2
	}

	ctx := gg.NewContext(1, 1)
	_ = ctx.ParseFontFace(pinyinFont, pinFontSize)
	pinWidth, pinHeight := ctx.MeasureString("w")
	_ = ctx.ParseFontFace(pinyinFont, hanFontSize)
	hanWidth, hanHeight := ctx.MeasureString("拼")

	space := int(pinHeight / 2)
	blockPinWidth := int(pinWidth*6) + space
	boxPadding := math.Min(math.Abs(float64(blockPinWidth)-hanWidth)/2, hanHeight*0.3)

	// 计算总宽度和高度
	width := space + charsPerRow*blockPinWidth + space
	height := space + rows*(int(pinHeight+hanHeight+boxPadding*2)+space*2) + space
	if len(exitPinyin) > 0 {
		height = space + rows*(int(pinHeight+hanHeight+boxPadding*2+pinHeight)+space*2) + space
	}

	ctx = gg.NewContext(width, height)
	ctx.SetColor(color.RGBA{255, 255, 255, 255})
	ctx.Clear()

	for i := range class {
		// 边界检查
		if i >= len(chars) || i >= len(pinyin) || i >= len(target.Pinyin) || i >= len(target.Chars) {
			break
		}

		// 计算当前字符在哪一行哪一列
		idiom_rows := 0
		col := i
		if rows > 1 {
			idiom_rows = i / charsPerRow
			col = i % charsPerRow
		}

		x := float64(space + col*blockPinWidth)
		// 如果上一层字数是奇数就额外移位
		if idiom_rows%2 == 1 {
			x += float64(blockPinWidth) / 2
		}
		y := float64(idiom_rows*(int(pinHeight+hanHeight+boxPadding*2)+space*2) + space)
		if len(exitPinyin) > 0 {
			y = float64(idiom_rows*(int(pinHeight+hanHeight+boxPadding*2+pinHeight)+space*2) + space)
		}

		// 绘制拼音
		_ = ctx.ParseFontFace(pinyinFont, pinFontSize)
		if i < len(pinyin) {
			targetPinyinByte := []rune(target.Pinyin[i])
			pinyinByte := []rune(pinyin[i])

			// 取两者中的最大长度
			pinTotalWidth := pinWidth * float64(len(pinyinByte))
			pinX := x + float64(blockPinWidth)/2 - pinTotalWidth/2
			pinY := y + pinHeight/2

			for k, ch := range pinyinByte {
				ctx.SetColor(colors[notexist])
				for m, c := range targetPinyinByte {
					if k == m && ch == c {
						ctx.SetColor(colors[match])
						break
					} else if ch == c {
						ctx.SetColor(colors[exist])
					}
				}
				ctx.DrawStringAnchored(string(ch), pinX+pinWidth*float64(k)+pinWidth/2, pinY, 0.5, 0.5)
			}
		}

		// 绘制汉字方框
		boxX := x + boxPadding
		boxY := y + pinHeight + float64(space)
		boxWidth := float64(blockPinWidth) - boxPadding*2
		boxHeight := float64(hanHeight) + boxPadding*2
		ctx.DrawRectangle(boxX, boxY, boxWidth, boxHeight)

		// 设置方框颜色
		char := chars[i]
		switch {
		case char == target.Chars[i]:
			ctx.SetColor(colors[blockmatch])
		case char != "" && strings.Contains(target.Word, char):
			ctx.SetColor(colors[blockexist])
		default:
			ctx.SetColor(colors[notexist])
		}
		ctx.Fill()

		// 绘制汉字
		_ = ctx.ParseFontFace(pinyinFont, hanFontSize)
		ctx.SetColor(color.RGBA{255, 255, 255, 255})
		hanX := boxX + boxWidth/2
		hanY := boxY + boxHeight/2
		ctx.DrawStringAnchored(char, hanX, hanY, 0.5, 0.5)

		// 绘制题目的拼音提示
		ctx.SetColor(colors[exist])
		_ = ctx.ParseFontFace(pinyinFont, pinFontSize)
		if len(exitPinyin) > i && exitPinyin[i] != "" {
			tickY := boxY + boxHeight + float64(space) + pinHeight/2
			ctx.DrawStringAnchored(exitPinyin[i], hanX, tickY, 0.5, 0.5)
		}
	}
	return ctx.Image()
}

func anserOutString(s idiomJson) string {
	msg := s.Word
	if s.Baobian != "" && s.Baobian != "-" {
		msg += "\n" + s.Baobian + "词"
	}
	if s.Derivation != "" && s.Derivation != "-" {
		msg += "\n词源:\n" + s.Derivation
	} else {
		msg += "\n词源:无"
	}
	if s.Explanation != "" && s.Explanation != "-" {
		msg += "\n解释:\n" + s.Explanation
	} else {
		msg += "\n解释:无"
	}
	if len(s.Synonyms) > 0 {
		msg += "\n近义词:\n" + strings.Join(s.Synonyms, ",")
	}

	return msg
}
