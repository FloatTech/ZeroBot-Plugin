// Package antiabuse defines antiabuse plugin ,support abuse words check and add/remove abuse words
package antiabuse

import (
	"fmt"
	"strings"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("antiabuse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Help:              "违禁词检测",
		PrivateDataFolder: "anti_abuse",
	})
	onceRule := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		managers = ctx.State["managers"].(*ctrl.Control[*zero.Ctx]).Manager
		db.DBPath = engine.DataFolder() + "anti_abuse.db"
		err := db.Open(time.Hour * 4)
		if err != nil {
			ctx.SendChain(message.Text("open db error: ", err))
			return false
		}
		err = db.Create("banWord", &banWord{})
		if err != nil {
			ctx.SendChain(message.Text("create table error: ", err))
			return false
		}
		err = recoverWord()
		if err != nil {
			ctx.SendChain(message.Text("recover data error: ", err))
			return false
		}
		return true
	})
	engine.OnMessage(onceRule, zero.OnlyGroup, banRule)
	engine.OnCommand("添加违禁词", zero.OnlyGroup, zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if err := insertWord(ctx.Event.GroupID, args); err != nil {
				ctx.SendChain(message.Text("error:", err))
			} else {
				ctx.SendChain(message.Text(fmt.Sprintf("添加违禁词 %s 成功", args)))
			}
		})
	engine.OnCommand("删除违禁词", zero.OnlyGroup, zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if err := deleteWord(ctx.Event.GroupID, args); err != nil {
				ctx.SendChain(message.Text("error:", err))
			} else {
				ctx.SendChain(message.Text(fmt.Sprintf("删除违禁词 %s 成功", args)))
			}
		})
	engine.OnCommand("查看违禁词", zero.OnlyGroup, onceRule).Handle(
		func(ctx *zero.Ctx) {
			if set, ok := wordMap[ctx.Event.GroupID]; !ok {
				ctx.SendChain(message.Text("本群无违禁词"))
			} else {
				ctx.SendChain(message.Text("本群违禁词有:", strings.Join(set.ToSlice(), " |")))
			}
		})
}
