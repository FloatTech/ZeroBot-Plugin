// Package tiangou 舔狗日记
package tiangou

import (
	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type tiangou struct {
	ID   uint32 `db:"id"`
	Text string `db:"text"`
}

var db = &sql.Sqlite{}

func init() {
	en := control.Register("tiangou", &control.Options{
		DisableOnDefault: false,
		Help: "舔狗日记\n" +
			"- 舔狗日记",
		PublicDataFolder: "Tiangou",
	})

	en.OnFullMatch("舔狗日记", ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			db.DBPath = en.DataFolder() + "tiangou.db"
			_, err := en.GetLazyData("tiangou.db", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = db.Create("tiangou", &tiangou{})
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			c, err := db.Count("tiangou")
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			logrus.Infoln("[tiangou]加载", c, "条舔狗日记")
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		var t tiangou
		err := db.Pick("tiangou", &t)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(t.Text))
	})
}
