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
	// 优先级设为 10，低于控制命令(SecondPriority=1)，
	// 避免 /全局禁用、/启用 等管理命令被词库回复抢先拦截
	chatm := engine.OnMessage(zero.OnlyToMe, canmatch()).SetBlock(false)
	(*zero.Matcher)(chatm).SetPriority(10)
	chatm.Handle(func(ctx *zero.Ctx) {
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
