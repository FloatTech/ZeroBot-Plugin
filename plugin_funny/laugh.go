// Package funny 冷笑话
package funny

import (
	"strings"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	sql "github.com/FloatTech/sqlite"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"

	"github.com/FloatTech/zbputils/control/order"
)

type joke struct {
	ID   uint32 `db:"id"`
	Text string `db:"text"`
}

var db = &sql.Sqlite{}

func init() {
	en := control.Register("funny", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "讲个笑话\n" +
			"- 讲个笑话[@xxx] | 讲个笑话[qq号]",
		PublicDataFolder: "Funny",
	})

	go func() {
		dbpath := en.DataFolder()
		db.DBPath = dbpath + "jokes.db"
		defer order.DoneOnExit()()
		_, err := file.GetLazyData(db.DBPath, false, true)
		if err != nil {
			panic(err)
		}
		err = db.Create("jokes", &joke{})
		if err != nil {
			panic(err)
		}
		c, _ := db.Count("jokes")
		logrus.Infoln("[funny]加载", c, "个笑话")
	}()

	en.OnPrefix("讲个笑话").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		// 获取名字
		name := ctxext.NickName(ctx)
		var j joke
		err := db.Pick("jokes", &j)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(strings.ReplaceAll(j.Text, "%name", name)))
	})
}
