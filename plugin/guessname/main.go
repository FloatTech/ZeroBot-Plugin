// Package guessname 音 游 开 字 符
package guessname

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	zbmath "github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	songdata string
	engine   = control.Register("guessname", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "音游开字符",
		Help:             "",
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

func main() {
	engine.OnRegex(`开字符([0-9]{1,2})`, getdata).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			n := zbmath.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if n == 0 || n > 10 {
				n = 10
			}
			songlist := strings.Split(songdata, "\n")
			rand.Shuffle(len(songlist), func(i, j int) {
				songlist[i], songlist[j] = songlist[j], songlist[i]
			})
			songlist = songlist[:n]
			sb := &strings.Builder{}
			sb2 := &strings.Builder{}
			for i, v := range songlist {
				sb.WriteString(strconv.Itoa(i))
				sb.WriteString(". ")
				sb.WriteString(v)
				sb.WriteString("\n")

				sb2.WriteString(strconv.Itoa(i))
				sb2.WriteString(". ")
				sb2.WriteString(strings.Replace(v, v, "*", -1))
				sb2.WriteString("\n")
			}
			answer := sb.String()
			disaplay := sb2.String()
		})
}
