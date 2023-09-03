// Package diana 虚拟偶像女团 A-SOUL 成员嘉然相关
package diana

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/diana/data"
)

var engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
	DisableOnDefault: false,
	Brief:            "嘉然相关", // 也许使用常用功能当Brief更好
	Help: "- 小作文\n" +
		"- 发大病\n" +
		"- 教你一篇小作文[作文]",
	PublicDataFolder: "Diana",
})

func init() {
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		err := data.LoadText(engine.DataFolder() + "text.db")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})

	// 随机发送一篇上面的小作文
	engine.OnFullMatch("小作文", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 绕过第一行发病
			ctx.SendChain(message.Text(data.RandText()))
		})
	// 逆天
	engine.OnFullMatch("发大病", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 第一行是发病
			ctx.SendChain(message.Text(data.HentaiText()))
		})
	// 增加小作文
	engine.OnRegex(`^教你一篇小作文(.*)$`, zero.AdminPermission, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := data.AddText(ctx.State["regex_matched"].([]string)[1])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			} else {
				ctx.SendChain(message.Text("记住啦!"))
			}
		})
}
