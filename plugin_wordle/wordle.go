// Package wordle 猜单词
package wordle

import (
	"errors"
	"image/color"
	"math/rand"
	"sort"
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
	"github.com/wdvxdr1123/ZeroBot/extension/single"
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

var words []string

func init() {
	en := control.Register("wordle", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "猜单词\n" +
			"- 个人猜单词" +
			"- 团队猜单词",
		PublicDataFolder: "Wordle",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) interface{} { return ctx.Event.GroupID }),
		single.WithPostFn(func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(
					ctx.Event.MessageID,
					message.Text("已经有正在进行的游戏..."),
				),
			)
		}),
	))
	go func() {
		data, err := file.GetLazyData(en.DataFolder()+"words.bin", true, true)
		if err != nil {
			panic(err)
		}
		var wordpacks = loadwords(data)
		words = make([]string, 0, len(wordpacks))
		for i := range wordpacks {
			words = append(words, wordpacks[i].String())
		}
		sort.Strings(words)
	}()
	en.OnRegex(`(个人|团队)猜单词`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			game := newWordleGame()
			_, img, _ := game("")
			typ := ctx.State["regex_matched"].([]string)[1]
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
				message.Image("base64://"+binary.BytesToString(img)),
				message.Text("你有6次机会猜出单词，单词长度为5，请发送单词")))
			var next *zero.FutureEvent
			if typ == "个人" {
				next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^[A-Z]{5}$|^[a-z]{5}$`), zero.OnlyGroup, zero.CheckUser(ctx.Event.UserID))
			} else {
				next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(`^[A-Z]{5}$|^[a-z]{5}$`), zero.OnlyGroup)
			}
			recv, cancel := next.Repeat()
			defer cancel()
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID,
						message.Image("base64://"+binary.BytesToString(img)),
						message.Text("猜单词超时，游戏结束...")))
					return
				case e := <-recv:
					win, img, err := game(e.Message.String())
					switch {
					case win:
						ctx.Send(message.ReplyWithMessage(e.MessageID,
							message.Image("base64://"+binary.BytesToString(img)),
							message.Text("太棒了，你猜出来了！")))
						return
					case err == errTimesRunOut:
						ctx.Send(message.ReplyWithMessage(e.MessageID,
							message.Image("base64://"+binary.BytesToString(img)),
							message.Text("游戏结束...")))
						return
					case err == errLengthNotEnough:
						ctx.Send(message.ReplyWithMessage(e.MessageID,
							message.Image("base64://"+binary.BytesToString(img)),
							message.Text("单词长度错误")))
					case err == errUnknownWord:
						ctx.Send(message.ReplyWithMessage(e.MessageID,
							message.Image("base64://"+binary.BytesToString(img)),
							message.Text("你确定存在这样的单词吗？")))
					default:
						ctx.Send(message.ReplyWithMessage(e.MessageID,
							message.Image("base64://"+binary.BytesToString(img))))
					}
				}
			}
		})
}

func newWordleGame() func(string) (bool, []byte, error) {
	onhand := words[rand.Intn(len(words))]
	record := make([]string, 0, len(onhand)+1)
	return func(s string) (win bool, base64Image []byte, err error) {
		if s != "" {
			s = strings.ToLower(s)
			if onhand == s {
				win = true
			} else {
				if len(s) != len(onhand) {
					err = errLengthNotEnough
					return
				}
				i := sort.SearchStrings(words, s)
				if i >= len(words) || words[i] != s {
					err = errUnknownWord
					return
				}
			}
			record = append(record, s)
			if len(record) >= cap(record) {
				err = errTimesRunOut
			}
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
		base64Image, _ = writer.ToBase64(ctx.Image())
		return
	}
}
