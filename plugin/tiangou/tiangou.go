// Package tiangou 舔狗日记
package tiangou

import (
	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
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
	en := control.Register("tiangou", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "舔狗日记\n" +
			"- 舔狗日记",
		PublicDataFolder: "Tiangou",
	})

	go func() {
		dbpath := en.DataFolder()
		db.DBPath = dbpath + "tiangou.db"
		_, err := file.GetLazyData(db.DBPath, false, true)
		if err != nil {
			panic(err)
		}
		err = db.Create("tiangou", &tiangou{})
		if err != nil {
			panic(err)
		}
		c, _ := db.Count("tiangou")
		logrus.Infoln("[tiangou]加载", c, "条舔狗日记")
	}()

	en.OnFullMatch("舔狗日记").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		var t tiangou
		err := db.Pick("tiangou", &t)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(t.Text))
	})
}
