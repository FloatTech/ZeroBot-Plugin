// Package thesaurus 修改过的单纯回复插件, 仅@触发
package thesaurus

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/kimoi"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "词典匹配回复, 仅@触发",
		PublicDataFolder: "Chat",
	})
	engine.OnMessage(zero.OnlyToMe, canmatch()).
		SetBlock(false).Handle(func(ctx *zero.Ctx) {
		msg := ctx.ExtractPlainText()
		r, err := kimoi.Chat(msg)
		if err == nil {
			c := 0
			for r.Confidence < 0.2 && c < 3 {
				r, err = kimoi.Chat(msg)
				if err != nil {
					return
				}
				c++
			}
			if r.Confidence < 0.2 {
				return
			}
			ctx.Block()
			ctx.SendChain(message.Text(r.Reply))
		}
	})
}

func canmatch() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if zero.HasPicture(ctx) {
			return false
		}
		return ctx.ExtractPlainText() != ""
	}
}
