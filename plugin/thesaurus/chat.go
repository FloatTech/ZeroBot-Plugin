// Package thesaurus 修改过的单纯回复插件
package thesaurus

import (
	"encoding/json"
	"math/rand"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type kimo = map[string][]string

type simai struct {
	D map[string][]string `yaml:"傲娇"`
	K map[string][]string `yaml:"可爱"`
}

func init() {
	engine := control.Register("thesaurus", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "词典匹配回复",
		Help:             "- 词典匹配回复",
		PublicDataFolder: "Chat",
	})
	go func() {
		data, err := engine.GetLazyData("kimoi.json", false)
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
				val := kimomap[key]
				text := val[rand.Intn(len(val))]
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text)) // 来自于 https://github.com/Kyomotoi/AnimeThesaurus 的回复 经过二次修改
			})
	}()
}
