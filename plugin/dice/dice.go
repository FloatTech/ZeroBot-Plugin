package dice

import (
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/manager/timer"
)

var (
	engine = control.Register("dice", &control.Options{
		DisableOnDefault: true,
		Help:             "使用.help来查看帮助",
		PublicDataFolder: "Dice",
	})
)

func init() {
	go func() {
		db.DBPath = engine.DataFolder() + "dice.db"
		clock = timer.NewClock(db)
		err := db.Create("strjrrp", &strjrrp{})
		if err != nil {
			panic(err)
		}
		err = db.Create("rsl", &rsl{})
		if err != nil {
			panic(err)
		}
		err = db.Create("set", &set{})
		if err != nil {
			panic(err)
		}
	}()
	engine.OnFullMatchGroup([]string{".help", "。help"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			help := "试图移植的dice\n-.jrrp\n-.ra\n-.rd"
			ctx.SendChain(message.Text(help))
		})
}
