package manager

import (
	"time"

	"github.com/FloatTech/zbputils/ctxext"
	"github.com/RomiChan/syncx"
	"github.com/fumiama/slowdo"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var slowsenders = syncx.Map[int64, *syncx.Lazy[*slowdo.Job[message.MessageSegment, *zero.Ctx]]]{}

func collectsend(ctx *zero.Ctx, msgs ...message.MessageSegment) {
	id := ctx.Event.GroupID
	if id == 0 {
		// only support group
		return
	}
	lazy, _ := slowsenders.LoadOrStore(id, &syncx.Lazy[*slowdo.Job[message.MessageSegment, *zero.Ctx]]{
		Init: func() *slowdo.Job[message.MessageSegment, *zero.Ctx] {
			x, err := slowdo.NewJob(time.Second*5, ctx, func(ctx *zero.Ctx, msg []message.MessageSegment) {
				m := make(message.Message, len(msg))
				for i, item := range msg {
					m[i] = ctxext.FakeSenderForwardNode(ctx, item)
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
