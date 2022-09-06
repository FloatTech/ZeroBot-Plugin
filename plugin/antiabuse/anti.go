// Package antiabuse defines antiabuse plugin ,support abuse words check and add/remove abuse words
package antiabuse

import (
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/ttl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	managers *ctrl.Manager[*zero.Ctx] // managers lazy load
	cache    = ttl.NewCacheOn(4*time.Hour, [4]func(int64, struct{}){nil, nil, onDel, nil})
	db       = &antidb{}
)

func onDel(uid int64, _ struct{}) {
	if managers == nil {
		return
	}
	if err := managers.DoUnblock(uid); err != nil {
		logrus.Errorln("[antiabuse] do unblock:", err)
	}
}

func init() {
	engine := control.Register("antiabuse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Help:              "违禁词检测\n- /[添加|删除|查看]违禁词",
		PrivateDataFolder: "anti_abuse",
	})

	onceRule := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		managers = ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).Manager
		db.DBPath = engine.DataFolder() + "anti_abuse.db"
		err := db.Open(time.Hour * 4)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
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
		msg := strings.ReplaceAll(ctx.MessageString(), "\n", "")
		msg = strings.ReplaceAll(msg, "\r", "")
		msg = strings.ReplaceAll(msg, "\t", "")
		msg = strings.ReplaceAll(msg, ";", "")
		if db.isInAntiList(uid, gid, msg) {
			if err := ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).Manager.DoBlock(uid); err == nil {
				cache.Set(uid, struct{}{})
				ctx.SetGroupBan(gid, uid, 4*3600)
				ctx.SendChain(message.Text("检测到违禁词, 已封禁/屏蔽4小时"))
			} else {
				ctx.SendChain(message.Text("ERROR: block user: ", err))
			}
			return false
		}
		return true
	})

	engine.OnCommand("添加违禁词", zero.OnlyGroup, zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if err := db.insertWord(ctx.Event.GroupID, args); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			} else {
				ctx.SendChain(message.Text("成功"))
			}
		})

	engine.OnCommand("删除违禁词", zero.OnlyGroup, zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if err := db.deleteWord(ctx.Event.GroupID, args); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			} else {
				ctx.SendChain(message.Text("成功"))
			}
		})

	engine.OnCommand("查看违禁词", zero.OnlyGroup, onceRule).Handle(
		func(ctx *zero.Ctx) {
			b, err := text.RenderToBase64(db.listWords(ctx.Event.GroupID), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("本群违禁词有\n"), message.Image("base64://"+binary.BytesToString(b)))
		})
}
