package aireply

import (
	"errors"
	"net/url"
	"sync"

	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/tts"
	"github.com/FloatTech/AnimeAPI/tts/baidutts"
	"github.com/FloatTech/AnimeAPI/tts/genshin"
	"github.com/FloatTech/AnimeAPI/tts/mockingbird"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

// 数据结构: [4 bits] [4 bits] [8 bits] [8 bits]
// 			[拟声鸟模式] [百度模式] [tts模式] [回复模式]

// defaultttsindexkey
// 数据结构: [4 bits] [4 bits] [8 bits]
// 			[拟声鸟模式] [百度模式] [tts模式]

// [tts模式]: 0~63 genshin 64 baidu 65 mockingbird

const (
	lastgsttsindex = 63 + iota
	baiduttsindex
	mockingbirdttsindex
)

const (
	defaultttsindexkey = -2905
)

var replyModes = [...]string{"青云客", "小爱", "ChatGPT"}

func setReplyMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range replyModes {
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
	return m.SetData((m.GetData(gid)&^0xff)|(gid&0xff), index)
}

var chats *aireply.ChatGPT

func getReplyMode(ctx *zero.Ctx) aireply.AIReply {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if ok {
		switch m.GetData(gid) & 0xff {
		case 0:
			return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
		case 1:
			return aireply.NewXiaoAi(aireply.XiaoAiURL, aireply.XiaoAiBotName)
		case 2:
			if chats != nil {
				return chats
			}
		}
	}
	return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
}

var ttsins = func() map[string]tts.TTS {
	m := make(map[string]tts.TTS, 128)
	for _, mode := range append(genshin.SoundList[:], "百度", "拟声鸟") {
		m[mode] = nil
	}
	return m
}()

var ttsModes = func() []string {
	s := append(genshin.SoundList[:], make([]string, 64-len(genshin.SoundList))...) // 0-63
	s = append(s, "百度", "拟声鸟")                                                      // 64 65
	return s
}()

type ttsmode struct {
	sync.Mutex `json:"-"`
	APIKey     string          // APIKey is for genshin vits
	mode       map[int64]int64 `json:"-"` // mode grp index
}

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
	t.Lock()
	defer t.Unlock()
	m, ok := control.Lookup("tts")
	t.mode = make(map[int64]int64, 2*len(genshin.SoundList))
	t.mode[defaultttsindexkey] = 0
	if ok {
		index := m.GetData(defaultttsindexkey)
		msk := index & 0xff
		if msk >= 0 && (msk < int64(len(genshin.SoundList)) || msk == baiduttsindex || msk == mockingbirdttsindex) {
			t.mode[defaultttsindexkey] = index
		}
	}
	return t
}

func (t *ttsmode) getAPIKey(ctx *zero.Ctx) string {
	if t.APIKey == "" {
		t.Lock()
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		_ = m.Manager.GetExtra(gid, &t)
		t.Unlock()
	}
	return url.QueryEscape(t.APIKey)
}

func (t *ttsmode) setAPIKey(m *ctrl.Control[*zero.Ctx], grp int64, key string) error {
	t.Lock()
	defer t.Unlock()
	err := m.Manager.SetExtra(grp, &key)
	if err != nil {
		return err
	}
	t.APIKey = key
	return nil
}

func (t *ttsmode) setSoundMode(ctx *zero.Ctx, name string, baiduper, mockingsynt int) error {
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
			index = int64(i)
			break
		}
	}
	if index == -1 {
		switch name {
		case "百度":
			index = baiduttsindex
		case "拟声鸟":
			index = mockingbirdttsindex
		default:
			return errors.New("语音人物" + name + "未注册index")
		}
	}
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	t.Lock()
	defer t.Unlock()
	t.mode[gid] = index
	return m.SetData(gid, (m.GetData(gid)&^0xffff00)|((index<<8)&0xff00)|((int64(baiduper)<<16)&0x0f0000)|((int64(mockingsynt)<<20)&0xf00000))
}

func (t *ttsmode) getSoundMode(ctx *zero.Ctx) (tts.TTS, error) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	t.Lock()
	defer t.Unlock()
	i, ok := t.mode[gid]
	if !ok {
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		i = m.GetData(gid) >> 8
	}
	m := i & 0xff
	if m < 0 || (m >= int64(len(genshin.SoundList)) && m != baiduttsindex && m != mockingbirdttsindex) {
		i = t.mode[defaultttsindexkey]
		m = i & 0xff
	}
	mode := ttsModes[m]
	ins, ok := ttsins[mode]
	if !ok {
		switch mode {
		case "百度":
			ins = baidutts.NewBaiduTTS(int(i&0x0f00) >> 8)
		case "拟声鸟":
			var err error
			ins, err = mockingbird.NewMockingBirdTTS(int(i&0xf000) >> 12)
			if err != nil {
				return nil, err
			}
		default: // 原神
			ins = genshin.NewGenshin(int(m), t.getAPIKey(ctx))
			ttsins[mode] = ins
		}
	}
	return ins, nil
}

func (t *ttsmode) resetSoundMode(ctx *zero.Ctx) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	t.Lock()
	defer t.Unlock()
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	index := m.GetData(defaultttsindexkey)
	return m.SetData(gid, (m.GetData(gid)&0xff)|((index&^0xff)<<8)) // 重置数据
}

func (t *ttsmode) setDefaultSoundMode(name string, baiduper, mockingsynt int) error {
	_, ok := ttsins[name]
	if !ok {
		return errors.New("不支持设置语音人物" + name)
	}
	index := int64(-1)
	for i, s := range genshin.SoundList {
		if s == name {
			index = int64(i)
			break
		}
	}
	if index == -1 {
		switch name {
		case "百度":
			index = baiduttsindex
		case "拟声鸟":
			index = mockingbirdttsindex
		default:
			return errors.New("语音人物" + name + "未注册index")
		}
	}
	t.Lock()
	defer t.Unlock()
	m, ok := control.Lookup("tts")
	if !ok {
		return errors.New("[tts] service not found")
	}
	t.mode[defaultttsindexkey] = index
	return m.SetData(defaultttsindexkey, (index&0xff)|((int64(baiduper)<<8)&0x0f00)|((int64(mockingsynt)<<12)&0xf000))
}
