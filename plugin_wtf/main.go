package wtf

import (
	"fmt"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

var (
	// 限制调用频率
	limit = rate.NewManager(time.Minute*5, 5)
)

func init() {
	en := control.Register("wtf", &control.Options{
		DisableOnDefault: false,
		Help:             "鬼东西\n- 鬼东西列表\n- 查询鬼东西xxx(@xxx)",
	})
	en.OnFullMatch("鬼东西列表").SetBlock(true).SetPriority(30).
		Handle(func(ctx *zero.Ctx) {
			s := ""
			i := 0
			for k := range pathtable {
				s += fmt.Sprintf("%02d. %s\n", i, k)
				i++
			}
			ctx.SendChain(message.Text(s))
		})
	en.OnRegex(`^查询鬼东西(.*)$`).SetBlock(false).SetPriority(30).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.SendChain(message.Text("请稍后重试0x0..."))
				return
			}
			// 调用接口
			w := NewWtf(ctx.State["regex_matched"].([]string)[1])
			if w == nil {
				ctx.SendChain(message.Text("没有这项内容！"))
				return
			}
			// 获取名字
			name := ctx.State["args"].(string)
			if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
				qq, _ := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
				name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("nickname").Str
			} else if name == "" {
				name = ctx.Event.Sender.NickName
			}
			text, err := w.Predict(name)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			// TODO: 可注入
			ctx.Send(text)
		})
}
