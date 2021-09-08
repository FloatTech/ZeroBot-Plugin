// Package diana 虚拟偶像女团 A-SOUL 成员嘉然相关
package diana

import (
	fmt "fmt"
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_diana/data"
)

var engine *zero.Engine

func init() {
	engine = control.Register("diana", &control.Options{
		DisableOnDefault: false,
		Help: "嘉然\n" +
			"- @BOT 小作文\n" +
			"- @BOT 发大病\n" +
			"- @BOT 教你一篇小作文[作文]\n" +
			"- [回复]查重",
	})
	// 随机发送一篇上面的小作文
	engine.OnFullMatch("小作文", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			rand.Seed(time.Now().UnixNano())
			// 绕过第一行发病
			ctx.Send((*data.Array)[rand.Intn(len(*data.Array)-1)+1])
		})
	// 逆天
	engine.OnFullMatch("发大病", zero.OnlyToMe).
		Handle(func(ctx *zero.Ctx) {
			// 第一行是发病
			ctx.Send((*data.Array)[0])
		})
	// 增加小作文
	engine.OnRegex(`^教你一篇小作文(.*)$`, zero.AdminPermission).
		Handle(func(ctx *zero.Ctx) {
			err := data.AddText(ctx.State["regex_matched"].([]string)[1])
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
			} else {
				ctx.Send("记住啦!")
			}
		})
}
