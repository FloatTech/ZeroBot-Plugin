// Package diana 虚拟偶像女团 A-SOUL 成员嘉然相关
package diana

import (
	fmt "fmt"
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_diana/data"
)

func init() {
	// 随机发送一篇上面的小作文
	zero.OnFullMatch("小作文", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			rand.Seed(time.Now().UnixNano())
			// 绕过第一行发病
			ctx.SendChain(message.Text((*data.Array)[rand.Intn(len(*data.Array)-1)+1]))
		})
	// 逆天
	zero.OnFullMatch("发大病", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			// 第一行是发病
			ctx.Send((*data.Array)[0])
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
