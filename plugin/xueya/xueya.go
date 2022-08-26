package xueya

import (
	"math/rand"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/process"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("xueya", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help:             "增加血压! 在匹配字符时为用户增加血压值, 并在血压爆表的时候提示.",
	})

	engine.OnKeyword("啧").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randText("123", "321"))
		})
}

func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}
