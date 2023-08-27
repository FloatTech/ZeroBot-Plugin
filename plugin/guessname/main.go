// Package guessname 音 游 开 字 符
package guessname

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	zbmath "github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	songdata string
	open     = regexp.MustCompile(`^开(.{1})`)
	nospace  = regexp.MustCompile(`([^\s])`)
	shoot    = regexp.MustCompile(`^盲狙(.*)`)

	engine = control.Register("guessname", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "音游开字符",
		Help: "开始游戏: \n" +
			"开字符[1-10]\n" +
			"游戏中: \n" +
			"开[任意字符]\n" +
			"盲狙[歌曲名]\n" +
			"tips: 忽略了大小写",
		PublicDataFolder: "Guessname",
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
	getdata = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		data, err := engine.GetLazyData("song.data", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: 数据下载失败"))
			return false
		}
		songdata = binary.BytesToString(data)
		return true
	})
)

func init() {
	engine.OnRegex(`^开字符([0-9]{1,2})?`, getdata).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			n := zbmath.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if n == 0 || n > 10 {
				n = 10
			}

			songlist := strings.Split(songdata, "\n")
			rand.Shuffle(len(songlist), func(i, j int) {
				songlist[i], songlist[j] = songlist[j], songlist[i]
			})

			answer := songlist[:n]

			display := make([]string, len(answer))
			for i, v := range answer {
				display[i] = nospace.ReplaceAllString(v, "*")
			}

			ctx.SendChain(message.Text(slice2string(display), "\n请发送你要开的字符"))
			next, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, ctx.CheckSession()).Repeat()
			defer cancel()

			tick := time.NewTimer(105 * time.Second)
			after := time.NewTimer(120 * time.Second)
			for {
				select {
				case <-tick.C:
					ctx.SendChain(message.Text("你还有15s作答时间"))
				case <-after.C:
					ctx.Send(
						message.ReplyWithMessage(ctx.Event.MessageID,
							message.Text("超时, 游戏结束...答案是: \n", slice2string(answer)),
						),
					)
					return
				case c := <-next:
					rawmsg := c.Event.Message.ExtractPlainText()

					if rawmsg == "结束" {
						ctx.SendChain(message.Text("已结束!\n", slice2string(answer)))
						return
					}

					if open.MatchString(rawmsg) {
						rawmsg = open.FindStringSubmatch(rawmsg)[1]
						display, ok := foreachwithreplace(rawmsg, answer, display)
						if !ok {
							ctx.SendChain(message.Text("什么都没开出来哦~"))
							continue
						}
						ctx.SendChain(message.Text("开出了以下字符: \n", slice2string(display)))
						continue
					}

					ok := false

					if shoot.MatchString(rawmsg) {
						rawmsg = shoot.FindStringSubmatch(rawmsg)[1]
						for i, v := range answer {
							if strings.EqualFold(display[i], rawmsg) {
								ctx.SendChain(message.Text("你已经猜过这首了!"))
								break
							}
							if !strings.EqualFold(v, rawmsg) {
								continue
							}
							ok = true
							display[i] = answer[i]
							ctx.SendChain(message.Text("你猜出来了一首, 继续! \n", slice2string(display)))
							break
						}
					}

					if !ok {
						ctx.SendChain(message.Text("你猜错了, 可惜!"))
					}
				}
			}
		})
}

func slice2string(s []string) string {
	sb := &strings.Builder{}
	for i, v := range s[:len(s)-1] {
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(".")
		sb.WriteString(v)
		sb.WriteString("\n")
	}
	sb.WriteString(strconv.Itoa(len(s)))
	sb.WriteString(".")
	sb.WriteString(s[len(s)-1])
	return sb.String()
}

func foreachwithreplace(str string, answer, display []string) ([]string, bool) {
	ok := false
	for i, v := range answer {
		if !strings.Contains(v, strings.ToLower(str)) && !strings.Contains(v, strings.ToUpper(str)) {
			continue
		}
		ok = true
		for j, s := range v {
			if !strings.EqualFold(string(s), str) {
				continue
			}
			if j == 0 {
				display[i] = string(s) + display[i][j+1:]
				continue
			}
			display[i] = display[i][:j-1] + string(s) + display[i][j:]
		}
	}

	return display, ok
}
