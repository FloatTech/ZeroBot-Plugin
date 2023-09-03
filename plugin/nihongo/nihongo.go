// Package nihongo 日语学习
package nihongo

import (
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "日语学习",
		Help: "- 日语语法[xxx](使用tag随机)\n" +
			"搜索日语语法[xxx]",
		PublicDataFolder: "Nihongo",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = engine.DataFolder() + "nihongo.db"
		_, err := engine.GetLazyData("nihongo.db", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Create("grammar", &grammar{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		n, err := db.Count("grammar")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		log.Infof("[nihongo]读取%d条语法", n)
		return true
	})

	engine.OnRegex(`^日语语法\s?([0-9A-Za-zぁ-んァ-ヶ～]{1,6})$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			g := getRandomGrammarByTag(ctx.State["regex_matched"].([]string)[1])
			if g.ID == 0 {
				ctx.SendChain(message.Text("未能找到", ctx.State["regex_matched"].([]string)[1], "相关标签的语法"))
				return
			}
			data, err := text.RenderToBase64(g.string(), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	engine.OnRegex(`^搜索日语语法\s?([0-9A-Za-zぁ-んァ-ヶ～]{1,25})$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			g := getRandomGrammarByKeyword(ctx.State["regex_matched"].([]string)[1])
			if g.ID == 0 {
				ctx.SendChain(message.Text("未能找到", ctx.State["regex_matched"].([]string)[1], "相关标签的语法"))
				return
			}
			data, err := text.RenderToBase64(g.string(), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
