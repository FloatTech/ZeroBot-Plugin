// Package wordle 猜单词
package wordle

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FloatTech/AnimeAPI/tl"
	"github.com/FloatTech/imgfactory"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/gg"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
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

var classdict = map[string]int{
	"":   5,
	"五阶": 5,
	"六阶": 6,
	"七阶": 7,
}

type dictionary map[int]struct {
	dict []string
	cet4 []string
}

var words = make(dictionary)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "猜单词",
		Help: "- 个人猜单词\n" +
			"- 团队猜单词\n" +
			"- 团队六阶猜单词\n" +
			"- 团队七阶猜单词",
		PublicDataFolder: "Wordle",
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

	en.OnRegex(`^(个人|团队)(五阶|六阶|七阶)?猜单词$`, zero.OnlyGroup, fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			var errcnt uint32
			var wg sync.WaitGroup
			var mu sync.Mutex
			for i := 5; i <= 7; i++ {
				wg.Add(2)
				go func(i int) {
					defer wg.Done()
					dc, err := en.GetLazyData(fmt.Sprintf("cet-4_%d.txt", i), true)
					if err != nil {
						atomic.AddUint32(&errcnt, 1)
						return
					}
					c := strings.Split(binary.BytesToString(dc), "\n")
					sort.Strings(c)
					mu.Lock()
					tmp := words[i]
					tmp.cet4 = c
					words[i] = tmp
					mu.Unlock()
				}(i)
				go func(i int) {
					defer wg.Done()
					dd, err := en.GetLazyData(fmt.Sprintf("dict_%d.txt", i), true)
					if err != nil {
						atomic.AddUint32(&errcnt, 1)
						return
					}
					d := strings.Split(binary.BytesToString(dd), "\n")
					sort.Strings(d)
					mu.Lock()
					tmp := words[i]
					tmp.dict = d
					words[i] = tmp
					mu.Unlock()
				}(i)
			}
			wg.Wait()
			if errcnt > 0 {
				ctx.SendChain(message.Text("ERROR: 下载字典时发生", errcnt, "个错误"))
				return false
			}
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			class := classdict[ctx.State["regex_matched"].([]string)[2]]
			target := words[class].cet4[rand.Intn(len(words[class].cet4))]
			tt, err := tl.Translate(target)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			game := newWordleGame(target)
			_, img, _ := game("")
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.ImageBytes(img),
					message.Text("你有", class+1, "次机会猜出单词，单词长度为", class, "，请发送单词"),
				),
			)
			var next *zero.FutureEvent
			if ctx.State["regex_matched"].([]string)[1] == "个人" {
				next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(fmt.Sprintf(`^([A-Z]|[a-z]){%d}$`, class)),
					zero.OnlyGroup, ctx.CheckSession())
			} else {
				next = zero.NewFutureEvent("message", 999, false, zero.RegexRule(fmt.Sprintf(`^([A-Z]|[a-z]){%d}$`, class)),
					zero.OnlyGroup, zero.CheckGroup(ctx.Event.GroupID))
			}
			var win bool
			recv, cancel := next.Repeat()
			defer cancel()
			tick := time.NewTimer(105 * time.Second)
			after := time.NewTimer(120 * time.Second)
			for {
				select {
				case <-tick.C:
					ctx.SendChain(message.Text("猜单词，你还有15s作答时间"))
				case <-after.C:
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("猜单词超时，游戏结束...答案是: ", target, "(", tt, ")"),
						),
					)
					return
				case c := <-recv:
					tick.Reset(105 * time.Second)
					after.Reset(120 * time.Second)
					win, img, err = game(c.Event.Message.String())
					switch {
					case win:
						tick.Stop()
						after.Stop()
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.ImageBytes(img),
								message.Text("太棒了，你猜出来了！答案是: ", target, "(", tt, ")"),
							),
						)
						return
					case err == errTimesRunOut:
						tick.Stop()
						after.Stop()
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.ImageBytes(img),
								message.Text("游戏结束...答案是: ", target, "(", tt, ")"),
							),
						)
						return
					case err == errLengthNotEnough:
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("单词长度错误"),
							),
						)
					case err == errUnknownWord:
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.Text("你确定存在这样的单词吗？"),
							),
						)
					default:
						ctx.Send(
							message.ReplyWithMessage(c.Event.MessageID,
								message.ImageBytes(img),
							),
						)
					}
				}
			}
		})
}

func newWordleGame(target string) func(string) (bool, []byte, error) {
	var class = len(target)
	record := make([]string, 0, len(target)+1)
	return func(s string) (win bool, data []byte, err error) {
		if s != "" {
			s = strings.ToLower(s)
			if target == s {
				win = true
			} else {
				if len(s) != len(target) {
					err = errLengthNotEnough
					return
				}
				i := sort.SearchStrings(words[class].dict, s)
				if i >= len(words[class].dict) || words[class].dict[i] != s {
					err = errUnknownWord
					return
				}
			}
			record = append(record, s)
			if len(record) >= cap(record) {
				err = errTimesRunOut
				return
			}
		}
		var side = 20
		var space = 10
		ctx := gg.NewContext((side+4)*class+space*2-4, (side+4)*(class+1)+space*2-4)
		ctx.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Clear()
		for i := 0; i < class+1; i++ {
			for j := 0; j < class; j++ {
				if len(record) > i {
					ctx.DrawRectangle(float64(space+j*(side+4)), float64(space+i*(side+4)), float64(side), float64(side))
					switch {
					case record[i][j] == target[j]:
						ctx.SetColor(colors[match])
					case strings.IndexByte(target, record[i][j]) != -1:
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
		data, err = imgfactory.ToBytes(ctx.Image())
		return
	}
}
