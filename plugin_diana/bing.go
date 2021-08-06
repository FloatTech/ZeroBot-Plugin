// Package diana 虚拟偶像女团 A-SOUL 成员嘉然相关
package diana

import (
	fmt "fmt"
	"math/rand"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_diana/data"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const yamai = "贝拉抽我🥵嘉然骑在我背上🥵晚晚踩我🥵乃琳坐在王座是用看垃圾的眼神看我🥵🥵珈乐踢我🥵🥵，把我眼睛蒙住然后五只脚一起踩我🥵还让我猜脚是谁的，猜错了给我劈眼一铁棍🥵​"

func init() {
	// 随机发送一篇上面的小作文
	zero.OnFullMatch("小作文", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			rand.Seed(time.Now().UnixNano())
			ctx.SendChain(message.Text(data.Array[rand.Intn(len(data.Array))]))
		})

	// 逆天
	zero.OnFullMatch("发大病", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(yamai)
		})
	// 增加小作文
	zero.OnRegex(`^教你一篇小作文(.*)$`, zero.AdminPermission).
		Handle(func(ctx *zero.Ctx) {
			err := data.AddText(ctx.State["regex_matched"].([]string)[1])
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
			}
		})
}
