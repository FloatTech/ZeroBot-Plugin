// Package thesaurus 修改过的单纯回复插件
package thesaurus

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/fumiama/jieba"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gopkg.in/yaml.v3"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "词典匹配回复",
		Help:             "- 切换[kimo|傲娇|可爱|🦙]词库\n- 设置词库触发概率0.x (0<x<9)",
		PublicDataFolder: "Chat",
	})
	alpacafolder := engine.DataFolder() + "alpaca/"
	err := os.MkdirAll(alpacafolder, 0755)
	if err != nil {
		panic(err)
	}
	alpacapifile := alpacafolder + "api.txt"
	alpacapiurl := ""
	if file.IsExist(alpacapifile) {
		data, err := os.ReadFile(alpacapifile)
		if err != nil {
			panic(err)
		}
		alpacapiurl = binary.BytesToString(data)
	}
	alpacatokenfile := alpacafolder + "token.txt"
	alpacatoken := ""
	if file.IsExist(alpacatokenfile) {
		data, err := os.ReadFile(alpacatokenfile)
		if err != nil {
			panic(err)
		}
		alpacatoken = binary.BytesToString(data)
	}
	engine.OnRegex(`^切换(kimo|傲娇|可爱|🦙)词库$`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
		case "🦙":
			t = tALPACA
		}
		err := c.SetData(gid, (d&^3)|t)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	engine.OnRegex(`^设置词库触发概率\s*0.(\d)$`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: 找不到 manager"))
			return
		}
		n := ctx.State["regex_matched"].([]string)[1][0] - '0'
		if n <= 0 || n >= 9 {
			ctx.SendChain(message.Text("ERROR: 概率越界"))
			return
		}
		n-- // 0~7
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		d := c.GetData(gid)
		err := c.SetData(gid, (d&3)|(int64(n)<<59))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	engine.OnRegex(`^设置🦙API地址\s*(http.*)\s*$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		alpacapiurl = ctx.State["regex_matched"].([]string)[1]
		err := os.WriteFile(alpacapifile, binary.StringToBytes(alpacapiurl), 0644)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	engine.OnRegex(`^设置🦙token\s*([0-9a-f]{112})\s*$`, zero.SuperUserPermission, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		alpacatoken = ctx.State["regex_matched"].([]string)[1]
		err := os.WriteFile(alpacatokenfile, binary.StringToBytes(alpacatoken), 0644)
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
		data, err = engine.GetLazyData("kimoi.json", false)
		if err != nil {
			panic(err)
		}
		kimomap := make(kimo, 256)
		err = json.Unmarshal(data, &kimomap)
		if err != nil {
			panic(err)
		}
		chatList := make([]string, 0, len(kimomap))
		for k := range kimomap {
			chatList = append(chatList, k)
		}
		logrus.Infoln("[thesaurus]加载", len(chatList), "条kimoi")

		chatListD := make([]string, 0, len(sm.D))
		for k := range sm.D {
			chatListD = append(chatListD, k)
		}
		chatListK := make([]string, 0, len(sm.K))
		for k := range sm.K {
			chatListK = append(chatListK, k)
		}
		logrus.Infoln("[thesaurus]加载", len(chatListD), "条傲娇词库", len(chatListK), "条可爱词库")

		engine.OnMessage(canmatch(tKIMO), match(chatList, seg)).
			SetBlock(false).
			Handle(randreply(kimomap))
		engine.OnMessage(canmatch(tDERE), match(chatListD, seg)).
			SetBlock(false).
			Handle(randreply(sm.D))
		engine.OnMessage(canmatch(tKAWA), match(chatListK, seg)).
			SetBlock(false).
			Handle(randreply(sm.K))
		engine.OnMessage(canmatch(tALPACA), func(_ *zero.Ctx) bool {
			return alpacapiurl != "" && alpacatoken != ""
		}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			if msg != "" {
				data, err := web.RequestDataWithHeaders(http.DefaultClient, alpacapiurl+"/reply", "POST",
					func(r *http.Request) error {
						r.Header.Set("Authorization", alpacatoken)
						return nil
					}, bytes.NewReader(binary.NewWriterF(func(writer *binary.Writer) {
						_ = json.NewEncoder(writer).Encode(&[]alpacamsg{{
							Name:    ctx.CardOrNickName(ctx.Event.UserID),
							Message: msg,
						}})
					})))
				if err != nil {
					logrus.Warnln("[chat] 🦙 err:", err)
					return
				}
				type reply struct {
					ID  int
					Msg string
				}
				m := reply{}
				err = json.Unmarshal(data, &m)
				if err != nil {
					logrus.Warnln("[chat] 🦙 unmarshal err:", err)
					return
				}
				for i := 0; i < 60; i++ {
					time.Sleep(time.Second * 4)
					data, err := web.RequestDataWithHeaders(http.DefaultClient, alpacapiurl+"/get?id="+strconv.Itoa(m.ID), "GET",
						func(r *http.Request) error {
							r.Header.Set("Authorization", alpacatoken)
							return nil
						}, nil)
					if err != nil {
						continue
					}
					err = json.Unmarshal(data, &m)
					if err != nil {
						logrus.Warnln("[chat] 🦙 unmarshal err:", err)
						return
					}
					if len(m.Msg) > 0 {
						ctx.Send(message.Text(m.Msg))
					}
					return
				}
			}
		})
	}()
}

type kimo = map[string][]string

type simai struct {
	D map[string][]string `yaml:"傲娇"`
	K map[string][]string `yaml:"可爱"`
}

type alpacamsg struct {
	Name    string
	Message string
}

const (
	tKIMO = iota
	tDERE
	tKAWA
	tALPACA
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
		return d&3 == typ && rand.Int63n(10) <= d>>59
	}
}

func randreply(m map[string][]string) zero.Handler {
	return func(ctx *zero.Ctx) {
		key := ctx.State["matched"].(string)
		val := m[key]
		nick := zero.BotConfig.NickName[rand.Intn(len(zero.BotConfig.NickName))]
		text := val[rand.Intn(len(val))]
		text = strings.ReplaceAll(text, "{name}", ctx.CardOrNickName(ctx.Event.UserID))
		text = strings.ReplaceAll(text, "{me}", nick)
		id := ctx.Event.MessageID
		for _, t := range strings.Split(text, "{segment}") {
			process.SleepAbout1sTo2s()
			id = ctx.SendChain(message.Reply(id), message.Text(t))
		}
	}
}
