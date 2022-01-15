// Package wtf 鬼东西
package wtf

import (
	"fmt"
	"strconv"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/order"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// 限制调用频率
	limit = rate.NewManager(time.Minute*5, 5)
)

func init() {
	en := control.Register("wtf", order.PrioWtf, &control.Options{
		DisableOnDefault: false,
		Help:             "鬼东西\n- 鬼东西列表\n- 查询鬼东西[序号][@xxx]",
	})
	en.OnFullMatch("鬼东西列表").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			s := ""
			for i, w := range table {
				s += fmt.Sprintf("%02d. %s\n", i, w.name)
			}
			ctx.SendChain(message.Text(s))
		})
	en.OnRegex(`^查询鬼东西(\d*)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.SendChain(message.Text("请稍后重试0x0..."))
				return
			}
			// 调用接口
			i, err := strconv.Atoi(ctx.State["regex_matched"].([]string)[1])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			w := newWtf(i)
			if w == nil {
				ctx.SendChain(message.Text("没有这项内容！"))
				return
			}
			// 获取名字
			var name string
			var secondname string
			if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
				qq, _ := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
				secondname = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("nickname").Str
			}
			name = ctx.Event.Sender.NickName
			var text string
			if secondname != "" {
				text, err = w.predict(name, secondname)
			} else {
				text, err = w.predict(name)
			}
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// TODO: 可注入
			ctx.Send(text)
		})
}
