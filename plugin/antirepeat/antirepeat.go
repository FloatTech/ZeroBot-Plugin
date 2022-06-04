// Package antirepeat 限制复读
package antirepeat

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	limit  = make(map[int64]int)
	rawMsg = make(map[int64]string)
)

func init() {
	en := control.Register("antirepeat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help:             "限制复读",
		PublicDataFolder: "Antirepeat",
	})
	// 只接收群聊消息
	en.On(`message/group`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			// 定义常用变量
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			raw := ctx.Event.RawMessage
			// 检查是否不是复读
			if rawMsg[gid] != raw {
				// 重置rawMsg
				rawMsg[gid] = raw
				// 重置limit
				limit[gid] = 0
				return
			}
			limit[gid]++
			// 检查是否到达limit
			if limit[gid] >= 3 {
				ctx.SetGroupBan(gid, uid, 60*60*3)
			}
		})
}
