package aireply

import (
	"errors"
	"strings"

	"github.com/RomiChan/syncx"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/tts"
	"github.com/FloatTech/AnimeAPI/tts/baidutts"
	"github.com/FloatTech/AnimeAPI/tts/genshin"
	"github.com/FloatTech/AnimeAPI/tts/lolimi"
	"github.com/FloatTech/AnimeAPI/tts/ttscn"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

// 数据结构: [8 bits] [8 bits] [8 bits]
// 			[具体人物] [tts模式] [回复模式]

// defaultttsindexkey
// 数据结构: [8 bits] [8 bits]
// 			[具体人物] [tts模式]

// [tts模式]: 0~200 genshin 201 baidu 202 ttscn 203 lolimi

const (
	baiduttsindex = 201 + iota
	ttscnttsindex
	lolimittsindex
)

// extrattsname is the tts other than genshin vits
var extrattsname = []string{"百度", "TTSCN", "桑帛云"}

var ttscnspeakers = [...]string{
	"晓晓（女 - 年轻人）",
	"云扬（男 - 年轻人）",
	"晓辰（女 - 年轻人 - 抖音热门）",
	"晓涵（女 - 年轻人）",
	"晓墨（女 - 年轻人）",
	"晓秋（女 - 中年人）",
	"晓睿（女 - 老年）",
	"晓双（女 - 儿童）",
	"晓萱（女 - 年轻人）",
	"晓颜（女 - 年轻人）",
	"晓悠（女 - 儿童）",
	"云希（男 - 年轻人 - 抖音热门）",
	"云野（男 - 中年人）",
	"晓梦（女 - 年轻人）",
	"晓伊（女 - 儿童）",
	"晓甄（女 - 年轻人）",
}

const defaultttsindexkey = -2905

var (
	原  = newapikeystore("./data/tts/o.txt")
	ཆཏ = newapikeystore("./data/tts/c.txt")
	百  = newapikeystore("./data/tts/b.txt")
)

type replymode []string

func (r replymode) setReplyMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range r {
		if s == name {
			ok = true
			index = int64(i)
			break
		}
	}
	if !ok {
		return errors.New("no such mode")
	}
	m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return errors.New("no such plugin")
	}
	return m.SetData(gid, (m.GetData(gid)&^0xff)|(index&0xff))
}

func (r replymode) getReplyMode(ctx *zero.Ctx) aireply.AIReply {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if ok {
		switch m.GetData(gid) & 0xff {
		case 0:
			return aireply.NewLolimiAi(aireply.JingfengURL, aireply.JingfengBotName)
		case 1:
			return aireply.NewLolimiAi(aireply.MomoURL, aireply.MomoBotName)
		case 2:
			return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
		case 3:
			return aireply.NewXiaoAi(aireply.XiaoAiURL, aireply.XiaoAiBotName)
		case 4:
			k := ཆཏ.k
			if k != "" {
				return aireply.NewChatGPT(aireply.ChatGPTURL, k)
			}
			return aireply.NewLolimiAi(aireply.JingfengURL, aireply.JingfengBotName)
		}
	}
	return aireply.NewLolimiAi(aireply.JingfengURL, aireply.JingfengBotName)
}

var ttsins = func() map[string]tts.TTS {
	m := make(map[string]tts.TTS, 512)
	for _, mode := range append(genshin.SoundList[:], extrattsname...) {
		m[mode] = nil
	}
	return m
}()

var ttsModes = func() []string {
	s := append(genshin.SoundList[:], make([]string, baiduttsindex-len(genshin.SoundList))...) // 0-200
	s = append(s, extrattsname...)                                                             // 201 202 ...
	return s
}()

type ttsmode syncx.Map[int64, int64]

func list(list []string, num int) string {
	s := ""
	for i, value := range list {
		s += value
		if (i+1)%num == 0 {
			s += "\n"
		} else {
			s += " | "
		}
	}
	return s
}

func newttsmode() *ttsmode {
	t := &ttsmode{}
	m, ok := control.Lookup("tts")
	(*syncx.Map[int64, int64])(t).Store(defaultttsindexkey, 0)
	if ok {
		index := m.GetData(defaultttsindexkey)
		msk := index & 0xff
		if msk >= 0 && (msk < int64(len(ttsModes))) {
			(*syncx.Map[int64, int64])(t).Store(defaultttsindexkey, index)
		}
	}
	return t
}

func (t *ttsmode) setSoundMode(ctx *zero.Ctx, name string, character int) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	_, ok := ttsins[name]
	if !ok {
		return errors.New("不支持设置语音人物" + name)
	}
	var index = int64(-1)
	for i, s := range genshin.SoundList {
		if s == name {
			index = int64(i + 1)
			break
		}
	}
	if index == -1 {
		switch name {
		case extrattsname[0]:
			index = baiduttsindex
		case extrattsname[1]:
			index = ttscnttsindex
		case extrattsname[2]:
			index = lolimittsindex
		default:
			return errors.New("语音人物" + name + "未注册index")
		}
	}
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	// 按原来的逻辑map存的是前16位
	storeIndex := (m.GetData(gid) &^ 0xffff00) | ((index << 8) & 0xff00) | ((int64(character) << 16) & 0xff0000)
	(*syncx.Map[int64, int64])(t).Store(gid, (storeIndex>>8)&0xffff)
	return m.SetData(gid, storeIndex)
}

func (t *ttsmode) getSoundMode(ctx *zero.Ctx) (tts.TTS, error) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	i, ok := (*syncx.Map[int64, int64])(t).Load(gid)
	if !ok {
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		i = m.GetData(gid) >> 8
	}
	m := i & 0xff
	if m <= 0 || (m >= int64(len(ttsModes))) {
		i, _ = (*syncx.Map[int64, int64])(t).Load(defaultttsindexkey)
		if i == 0 {
			i = ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).GetData(defaultttsindexkey)
			(*syncx.Map[int64, int64])(t).Store(defaultttsindexkey, i)
		}
		m = i & 0xff
	}
	mode := ttsModes[m]
	ins, ok := ttsins[mode]
	if !ok || ins == nil {
		switch mode {
		case extrattsname[0]:
			id, sec, _ := strings.Cut(百.k, ",")
			ins = baidutts.NewBaiduTTS(int(i&0xff00)>>8, id, sec)
		case extrattsname[1]:
			var err error
			ins, err = ttscn.NewTTSCN("中文（普通话，简体）", ttscnspeakers[int(i&0xff00)>>8], ttscn.KBRates[0])
			if err != nil {
				return nil, err
			}
		case extrattsname[2]:
			ins = lolimi.NewLolimi(int(i&0xff00) >> 8)
		default: // 原神
			k := 原.k
			if k != "" {
				ins = genshin.NewGenshin(int(m-1), 原.k)
				ttsins[mode] = ins
			} else {
				ins = lolimi.NewLolimi(int(i&0xff00) >> 8)
			}
		}
	}
	return ins, nil
}

func (t *ttsmode) resetSoundMode(ctx *zero.Ctx) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	// 只保留后面8位
	(*syncx.Map[int64, int64])(t).Delete(gid)
	return m.SetData(gid, (m.GetData(gid) & 0xff)) // 重置数据
}

func (t *ttsmode) setDefaultSoundMode(name string, character int) error {
	_, ok := ttsins[name]
	if !ok {
		return errors.New("不支持设置语音人物" + name)
	}
	index := int64(-1)
	for i, s := range genshin.SoundList {
		if s == name {
			index = int64(i + 1)
			break
		}
	}
	if index == -1 {
		switch name {
		case extrattsname[0]:
			index = baiduttsindex
		case extrattsname[1]:
			index = ttscnttsindex
		case extrattsname[2]:
			index = lolimittsindex
		default:
			return errors.New("语音人物" + name + "未注册index")
		}
	}
	m, ok := control.Lookup("tts")
	if !ok {
		return errors.New("[tts] service not found")
	}
	storeIndex := (index & 0xff) | ((int64(character) << 8) & 0xff00)
	(*syncx.Map[int64, int64])(t).Store(defaultttsindexkey, storeIndex)
	return m.SetData(defaultttsindexkey, storeIndex)
}
