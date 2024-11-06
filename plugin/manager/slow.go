package manager

import (
	"time"

	"github.com/RomiChan/syncx"
	"github.com/fumiama/slowdo"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var slowsenders = syncx.Map[int64, *syncx.Lazy[*slowdo.Job[*zero.Ctx, message.Segment]]]{}

func collectsend(ctx *zero.Ctx, msgs ...message.Segment) {
	id := ctx.Event.GroupID
	if id == 0 {
		// only support group
		return
	}
	lazy, _ := slowsenders.LoadOrStore(id, &syncx.Lazy[*slowdo.Job[*zero.Ctx, message.Segment]]{
		Init: func() *slowdo.Job[*zero.Ctx, message.Segment] {
			x, err := slowdo.NewJob(time.Second*5, ctx, func(ctx *zero.Ctx, msg []message.Segment) {
				if len(msg) == 1 {
					ctx.Send(msg)
					return
				}
				m := make(message.Message, len(msg))
				for i, item := range msg {
					m[i] = message.CustomNode(
						zero.BotConfig.NickName[0],
						ctx.Event.SelfID,
						message.Message{item})
				}
				ctx.SendGroupForwardMessage(id, m)
			})
			if err != nil {
				panic(err)
			}
			return x
		},
	})
	job := lazy.Get()
	for _, msg := range msgs {
		job.Add(msg)
	}
}
