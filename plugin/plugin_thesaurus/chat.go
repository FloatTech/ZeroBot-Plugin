// Package thesaurus 修改过的单纯回复插件
package thesaurus

import (
	"encoding/json"
	"math/rand"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

type kimo = map[string]*[]string

func init() {
	engine := control.Register("thesaurus", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "thesaurus\n- 词典匹配回复",
		PublicDataFolder: "Chat",
	})
	go func() {
		defer order.DoneOnExit()()
		data, err := file.GetLazyData(engine.DataFolder()+"kimoi.json", true, true)
		if err != nil {
			panic(err)
		}
		kimomap := make(kimo, 256)
		err = json.Unmarshal(data, &kimomap)
		if err != nil {
			panic(err)
		}
		chatList := make([]string, 0, 256)
		for k := range kimomap {
			chatList = append(chatList, k)
		}
		logrus.Infoln("[thesaurus]加载", len(chatList), "条kimoi")
		engine.OnFullMatchGroup(chatList, zero.OnlyToMe).SetBlock(true).Handle(
			func(ctx *zero.Ctx) {
				key := ctx.MessageString()
				val := *kimomap[key]
				text := val[rand.Intn(len(val))]
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text)) // 来自于 https://github.com/Kyomotoi/AnimeThesaurus 的回复 经过二次修改
			})
	}()
}
