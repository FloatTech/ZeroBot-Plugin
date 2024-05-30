package control

import (
	"encoding/binary"
	"strings"
	"time"

	b14 "github.com/fumiama/go-base16384"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"

	binutils "github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/process"
)

var startTime int64

func init() {
	// 插件冲突检测 会在本群发送一条消息并在约 1s 后撤回
	zero.OnFullMatch("插件冲突检测", zero.OnlyGroup, zero.AdminPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			tok := genToken()
			if tok == "" || len([]rune(tok)) != 4 {
				return
			}
			t := message.Text("●cd" + tok)
			startTime = time.Now().Unix()
			id := ctx.SendChain(t)
			process.SleepAbout1sTo2s()
			ctx.DeleteMessage(id)
		})

	zero.OnRegex("^●cd([\u4e00-\u8e00]{4})$", zero.OnlyGroup).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			if isValidToken(ctx.State["regex_matched"].([]string)[1], 10) {
				gid := ctx.Event.GroupID
				w := binutils.SelectWriter()
				managers.ForEach(func(key string, manager *ctrl.Control[*zero.Ctx]) bool {
					if manager.IsEnabledIn(gid) {
						w.WriteString("\xfe\xff")
						w.WriteString(key)
					}
					return true
				})
				if w.Len() > 2 {
					my, cl := binutils.OpenWriterF(func(wr *binutils.Writer) {
						wr.WriteString("●cd●")
						wr.WriteString(b14.EncodeString(w.String()[2:]))
					})
					binutils.PutWriter(w)
					id := ctx.SendChain(message.Text(binutils.BytesToString(my)))
					cl()
					process.SleepAbout1sTo2s()
					ctx.DeleteMessage(id)
				}
			}
		})

	zero.OnRegex("^●cd●(([\u4e00-\u8e00]*[\u3d01-\u3d06]?))", zero.OnlyGroup).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			if time.Now().Unix()-startTime < 10 {
				gid := ctx.Event.GroupID
				for _, s := range strings.Split(b14.DecodeString(ctx.State["regex_matched"].([]string)[1]), "\xfe\xff") {
					managers.RLock()
					c, ok := managers.M[s]
					managers.RUnlock()
					if ok && c.IsEnabledIn(gid) {
						c.Disable(gid)
					}
				}
			}
		})
}

func genToken() (tok string) {
	timebytes, cl := binutils.OpenWriterF(func(w *binutils.Writer) {
		w.WriteUInt64(uint64(time.Now().Unix()))
	})
	tok = b14.EncodeString(binutils.BytesToString(timebytes[1:]))
	cl()
	return
}

func isValidToken(tok string, throttlesecond int64) (yes bool) {
	s := b14.DecodeString(tok)
	timebytes, cl := binutils.OpenWriterF(func(w *binutils.Writer) {
		_ = w.WriteByte(0)
		w.WriteString(s)
	})
	yes = math.Abs64(time.Now().Unix()-int64(binary.BigEndian.Uint64(timebytes))) < throttlesecond
	cl()
	return
}
