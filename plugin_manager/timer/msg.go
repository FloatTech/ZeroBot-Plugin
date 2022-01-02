package timer

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func (t *Timer) sendmsg(grp int64, ctx *zero.Ctx) {
	ctx.Event = new(zero.Event)
	ctx.Event.GroupID = grp
	if t.URL == "" {
		ctx.SendChain(atall, message.Text(t.Alert))
	} else {
		ctx.SendChain(atall, message.Text(t.Alert), message.Image(t.URL).Add("cache", "0"))
	}
}
