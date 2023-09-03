// Package dress 女装
package dress

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
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
		Brief:            "女装",
		Help: "女装\n" +
			"- 女装\n" +
			"- 男装\n" +
			"- 随机女装\n" +
			"- 随机男装",
		PrivateDataFolder: "dress",
	})
	engine.OnFullMatchGroup([]string{"女装", "男装"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			matched := ctx.State["matched"].(string)
			sex := male
			if matched == "男装" {
				sex = female
			}
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			nameList, err := dressList(sex)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			tex := "请输入" + matched + "序号\n"
			for i, v := range nameList {
				tex += fmt.Sprintf("%d. %s\n", i, v)
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
					ctx.SendChain(message.Text(matched, "指令过期"))
					return
				case c := <-recv:
					msg := c.Event.Message.ExtractPlainText()
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字!"))
						continue
					}
					if num < 0 || num >= len(nameList) {
						ctx.SendChain(message.Text("序号非法!"))
						continue
					}
					name := nameList[num]
					sendImage(ctx, sex, matched, name)
					return
				}
			}
		})
	engine.OnFullMatchGroup([]string{"随机女装", "随机男装"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			matched := strings.TrimPrefix(ctx.State["matched"].(string), "随机")
			sex := male
			if matched == "男装" {
				sex = female
			}
			nameList, err := dressList(sex)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			name := nameList[rand.Intn(len(nameList))]
			sendImage(ctx, sex, matched, name)
		})
}

func sendImage(ctx *zero.Ctx, sex, matched, name string) {
	ctx.SendChain(message.Text("请欣赏", matched, ": ", name))
	count, err := detail(sex, name)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	imageList := make([]string, count)
	for i := range imageList {
		imageList[i] = fmt.Sprintf(dressImageURL, sex, name, i+1)
	}
	m := message.Message{}
	for _, v := range imageList {
		m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(v)))
	}
	if id := ctx.Send(m).ID(); id == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
	}
}
