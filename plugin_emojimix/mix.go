// Package emojimix 合成emoji
package emojimix

import (
	"fmt"
	"net/http"
	"strconv"

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
	}).OnMessage(match).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			r1 := ctx.State["emojimix_e1"].(rune)
			r2 := ctx.State["emojimix_e2"].(rune)
			u1 := fmt.Sprintf("https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png", emojis[r1], r1, r1, r2)
			u2 := fmt.Sprintf("https://www.gstatic.com/android/keyboard/emojikitchen/%d/u%x/u%x_u%x.png", emojis[r2], r2, r2, r1)
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
				resp2.Body.Close()
				return
			}
		})
}

func match(ctx *zero.Ctx) bool {
	r := []rune(ctx.Event.RawMessage)
	if len(r) == 2 {
		if _, ok := emojis[r[0]]; !ok {
			return false
		}
		if _, ok := emojis[r[1]]; !ok {
			return false
		}
		ctx.State["emojimix_e1"] = r[0]
		ctx.State["emojimix_e2"] = r[1]
		return true
	}
	if len(ctx.Event.Message) == 2 {
		r1 := face2emoji(ctx.Event.Message[0])
		if _, ok := emojis[r1]; !ok {
			return false
		}
		r2 := face2emoji(ctx.Event.Message[1])
		if _, ok := emojis[r2]; !ok {
			return false
		}
		ctx.State["emojimix_e1"] = r1
		ctx.State["emojimix_e2"] = r2
		return true
	}
	return false
}

func face2emoji(face message.MessageSegment) rune {
	if face.Type == "text" {
		r := []rune(face.Data["text"])
		if len(r) != 1 {
			return 0
		}
		return r[0]
	}
	if face.Type != "face" {
		return 0
	}
	id, err := strconv.Atoi(face.Data["id"])
	if err != nil {
		return 0
	}
	if r, ok := qqface[id]; ok {
		return r
	}
	return 0
}
