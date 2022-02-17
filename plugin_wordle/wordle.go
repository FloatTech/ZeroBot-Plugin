// Package wordle 猜单词
package wordle

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/fogleman/gg"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var WORD_LEN = 5

var ErrLengthNotEnough = errors.New("length not enough")
var ErrUnknownWord = errors.New("unknown word")
var ErrTimesRunOut = errors.New("times run out")

const (
	match = iota
	exist
	notexist
	undone
)

var colormap = map[int]color.RGBA{
	0: {125, 166, 108, 255},
	1: {199, 183, 96, 255},
	2: {123, 123, 123, 255},
	3: {219, 219, 219, 255},
}

func init() {
	sort.Strings(words)
}

func init() {
	control.Register("wordle", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "猜单词\n" +
			"- 开始猜单词",
	}).OnFullMatch("猜单词").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			game := wordle(words)
			_, img, _ := game("")
			ctx.SendChain(message.ImageBytes(img), message.Text("请发送单词"))
			// 没有图片就索取
			next := zero.NewFutureEvent("message", 999, false,
				zero.RegexRule(fmt.Sprintf(`^[A-Z]{%d}$|^[a-z]{%d}$`, WORD_LEN, WORD_LEN)))
			recv, cancel := next.Repeat()
			defer cancel()
			for {
				select {
				case <-time.After(time.Second * 120):
					return
				case e := <-recv:
					win, img, err := game(e.Message.String())
					if err == ErrLengthNotEnough {
						ctx.SendChain(message.ImageBytes(img), message.Text("单词长度错误"))
					}
					if err == ErrUnknownWord {
						ctx.SendChain(message.ImageBytes(img), message.Text("不存在这样的单词"))
					}
					if win {
						ctx.SendChain(message.ImageBytes(img), message.Text("你赢了"))
						return
					}
					if err == ErrTimesRunOut {
						ctx.SendChain(message.ImageBytes(img), message.Text("你输了"))
						return
					}
					ctx.SendChain(message.ImageBytes(img))
				}
			}
		})
}

func wordle(words []string) func(string) (bool, []byte, error) {
	rand.Seed(time.Now().UnixMilli())
	index := rand.Intn(len(words))
	onhand := words[index]
	record := make([]string, 0, len(onhand)+1)
	return func(s string) (win bool, image []byte, err error) {
		if s != "" {
			s = strings.ToLower(s)
			if onhand == s {
				win = true
			} else {
				if len(s) != len(onhand) {
					err = ErrLengthNotEnough
					return
				}
				i := sort.SearchStrings(words, s)
				if i >= len(words) || words[i] != s {
					err = ErrUnknownWord
					return
				}
			}
			if len(record) >= cap(record) {
				err = ErrTimesRunOut
			}
			record = append(record, s)
		}
		var side = 20
		ctx := gg.NewContext((side+2)*5+26, (side+2)*6+26)
		ctx.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Clear()
		for i := 0; i < len(onhand)+1; i++ {
			for j := 0; j < len(onhand); j++ {
				if len(record) > i {
					ctx.DrawRectangle(float64(10+j*(side+4)), float64(10+i*(side+4)), float64(side), float64(side))
					switch {
					case record[i][j] == onhand[j]:
						ctx.SetColor(colormap[match])
					case strings.IndexByte(onhand, record[i][j]) != -1:
						ctx.SetColor(colormap[exist])
					default:
						ctx.SetColor(colormap[notexist])
					}
					ctx.Fill()
					ctx.SetColor(color.RGBA{255, 255, 255, 255})
					ctx.DrawString(strings.ToUpper(string(record[i][j])), float64(10+j*(side+4)+7), float64(10+i*(side+4)+15))
				} else {
					ctx.DrawRectangle(float64(10+j*(side+4)+1), float64(10+i*(side+4)+1), float64(side-2), float64(side-2))
					ctx.SetLineWidth(1)
					ctx.SetColor(colormap[undone])
					ctx.Stroke()
				}
			}
		}
		buf := bytes.NewBuffer(make([]byte, 0))
		png.Encode(buf, ctx.Image())
		return win, buf.Bytes(), err
	}
}
