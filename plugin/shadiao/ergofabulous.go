package shadiao

import (
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/antchfx/htmlquery"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnFullMatch("马丁路德骂我").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		doc, err := htmlquery.LoadURL(ergofabulousURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		node, err := htmlquery.Query(doc, "//main[@role=\"main\"]/p[@class=\"larger\"]/text()")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(node.Data))
	})
}
