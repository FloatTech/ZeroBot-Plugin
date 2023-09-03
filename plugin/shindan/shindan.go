// Package shindan 基于 https://shindanmaker.com 的测定小功能
package shindan

import (
	"github.com/FloatTech/AnimeAPI/shindanmaker"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "shindan测定",
		Help: "- 今天是什么少女[@xxx]\n" +
			"- 异世界转生[@xxx]\n" +
			"- 卖萌[@xxx]\n" +
			"- 今日老婆[@xxx]\n" +
			"- 黄油角色[@xxx]",
	})
	engine.OnPrefix("异世界转生", number(587874)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(handlepic)
	engine.OnPrefix("今天是什么少女", number(162207)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(handlepic)
	engine.OnPrefix("卖萌", number(360578)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(handletxt)
	engine.OnPrefix("今日老婆", number(1075116)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(handlecq)
	engine.OnPrefix("黄油角色", number(1115465)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(handlepic)
}

func handletxt(ctx *zero.Ctx) {
	// 获取名字
	name := ctx.NickName()
	// 调用接口
	txt, err := shindanmaker.Shindanmaker(ctx.State["id"].(int64), name)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(message.Text(txt))
}

func handlecq(ctx *zero.Ctx) {
	// 获取名字
	name := ctx.NickName()
	// 调用接口
	txt, err := shindanmaker.Shindanmaker(ctx.State["id"].(int64), name)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.Send(txt)
}

func handlepic(ctx *zero.Ctx) {
	// 获取名字
	name := ctx.NickName()
	// 调用接口
	txt, err := shindanmaker.Shindanmaker(ctx.State["id"].(int64), name)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	data, err := text.RenderToBase64(txt, text.FontFile, 400, 20)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(data))); id.ID() == 0 {
		ctx.SendChain(message.Text("ERROR: 可能被风控了"))
	}
}

// 传入 shindanmaker id
func number(id int64) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		ctx.State["id"] = id
		return true
	}
}
