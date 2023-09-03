// Package ymgal 月幕galgame
package ymgal

import (
	"strings"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "月慕galgame相关",
		Help:             "- 随机galCG\n- 随机gal表情包\n- galCG[xxx]\n- gal表情包[xxx]\n- 更新gal",
		PublicDataFolder: "Ymgal",
	})
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dbfile := engine.DataFolder() + "ymgal.db"
		_, err := engine.GetLazyData("ymgal.db", false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		gdb, err = initialize(dbfile)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})
	engine.OnRegex("^随机gal(CG|表情包)$", getdb).Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			pictureType := ctx.State["regex_matched"].([]string)[1]
			var y ymgal
			if pictureType == "表情包" {
				y = gdb.randomYmgal(emoticonType)
			} else {
				y = gdb.randomYmgal(cgType)
			}
			sendYmgal(y, ctx)
		})
	engine.OnRegex("^gal(CG|表情包)([一-龥ぁ-んァ-ヶA-Za-z0-9]{1,25})$", getdb).Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			pictureType := ctx.State["regex_matched"].([]string)[1]
			key := ctx.State["regex_matched"].([]string)[2]
			var y ymgal
			if pictureType == "CG" {
				y = gdb.getYmgalByKey(cgType, key)
			} else {
				y = gdb.getYmgalByKey(emoticonType, key)
			}
			sendYmgal(y, ctx)
		})
	engine.OnFullMatch("更新gal", zero.SuperUserPermission, getdb).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			err := updatePic()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send("ymgal数据库已更新")
		})
}

func sendYmgal(y ymgal, ctx *zero.Ctx) {
	if y.PictureList == "" {
		ctx.SendChain(message.Text(zero.BotConfig.NickName[0] + "暂时没有这样的图呢"))
		return
	}
	m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text(y.Title))}
	if y.PictureDescription != "" {
		m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(y.PictureDescription)))
	}
	for _, v := range strings.Split(y.PictureList, ",") {
		m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(v)))
	}
	if id := ctx.Send(m).ID(); id == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控了"))
	}
}
