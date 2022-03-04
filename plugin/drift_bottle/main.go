// Package driftbottle 漂流瓶
package driftbottle

import (
	"strconv"
	"strings"
	"sync"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	en := control.Register("driftbottle", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              "漂流瓶\n- (在群xxx)丢漂流瓶(到频道xxx) [消息]\n- (从频道xxx)捡漂流瓶\n- @BOT 创建频道 xxx\n- 跳入(频道)海中\n- 注：不显式限制时，私聊发送可在所有群抽到，群聊发送仅可在本群抽到，默认频道为 global",
		PrivateDataFolder: "driftbottle",
	})
	sea.DBPath = en.DataFolder() + "sea.db"
	err := sea.Open()
	if err != nil {
		panic(err)
	}
	_ = createChannel(sea, "global")
	en.OnRegex(`^(在群\d+)?丢漂流瓶(到频道\w+)?\s+(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			msgs := ctx.State["regex_matched"].([]string)
			grp := ctx.Event.GroupID
			channel := "global"
			msg := msgs[3]
			var err error
			if msgs[1] != "" {
				grp, err = strconv.ParseInt(msgs[1][6:], 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("群号非法!"))
					return
				}
			}
			if msgs[2] != "" {
				channel = msgs[2][9:]
			}
			if msg == "" {
				ctx.SendChain(message.Text("消息为空!"))
				return
			}
			logrus.Debugln("[driftbottle]", grp, channel, msg)
			err = newBottle(
				ctx.Event.UserID,
				grp,
				ctxext.CardOrNickName(ctx, ctx.Event.UserID),
				msg,
			).throw(sea, channel)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("你将它扔进大海，希望有人捞到吧~")))
		})
	en.OnRegex(`^(从频道\w+)?捡漂流瓶$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			msgs := ctx.State["regex_matched"].([]string)
			grp := ctx.Event.GroupID
			if grp == 0 {
				grp = -ctx.Event.UserID
			}
			if grp == 0 {
				ctx.SendChain(message.Text("找不到对象!"))
				return
			}
			channel := "global"
			if msgs[1] != "" {
				channel = msgs[1][9:]
			}
			logrus.Debugln("[driftbottle]", grp, channel)
			b, err := fetchBottle(sea, channel, grp)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				err = b.destroy(sea, channel)
				wg.Done()
			}()
			ctx.Send(
				message.ReplyWithMessage(
					ctx.Event.MessageID,
					message.Text("你在海边捡到了一个来自 ", b.Name, " 的漂流瓶，打开瓶子，里面有一张纸条，写着："),
					message.Text(b.Msg),
				),
			)
			wg.Wait()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
		})
	en.OnPrefix("创建频道", zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			channel := strings.TrimRight(ctx.State["args"].(string), " ")
			if channel == "" {
				ctx.SendChain(message.Text("频道名为空!"))
				return
			}
			err := createChannel(sea, channel)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("成功~")))
		})
	en.OnRegex(`^跳入(\w+)?海中$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			msgs := ctx.State["regex_matched"].([]string)
			channel := "global"
			if msgs[1] != "" {
				channel = msgs[1]
			}
			seamu.RLock()
			c, err := sea.Count(channel)
			seamu.RUnlock()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("你缓缓走入大海，感受着海浪轻柔地拍打着你的小腿，膝盖……\n波浪卷着你的腰腹，你感觉有些把握不住平衡了……\n……\n你沉入海中，", c, " 个物体与你一同沉浮。\n不知何处涌来一股暗流，你失去了意识。")))
		})
}
