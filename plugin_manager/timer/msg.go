package timer

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func (ts *Timer) sendmsg(grp int64, ctx *zero.Ctx) {
	ctx.Event = new(zero.Event)
	ctx.Event.GroupID = grp
	if ts.URL == "" {
		ctx.SendChain(atall, message.Text(ts.Alert))
	} else {
		ctx.SendChain(atall, message.Text(ts.Alert), message.Image(ts.URL).Add("cache", "0"))
	}
}
