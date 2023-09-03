// Package antiabuse defines antiabuse plugin ,support abuse words check and add/remove abuse words
package antiabuse

import (
	"strconv"
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

const bandur time.Duration = time.Minute * 10

var (
	managers *ctrl.Manager[*zero.Ctx] // managers lazy load
	cache    = ttl.NewCacheOn(bandur, [4]func(int64, struct{}){nil, nil, onDel, nil})
	db       *antidb
)

func onDel(uid int64, _ struct{}) {
	if managers == nil {
		return
	}
	if err := managers.DoUnblock(uid); err != nil {
		logrus.Errorln("[antiabuse.onDel] unblock:", err)
	}
	if err := db.Del("__bantime__", "WHERE id="+strconv.FormatInt(uid, 10)); err != nil {
		logrus.Errorln("[antiabuse.onDel] db:", err)
	}
}

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "违禁词检测",
		Help:              "- /[添加|删除|查看]违禁词",
		PrivateDataFolder: "anti_abuse",
	})

	onceRule := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		managers = ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).Manager
		var err error
		db, err = newantidb(engine.DataFolder() + "anti_abuse.db")
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
		if db.isInAntiList(gid, msg) {
			if err := ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).Manager.DoBlock(uid); err == nil {
				t := time.Now().Unix()
				cache.Set(uid, struct{}{})
				ctx.SetThisGroupBan(uid, int64(bandur.Minutes()))
				ctx.SendChain(message.Text("检测到违禁词, 已封禁/屏蔽", bandur))
				db.Lock()
				defer db.Unlock()
				err := db.Create("__bantime__", nilbt)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return false
				}
				err = db.Insert("__bantime__", &banTime{ID: uid, Time: t})
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return false
				}
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
