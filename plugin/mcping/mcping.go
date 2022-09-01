package mcping

import (
	"github.com/alteamc/minequery/ping"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	engine := control.Register("mcping", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help:             "-",
	})

	//engine.OnRegex(`^mcping\s(.+)`).SetBlock(true).
	//	Handle(func(ctx *zero.Ctx) {
	//		addr := ctx.State["regex_matched"].([]string)
	//		ctx.SendChain(message.Text(ping.Ping(addr[1], 25565)))
	//		ctx.SendChain(message.Text(addr))
	//	})

	engine.OnFullMatch("mcping").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			res, err := ping.Ping("cn-cd-dx-5.natfrp.cloud", 24010)
			if err != nil {
				return
			}
			ctx.SendChain(
				message.Text("Current players:\n", res.Players),
			)
		})
}
