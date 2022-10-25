package test

import (
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("tupian", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "全部图片指令\n" +
			"- 涩涩哒咩/我要涩涩\n",
	})
	engine.OnRegex(`^测试\s?(\S{1,25})\s?(\S{1,25})\s*(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		id := ctx.State["regex_matched"].(string)
		zzz := ctx.State["regex_matched"].([]string)[2]
		ccc := ctx.State["regex_matched"].([]string)[3]
		ctx.SendChain(message.Text(id))
		ctx.SendChain(message.Text(zzz))
		ctx.SendChain(message.Text(ccc))
	})
}
