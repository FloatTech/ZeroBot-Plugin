// Package breakrepeat 打断复读
package breakrepeat

import (
	"math/rand"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/RomiChan/syncx"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	maxLimit = 3
)

type result struct {
	Limit  int64
	RawMsg string
}

var (
	sm syncx.Map[int64, *result]
)

func init() {
	engine := control.Register("breakrepeat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "打断复读,打断3次以上复读\n",
	})
	engine.On(`message/group`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			raw := ctx.Event.RawMessage
			if r, ok := sm.Load(gid); !ok || r.RawMsg != raw {
				sm.Store(gid, &result{
					Limit:  0,
					RawMsg: raw,
				})
				return
			}
			if r, ok := sm.Load(gid); ok {
				sm.Store(gid, &result{
					Limit:  r.Limit + 1,
					RawMsg: raw,
				})
			}
			if res, ok := sm.Load(gid); ok && res.Limit >= maxLimit {
				r := []rune(res.RawMsg)
				if len(r) > 2 {
					rand.Shuffle(len(r), func(i, j int) {
						r[i], r[j] = r[j], r[i]
					})
					ctx.Send(string(r))
				}
				sm.Store(gid, &result{
					Limit:  0,
					RawMsg: res.RawMsg,
				})
			}
		})
}
