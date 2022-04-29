package dice

import (
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	engine = control.Register("dice", &control.Options{
		DisableOnDefault: true,
		Help:             "使用.help来查看帮助",
		PublicDataFolder: "Dice",
	})
)

func init() {
	engine.OnFullMatchGroup([]string{".help", "。help"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			help := "试图移植的dice\n-.jrrp\n-.ra\n-.rd"
			ctx.SendChain(message.Text(help))
		})
}
