// Package ymgal 月幕galgame
package ymgal

import (
	"fmt"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"log"
	"strings"
)

var (
	engine = control.Register("ymgal", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "月幕galgame\n- 随机galCG\n- 随机gal表情包\n- galCG[xxx]\n- gal表情包[xxx]\n- 更新gal\n",
		PublicDataFolder: "Ymgal",
	})
	dbfile = engine.DataFolder() + "ymgal.db"
)

func init() {
	go func() {
		defer order.DoneOnExit()()
		_, _ = file.GetLazyData(dbfile, false, false)
		gdb = initialize(dbfile)
		log.Println("[ymgal]加载月幕gal数据库")
	}()
	engine.OnRegex("^随机gal(CG|表情包)$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			pictureType := ctx.State["regex_matched"].([]string)[1]
			var y Ymgal
			if pictureType == "表情包" {
				y = gdb.randomYmgal(emoticonType)
			} else {
				y = gdb.randomYmgal(cgType)
			}
			fmt.Println(y)
			sendYmgal(y, ctx)
		})
	engine.OnRegex("^gal(CG|表情包)([一-龥A-Za-z0-9]{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			pictureType := ctx.State["regex_matched"].([]string)[1]
			key := ctx.State["regex_matched"].([]string)[2]
			var y Ymgal
			if pictureType == "CG" {
				y = gdb.getYmgalByKey(cgType, key)
			} else {
				y = gdb.getYmgalByKey(emoticonType, key)
			}
			fmt.Println(y)
			sendYmgal(y, ctx)
		})
	engine.OnFullMatch("更新gal", zero.SuperUserPermission).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			updatePic()
			ctx.Send("ymgal数据库已更新")
		})
}

func sendYmgal(y Ymgal, ctx *zero.Ctx) {
	if y.PictureList == "" {
		ctx.SendChain(message.Text(zero.BotConfig.NickName[0] + "暂时没有这样的图呢"))
		return
	}
	var m message.Message
	m = append(m,
		message.CustomNode(
			ctx.Event.Sender.NickName,
			ctx.Event.UserID,
			y.Title,
		))
	if y.PictureDescription != "" {
		m = append(m,
			message.CustomNode(
				ctx.Event.Sender.NickName,
				ctx.Event.UserID,
				y.PictureDescription,
			))
	}
	for _, v := range strings.Split(y.PictureList, ",") {
		m = append(m,
			message.CustomNode(
				ctx.Event.Sender.NickName,
				ctx.Event.UserID,
				[]message.MessageSegment{
					message.Image(v),
				}),
		)
	}
	if id := ctx.SendGroupForwardMessage(
		ctx.Event.GroupID,
		m).Get("message_id").Int(); id == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控了"))
	}
}
