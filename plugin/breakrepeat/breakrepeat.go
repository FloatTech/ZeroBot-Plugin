// Package breakrepeat 打断复读
package breakrepeat

import (
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/RomiChan/syncx"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const throttle = 3 // 不可超过 9

var sm syncx.Map[int64, string]

func init() {
	engine := control.Register("breakrepeat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "打断复读\n- 打断" + strconv.Itoa(throttle) + "次以上复读\n",
	})
	engine.On(`message/group`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			raw := ctx.Event.RawMessage
			r, ok := sm.Load(gid)
			if !ok || len(r) <= 3 || r[3:] != raw {
				sm.Store(gid, "0: "+raw)
				return
			}
			c := int(r[0] - '0')
			if c < throttle {
				sm.Store(gid, strconv.Itoa(c+1)+": "+raw)
				return
			}
			sm.Delete(gid)
			ctx.Send(r)
		})
}
