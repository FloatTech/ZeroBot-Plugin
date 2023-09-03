// Package breakrepeat 打断复读
package breakrepeat

import (
	"math/rand"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/RomiChan/syncx"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const throttle = 3 // 不可超过 9

var sm syncx.Map[int64, string]

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "打断复读",
		Help:             "- 打断" + strconv.Itoa(throttle) + "次以上复读\n",
	})
	engine.On("message/group", zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		return !zero.HasPicture(ctx)
	}).SetBlock(false).
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
			if len(r) > 5 {
				ru := []rune(r[3:])
				rand.Shuffle(len(ru), func(i, j int) {
					ru[i], ru[j] = ru[j], ru[i]
				})
				r = string(ru)
			}
			ctx.Send(r)
		})
}
