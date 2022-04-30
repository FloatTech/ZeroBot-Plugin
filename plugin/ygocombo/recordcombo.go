// Package sleepmanage 睡眠管理
package recordcombo

import (
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
)

func init() { // 插件主体
	engine := control.Register("recordcombo", &control.Options{
		DisableOnDefault: false,
		Help: "combo记录器\n" +
			"- combo列表\n" +
			"- 回复要记录的combo内容对话“记录combo”\n" +
			"- 查看combo [xxx]\n" +
			"- 删除combo [xxx]  (仅管理员可用)\n",
		PublicDataFolder: "YgoCombo",
	})
	go func() {
		sdb = initialize(engine.DataFolder() + "combo.db")
	}()
	engine.OnMessage(func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		if msg[0].Type != "reply" {
			return false
		}
		for _, elem := range msg {
			if elem.Type == "text" {
				text := elem.Data["text"]
				text = strings.ReplaceAll(text, " ", "")
				text = strings.ReplaceAll(text, "\r", "")
				text = strings.ReplaceAll(text, "\n", "")
				if text == "记录combo" {
					return true
				}
			}
		}
		return false
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		combocontent := ctx.GetMessage(message.NewMessageIDFromString(ctx.Event.Message[0].Data["id"])).Elements[0].Data["text"] //combo内容
		if combocontent == "" {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("你是想记录「空手假象」combo吗？"),
				),
			)
			return
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("请输入combo名称")))
		//等待用户下一步选择
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID)).Repeat()
		for {
			select {
			case <-time.After(time.Second * 120): //两分钟等待
				cancel()
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("等待超时,记录失败"),
					),
				)
				return
			case e := <-recv:
				comboName := e.Event.Message.String() //获取下一个指令
				username := ctx.CardOrNickName(ctx.Event.UserID)
				err := sdb.addmanage(comboName, username, ctx.Event.UserID, combocontent)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				ctx.SendChain(message.Text("成功添加“", comboName, "”combo\n内容：\n", combocontent))
				return
			}
		}
	})

	engine.OnRegex(`^删除combo (.+)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			comboName := ctx.State["regex_matched"].([]string)[1]
			err := sdb.removemanage(comboName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Text(comboName, "删除成功"))
		})
	//
	engine.OnFullMatchGroup([]string{"combo列表"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			state, err := sdb.managelist()
			if err != nil {
				ctx.SendChain(message.Text(err))
				return
			}
			ctx.SendChain(message.Text(strings.Join(state, "")))
		})

	engine.OnPrefix("查看combo", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		comboName := ctx.State["args"].(string)
		state, err := sdb.lookupmanage(comboName)
		if err != nil {
			ctx.SendChain(message.Text(err))
			return
		}
		ctx.SendChain(message.Text(strings.Join(state, "")))
	})
}
