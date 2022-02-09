// Package reborn 投胎 来自 https://github.com/YukariChiba/tgbot/blob/main/modules/Reborn.py
package reborn

import (
	"fmt"
	"math/rand"
	"time"

	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	control.Register("reborn", order.PrioReborn, &control.Options{
		DisableOnDefault: false,
		Help:             "投胎\n- reborn",
	}).OnFullMatch("reborn").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if rand.Int31() > 1<<27 {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("投胎成功！\n您出生在 %s, 是 %s。", randcoun(), randgen())))
			} else {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text("投胎失败！\n您没能活到出生，祝您下次好运！"))
			}
		})
}
