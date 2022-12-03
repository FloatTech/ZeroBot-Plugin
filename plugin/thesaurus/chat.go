// Package thesaurus 修改过的单纯回复插件
package thesaurus

import (
	"encoding/json"
	"io"
	"io/fs"
	"math/rand"
	"strings"

	"github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/fumiama/jieba"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gopkg.in/yaml.v3"
)

func init() {
	engine := control.Register("thesaurus", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "词典匹配回复",
		Help:             "- 切换[kimo|傲娇|可爱]词库\n- 设置词库触发概率0.x (0<x<9)",
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
	go func() {
		data, err := engine.GetLazyData("dict.txt", false)
		if err != nil {
			panic(err)
		}
		seg, err := jieba.LoadDictionary(&mockfile{data: data})
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

		engine.OnMessage(zero.OnlyToMe,
			ctxext.JiebaFullMatch(seg, getmsg, chatList...),
		).SetBlock(true).Handle(randreply(sm.D))
		engine.OnMessage(zero.OnlyToMe,
			ctxext.JiebaFullMatch(seg, getmsg, chatList...),
		).SetBlock(true).Handle(randreply(sm.K))
		engine.OnMessage(zero.OnlyToMe,
			ctxext.JiebaFullMatch(seg, getmsg, chatList...),
		).SetBlock(true).Handle(randreply(kimomap))
		engine.OnMessage(canmatch(tKIMO),
			ctxext.JiebaFullMatch(seg, getmsg, chatList...),
		).SetBlock(false).Handle(randreply(kimomap))
		engine.OnMessage(canmatch(tDERE),
			ctxext.JiebaFullMatch(seg, getmsg, chatListD...),
		).SetBlock(false).Handle(randreply(sm.D))
		engine.OnMessage(canmatch(tKAWA),
			ctxext.JiebaFullMatch(seg, getmsg, chatListK...),
		).SetBlock(false).Handle(randreply(sm.K))
	}()
}

type kimo = map[string][]string

type simai struct {
	D map[string][]string `yaml:"傲娇"`
	K map[string][]string `yaml:"可爱"`
}

type mockfile struct {
	p    uintptr
	data []byte
}

func (*mockfile) Stat() (fs.FileInfo, error) {
	return nil, nil
}
func (f *mockfile) Read(buf []byte) (int, error) {
	if int(f.p) >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(buf, f.data[f.p:])
	f.p += uintptr(n)
	return n, nil
}
func (f *mockfile) Close() error {
	if f.data == nil {
		return fs.ErrClosed
	}
	f.data = nil
	return nil
}

const (
	tKIMO = iota
	tDERE
	tKAWA
)

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

func getmsg(ctx *zero.Ctx) string {
	return ctx.MessageString()
}

func randreply(m map[string][]string) zero.Handler {
	return func(ctx *zero.Ctx) {
		key := ctx.State["matched"].(string)
		val := m[key]
		text := val[rand.Intn(len(val))]
		text = strings.ReplaceAll(text, "{name}", ctx.CardOrNickName(ctx.Event.UserID))
		id := ctx.Event.MessageID
		for _, t := range strings.Split(text, "{segment}") {
			process.SleepAbout1sTo2s()
			id = ctx.SendChain(message.Reply(id), message.Text(t))
		}
	}
}
