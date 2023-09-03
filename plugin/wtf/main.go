// Package wtf 鬼东西
package wtf

import (
	"fmt"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "鬼东西",
		Help:             "- 鬼东西列表\n- 查询鬼东西[序号][@xxx]",
	})
	en.OnFullMatch("鬼东西列表").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			s := ""
			for i, w := range table {
				s += fmt.Sprintf("%02d. %s\n", i, w.name)
			}
			ctx.SendChain(message.Text(s))
		})
	en.OnRegex(`^查询鬼东西(\d*)`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
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
				secondname = ctx.GetThisGroupMemberInfo(qq, false).Get("nickname").Str
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
