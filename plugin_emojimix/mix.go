package emojimix

import (
	"fmt"
	"net/http"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	control.Register("emojimix", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "合成emoji\n" +
			"- [emoji][emoji]",
	}).OnMessage(match(emojis)).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			r := []rune(ctx.Event.RawMessage)
			e1 := string(r[0])
			e2 := string(r[1])
			u1 := fmt.Sprintf("https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png", emojis[e1], []rune(e1)[0], []rune(e1)[0], []rune(e2)[0])
			u2 := fmt.Sprintf("https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png", emojis[e2], []rune(e2)[0], []rune(e2)[0], []rune(e1)[0])
			client := &http.Client{}
			resp1, err := client.Head(u1)
			if err == nil && resp1.StatusCode == http.StatusOK {
				ctx.SendChain(message.Image(u1))
				resp1.Body.Close()
				return
			}
			resp2, err := client.Head(u2)
			if err == nil && resp2.StatusCode == http.StatusOK {
				ctx.SendChain(message.Image(u2))
				resp1.Body.Close()
				return
			}
			ctx.SendChain(message.Text("404 Not found"))
		})
}

func match(emojis map[string]int64) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		r := []rune(ctx.Event.RawMessage)
		if len(r) != 2 {
			return false
		}
		e1 := string(r[0])
		e2 := string(r[1])
		if _, ok := emojis[e1]; !ok {
			return false
		}
		if _, ok := emojis[e2]; !ok {
			return false
		}
		return true
	}
}
