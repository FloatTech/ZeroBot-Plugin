// Package xhstext 小红书文案
package xhstext

import (
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type xhstext struct {
	ID    uint32 `db:"id"`
	Text  string `db:"text"`
	Label string `db:"label"`
}

var db sql.Sqlite

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "小红书文案",
		Help:             "- 捧场\n- 有梗",
		PublicDataFolder: "Xhstext",
	})

	// 初始化数据库
	initDB := fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			db = sql.New(en.DataFolder() + "xhstext.db")
			_, err := en.GetLazyData("xhstext.db", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = db.Open(time.Hour)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = db.Create("all_texts", &xhstext{})
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			c, err := db.Count("all_texts")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			logrus.Infoln("[xhstext]加载", c, "条小红书文案")
			return true
		},
	)

	// 捧场命令
	en.OnFullMatch("捧场", initDB).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		var x xhstext
		err := db.Find("all_texts", &x, "WHERE label = '捧场' ORDER BY RANDOM() LIMIT 1")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(x.Text))
	})

	// 有梗命令
	en.OnFullMatch("有梗", initDB).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		var x xhstext
		err := db.Find("all_texts", &x, "WHERE label = '有梗' ORDER BY RANDOM() LIMIT 1")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(x.Text))
	})
}
