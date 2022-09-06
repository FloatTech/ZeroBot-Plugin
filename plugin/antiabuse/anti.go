// Package antiabuse defines antiabuse plugin ,support abuse words check and add/remove abuse words
package antiabuse

import (
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
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
		managers = ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).Manager
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
		return true
	})

	engine.OnMessage(onceRule, zero.OnlyGroup, func(ctx *zero.Ctx) bool {
		if !ctx.Event.IsToMe {
			return true
		}
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		grp := strconv.FormatInt(gid, 36)
		msg := strings.ReplaceAll(ctx.MessageString(), "\n", "")
		msg = strings.ReplaceAll(msg, "\r", "")
		msg = strings.ReplaceAll(msg, "\t", "")
		msg = strings.ReplaceAll(msg, ";", "")
		mu.RLock()
		defer mu.RUnlock()
		if db.CanFind(grp, "WHERE instr('"+msg+"', word)>=0") {
			if err := managers.DoBlock(uid); err == nil {
				cache.Set(uid, struct{}{})
				ctx.SetGroupBan(gid, uid, 4*3600)
				ctx.SendChain(message.Text("检测到违禁词, 已封禁/屏蔽4小时"))
			} else {
				ctx.SendChain(message.Text("block user error: ", err))
			}
			return false
		}
		return true
	})

	engine.OnCommand("添加违禁词", zero.OnlyGroup, zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if err := insertWord(ctx.Event.GroupID, args); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			} else {
				ctx.SendChain(message.Text("成功"))
			}
		})

	engine.OnCommand("删除违禁词", zero.OnlyGroup, zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if err := deleteWord(ctx.Event.GroupID, args); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			} else {
				ctx.SendChain(message.Text("成功"))
			}
		})

	engine.OnCommand("查看违禁词", zero.OnlyGroup, onceRule).Handle(
		func(ctx *zero.Ctx) {
			b, err := text.RenderToBase64(listWords(ctx.Event.GroupID), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("本群违禁词有\n"), message.Image("base64://"+binary.BytesToString(b)))
		})
}
