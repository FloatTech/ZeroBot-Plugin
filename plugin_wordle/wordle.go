// Package wordle 猜单词
package wordle

import (
	"errors"
	"image/color"
	"math/rand"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/fogleman/gg"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	errLengthNotEnough = errors.New("length not enough")
	errUnknownWord     = errors.New("unknown word")
	errTimesRunOut     = errors.New("times run out")
)

const (
	match = iota
	exist
	notexist
	undone
)

var colors = [...]color.RGBA{
	{125, 166, 108, 255},
	{199, 183, 96, 255},
	{123, 123, 123, 255},
	{219, 219, 219, 255},
}

var words []wordpack

func init() {
	en := control.Register("wordle", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "猜单词\n" +
			"- 开始猜单词",
		PublicDataFolder: "Wordle",
	})
	go func() {
		data, err := file.GetLazyData(en.DataFolder()+"words.bin", true, true)
		if err != nil {
			panic(err)
		}
		words = loadwords(data)
	}()
	en.OnFullMatch("猜单词").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			game := newWordleGame()
			_, img, _ := game("")
			ctx.SendChain(message.Image("base64://"+binary.BytesToString(img)), message.Text("请发送单词"))
			// 没有图片就索取
			next := zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^[A-Z]{5}$|^[a-z]{5}$`))
			recv, cancel := next.Repeat()
			var msg message.Message
			defer func() {
				cancel()
				ctx.Send(msg)
			}()
			for {
				select {
				case <-time.After(time.Second * 120):
					return
				case e := <-recv:
					win, img, err := game(e.Message.String())
					msg = []message.MessageSegment{message.Image("base64://" + binary.BytesToString(img))}
					switch err {
					case nil:
						if win {
							msg = append(msg, message.Text("你赢了"))
							return
						}
					case errLengthNotEnough:
						msg = append(msg, message.Text("单词长度错误"))
					case errUnknownWord:
						msg = append(msg, message.Text("不存在这样的单词"))
					case errTimesRunOut:
						msg = append(msg, message.Text("你输了"))
						return
					}
					ctx.Send(msg)
				}
			}
		})
}

func newWordleGame() func(string) (bool, []byte, error) {
	onhandpack := words[rand.Intn(len(words))]
	onhand := onhandpack.String()
	record := make([]string, 0, len(onhand)+1)
	return func(s string) (win bool, base64Image []byte, err error) {
		if s != "" {
			s = strings.ToLower(s)
			sp := pack(s)
			if onhandpack == sp {
				win = true
			} else {
				if len(s) != len(onhand) {
					err = errLengthNotEnough
					return
				}
				i := 0
				for ; i < len(words); i++ {
					if words[i] == sp {
						break
					}
				}
				if i >= len(words) || words[i] != sp {
					err = errUnknownWord
					return
				}
			}
			if len(record) >= cap(record) {
				err = errTimesRunOut
				return
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
						ctx.SetColor(colors[match])
					case strings.IndexByte(onhand, record[i][j]) != -1:
						ctx.SetColor(colors[exist])
					default:
						ctx.SetColor(colors[notexist])
					}
					ctx.Fill()
					ctx.SetColor(color.RGBA{255, 255, 255, 255})
					ctx.DrawString(strings.ToUpper(string(record[i][j])), float64(10+j*(side+4)+7), float64(10+i*(side+4)+15))
				} else {
					ctx.DrawRectangle(float64(10+j*(side+4)+1), float64(10+i*(side+4)+1), float64(side-2), float64(side-2))
					ctx.SetLineWidth(1)
					ctx.SetColor(colors[undone])
					ctx.Stroke()
				}
			}
		}
		base64Image, err = writer.ToBase64(ctx.Image())
		return
	}
}
