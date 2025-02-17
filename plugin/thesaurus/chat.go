// Package thesaurus 修改过的单纯回复插件, 仅@触发
package thesaurus

import (
	"bytes"
	"math/rand"
	"strings"

	"github.com/fumiama/jieba"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/kimoi"
	"github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "词典匹配回复, 仅@触发",
		Help:             "- 切换[kimo|傲娇|可爱]词库",
		PublicDataFolder: "Chat",
	})
	engine.OnRegex(`^切换(kimo|傲娇|可爱)词库$`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: 找不到 manager"))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		d := c.GetData(gid)
		t := int64(0)
		switch ctx.State["regex_matched"].([]string)[1] {
		case "kimo":
			t = tKIMO
		case "傲娇":
			t = tDERE
		case "可爱":
			t = tKAWA
		}
		err := c.SetData(gid, (d&^3)|t)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	go func() {
		data, err := engine.GetLazyData("dict.txt", false)
		if err != nil {
			panic(err)
		}
		seg, err := jieba.LoadDictionary(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		smd, err := engine.GetLazyData("simai.yml", false)
		if err != nil {
			panic(err)
		}
		sm := simai{D: make(map[string][]string, 8192), K: make(map[string][]string, 16384)}
		err = yaml.Unmarshal(smd, &sm)
		if err != nil {
			panic(err)
		}
		chatListD := make([]string, 0, len(sm.D))
		for k := range sm.D {
			chatListD = append(chatListD, k)
		}
		chatListK := make([]string, 0, len(sm.K))
		for k := range sm.K {
			chatListK = append(chatListK, k)
		}
		logrus.Infoln("[thesaurus]加载", len(chatListD), "条傲娇词库", len(chatListK), "条可爱词库")

		engine.OnMessage(zero.OnlyToMe, canmatch(tKIMO)).
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
		engine.OnMessage(zero.OnlyToMe, canmatch(tDERE), match(chatListD, seg)).
			SetBlock(false).
			Handle(randreply(sm.D))
		engine.OnMessage(zero.OnlyToMe, canmatch(tKAWA), match(chatListK, seg)).
			SetBlock(false).
			Handle(randreply(sm.K))
	}()
}

type simai struct {
	D map[string][]string `yaml:"傲娇"`
	K map[string][]string `yaml:"可爱"`
}

const (
	tKIMO = iota
	tDERE
	tKAWA
)

func match(l []string, seg *jieba.Segmenter) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		return ctxext.JiebaSimilarity(0.66, seg, func(ctx *zero.Ctx) string {
			return ctx.ExtractPlainText()
		}, l...)(ctx)
	}
}

func canmatch(typ int64) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if zero.HasPicture(ctx) {
			return false
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			return false
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		d := c.GetData(gid)
		return ctx.ExtractPlainText() != "" && d&3 == typ
	}
}

func randreply(m map[string][]string) zero.Handler {
	return func(ctx *zero.Ctx) {
		ctx.Block()
		key := ctx.State["matched"].(string)
		val := m[key]
		nick := zero.BotConfig.NickName[rand.Intn(len(zero.BotConfig.NickName))]
		text := val[rand.Intn(len(val))]
		text = strings.ReplaceAll(text, "{name}", ctx.CardOrNickName(ctx.Event.UserID))
		text = strings.ReplaceAll(text, "{me}", nick)
		id := ctx.Event.MessageID
		for _, t := range strings.Split(text, "{segment}") {
			if t == "" {
				continue
			}
			process.SleepAbout1sTo2s()
			id = ctx.SendChain(message.Reply(id), message.Text(t))
		}
	}
}
