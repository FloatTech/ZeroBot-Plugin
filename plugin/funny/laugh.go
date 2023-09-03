// Package funny 冷笑话
package funny

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

type joke struct {
	ID   uint32 `db:"id"`
	Text string `db:"text"`
}

var db = &sql.Sqlite{}

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "讲个笑话",
		Help:             "- 讲个笑话[@xxx|qq号|人名] | 夸夸[@xxx|qq号|人名] ",
		PublicDataFolder: "Funny",
	})

	en.OnPrefixGroup([]string{"讲个笑话", "夸夸"}, fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = en.DataFolder() + "jokes.db"
		_, err := en.GetLazyData("jokes.db", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Create("jokes", &joke{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		c, err := db.Count("jokes")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(strings.ReplaceAll(j.Text, "%name", name)))
	})
}
