// Package hitokoto 一言
package hitokoto

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.Register("hitokoto", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "一言",
		Help: "- 一言[xxx]\n" +
			"- 系列一言",
		PublicDataFolder: "Hitokoto",
	})
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dbfile := engine.DataFolder() + "hitokoto.db"
		_, err := engine.GetLazyData("hitokoto.db", false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		hdb, err = initialize(dbfile)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})
	engine.OnPrefix(`一言`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			args := ctx.State["args"].(string)
			blist, err := hdb.getByKey(strings.TrimSpace(args))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			textList := make([]string, 0, 10)
			for _, v := range blist {
				textList = append(textList, v.Hitokoto+"\n——"+v.From)
			}
			rand.Shuffle(len(textList), func(i, j int) {
				textList[i], textList[j] = textList[j], textList[i]
			})
			m := message.Message{}
			for _, v := range textList[:10] {
				m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(v)))
			}
			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
		})
	engine.OnFullMatch(`系列一言`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			results, err := hdb.getAllCategory()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			tex := "请输入系列一言序号\n"
			for i, v := range results {
				tex += fmt.Sprintf("%d. %s\n", i, v.Category)
			}
			base64Str, err := text.RenderToBase64(tex, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(base64Str)))
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.SendChain(message.Text("系列一言指令过期"))
					return
				case c := <-recv:
					msg := c.Event.Message.ExtractPlainText()
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字!"))
						continue
					}
					if num < 0 || num >= len(results) {
						ctx.SendChain(message.Text("序号非法!"))
						continue
					}
					ctx.SendChain(message.Text("请欣赏系列一言: ", results[num].Category))
					hlist, err := hdb.getByCategory(results[num].Category)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					textList := make([]string, 0, 10)
					for _, v := range hlist {
						textList = append(textList, v.Hitokoto+"\n——"+v.From)
					}
					rand.Shuffle(len(textList), func(i, j int) {
						textList[i], textList[j] = textList[j], textList[i]
					})
					m := message.Message{}
					for _, v := range textList[:10] {
						m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(v)))
					}
					if id := ctx.Send(m).ID(); id == 0 {
						ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
					}
					return
				}
			}
		})
}
