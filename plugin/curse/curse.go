// Package curse 骂人插件(求骂,自卫)
package curse

import (
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
)

const (
	minLevel = "min"
	maxLevel = "max"
)

func init() {
	engine := control.Register("curse", &control.Options{
		DisableOnDefault: true,
		Help:             "骂人(求骂,自卫)\n- 骂我\n- 大力骂我",
		PublicDataFolder: "Curse",
	})

	getdb := ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dbpath := engine.DataFolder()
		db.DBPath = dbpath + "curse.db"
		_, err := file.GetLazyData(db.DBPath, false, true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		err = db.Create("curse", &curse{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		c, err := db.Count("curse")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		logrus.Infoln("[curse]加载", c, "条骂人语录")
		return true
	})

	engine.OnFullMatch("骂我", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		text := getRandomCurseByLevel(minLevel).Text
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})

	engine.OnFullMatch("大力骂我", getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		text := getRandomCurseByLevel(maxLevel).Text
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})

	engine.OnKeywordGroup([]string{"他妈", "公交车", "你妈", "操", "屎", "去死", "快死", "我日", "逼", "尼玛", "艾滋", "癌症", "有病", "烦你", "你爹", "屮", "cnm"}, zero.OnlyToMe, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := getRandomCurseByLevel(maxLevel).Text
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
		})
}
