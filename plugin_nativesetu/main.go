// Package nativesetu 本地setu
package nativesetu

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/rule"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	datapath = "data/nsetu"
	dbfile   = datapath + "/data.db"
	cfgfile  = datapath + "/setupath.txt"
)

var (
	setupath = "/tmp" // 绝对路径，图片根目录
)

func init() {
	engine := control.Register("nativesetu", order.PrioNativeSetu, &control.Options{
		DisableOnDefault: false,
		Help: "本地涩图\n" +
			"- 本地[xxx]\n" +
			"- 刷新本地[xxx]\n" +
			"- 设置本地setu绝对路径[xxx]\n" +
			"- 刷新所有本地setu\n" +
			"- 所有本地setu分类",
	})
	engine.OnRegex(`^本地(.*)$`, func(ctx *zero.Ctx) bool { return rule.FirstValueInList(setuclasses)(ctx) }).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			imgtype := ctx.State["regex_matched"].([]string)[1]
			sc := new(setuclass)
			mu.RLock()
			err := db.Pick(imgtype, sc)
			mu.RUnlock()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			} else {
				p := "file:///" + setupath + "/" + sc.Path
				ctx.SendChain(message.Text(imgtype, ": ", sc.Name, "\n"), message.Image(p))
			}
		})
	engine.OnRegex(`^刷新本地(.*)$`, func(ctx *zero.Ctx) bool { return rule.FirstValueInList(setuclasses)(ctx) }, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			imgtype := ctx.State["regex_matched"].([]string)[1]
			err := scanclass(os.DirFS(setupath), imgtype, imgtype)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
	engine.OnRegex(`^设置本地setu绝对路径(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			setupath = ctx.State["regex_matched"].([]string)[1]
			err := os.WriteFile(cfgfile, helper.StringToBytes(setupath), 0644)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
	engine.OnFullMatch("刷新所有本地setu", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := scanall(setupath)
			if err == nil {
				ctx.SendChain(message.Text("成功！"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
	engine.OnFullMatch("所有本地setu分类").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			msg := "所有本地setu分类"
			mu.RLock()
			for i, c := range setuclasses {
				n, err := db.Count(c)
				if err == nil {
					msg += fmt.Sprintf("\n%02d. %s(%d)", i, c, n)
				} else {
					msg += fmt.Sprintf("\n%02d. %s(error)", i, c)
					logrus.Errorln("[nsetu]", err)
				}
			}
			mu.RUnlock()
			ctx.SendChain(message.Text(msg))
		})
}
