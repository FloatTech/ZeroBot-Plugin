package anti_abuse

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
	engine := control.Register("anti_abuse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Help:              "违禁词检测",
		PrivateDataFolder: "anti_abuse",
	})
	onceRule := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = engine.DataFolder() + "anti_abuse.db"
		_, _ = engine.GetLazyData("anti_abuse.db", true)
		err := db.Open(time.Hour * 4)
		if err != nil {
			ctx.SendChain(message.Text("open db error: ", err))
			return false
		}
		err = db.Create("BanTime", &BanTime{})
		if err != nil {
			ctx.SendChain(message.Text("create table error: ", err))
			return false
		}
		err = db.Create("BanWord", &BanWord{})
		if err != nil {
			ctx.SendChain(message.Text("create table error: ", err))
			return false
		}
		err = recoverUser()
		if err != nil {
			ctx.SendChain(message.Text("recover data error: ", err))
			return false
		}
		err = recoverWord()
		if err != nil {
			ctx.SendChain(message.Text("recover data error: ", err))
			return false
		}
		return true
	})
	engine.OnMessage(onceRule, banRule)
	engine.OnCommand("添加违禁词", zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			if err := insertWord(ctx.Event.GroupID, ctx.State["args"].(string)); err != nil {
				ctx.SendChain(message.Text("add ban word error:", err))
			}
		})
	engine.OnCommand("删除违禁词", zero.AdminPermission, onceRule).Handle(
		func(ctx *zero.Ctx) {
			if err := deleteWord(ctx.Event.GroupID, ctx.State["args"].(string)); err != nil {
				ctx.SendChain(message.Text("add ban word error:", err))
			}
		})
	engine.OnCommand("查看违禁词", onceRule).Handle(
		func(ctx *zero.Ctx) {
			gidPrefix := fmt.Sprintf("%d-", ctx.Event.GroupID)
			var words []string
			_ = wordSet.Iter(func(s string) error {
				trueWord := strings.SplitN(s, gidPrefix, 1)[1]
				words = append(words, trueWord)
				return nil
			})
			ctx.SendChain(message.Text("本群违禁词有:\n", strings.Join(words, " |")))
		})

}
