// Package hitokoto 一言
package hitokoto

import (
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
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
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
			if len(blist) == 0 {
				ctx.SendChain(message.Text("ERROR: hitokoto empty"))
				return
			}
			m := make(message.Message, 0, 10)
			text := strings.Builder{}
			if len(blist) <= 10 {
				for _, b := range blist {
					text.WriteString(b.Hitokoto)
					text.WriteString("\n——")
					text.WriteString(b.From)
					m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(text.String())))
					text.Reset()
				}
			} else {
				indexes := map[int]struct{}{}
				for i := 0; i < 10; i++ {
					ind := rand.Intn(len(blist))
					if _, ok := indexes[ind]; ok {
						i--
						continue
					}
					indexes[ind] = struct{}{}
				}
				for k := range indexes {
					b := blist[k]
					text.WriteString(b.Hitokoto)
					text.WriteString("\n——")
					text.WriteString(b.From)
					m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(text.String())))
					text.Reset()
				}
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
			tex := strings.Builder{}
			tex.WriteString("请输入系列一言序号\n")
			for i, v := range results {
				tex.WriteString(strconv.Itoa(i))
				tex.WriteString(". ")
				tex.WriteString(v.Category)
				tex.WriteString("\n")
			}
			base64Str, err := text.RenderToBase64(tex.String(), text.FontFile, 400, 20)
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
					if len(hlist) == 0 {
						ctx.SendChain(message.Text("ERROR: hitokoto empty"))
						return
					}
					m := make(message.Message, 0, 10)
					text := strings.Builder{}
					if len(hlist) <= 10 {
						for _, b := range hlist {
							text.WriteString(b.Hitokoto)
							text.WriteString("\n——")
							text.WriteString(b.From)
							m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(text.String())))
							text.Reset()
						}
					} else {
						indexes := map[int]struct{}{}
						for i := 0; i < 10; i++ {
							ind := rand.Intn(len(hlist))
							if _, ok := indexes[ind]; ok {
								i--
								continue
							}
							indexes[ind] = struct{}{}
						}
						for k := range indexes {
							b := hlist[k]
							text.WriteString(b.Hitokoto)
							text.WriteString("\n——")
							text.WriteString(b.From)
							m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text(text.String())))
							text.Reset()
						}
					}
					if id := ctx.Send(m).ID(); id == 0 {
						ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
					}
					return
				}
			}
		})
}
