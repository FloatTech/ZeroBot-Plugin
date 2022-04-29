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
)

type joke struct {
	ID   uint32 `db:"id"`
	Text string `db:"text"`
}

var db = &sql.Sqlite{}

func init() {
	en := control.Register("funny", &control.Options{
		DisableOnDefault: false,
		Help: "讲个笑话\n" +
			"- 讲个笑话[@xxx] | 讲个笑话[qq号]",
		PublicDataFolder: "Funny",
	})

	en.OnPrefix("讲个笑话", ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dbpath := en.DataFolder()
		db.DBPath = dbpath + "jokes.db"
		_, err := file.GetLazyData(db.DBPath, false, true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		err = db.Create("jokes", &joke{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		c, err := db.Count("jokes")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		logrus.Infoln("[funny]加载", c, "个笑话")
		return true
	})).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		// 获取名字
		name := ctx.NickName()
		var j joke
		err := db.Pick("jokes", &j)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(strings.ReplaceAll(j.Text, "%name", name)))
	})
}
