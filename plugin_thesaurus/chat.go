// Package thesaurus 修改过的单纯回复插件
package thesaurus

import (
	"math/rand"

	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	dbpath = "data/Purechat/"
	dbfile = dbpath + "kimoi_clear.json"
)

var (
	engine = control.Register("thesaurus", order.AcquirePrio(), &control.Options{
		DisableOnDefault: ture,
		Help:             "thesaurus\n- 稍微干净一点的chat回复 删掉了一些有点ex的w((",
	})
	kimogomap  = make(kimogo, 256)
	chatList = make([]string, 0, 256)
)

func init() { 
	initPureChatList(func() {
		engine.OnFullMatchGroup(chatList, zero.OnlyToMe).SetBlock(true).Handle(
			func(ctx *zero.Ctx) {
				key := ctx.MessageString()
				val := *kimogomap[key]
				text := val[rand.Intn(len(val))]
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text)) // 来自于 https://github.com/Kyomotoi/AnimeThesaurus 的回复 经过二次修改
			})
	})
}