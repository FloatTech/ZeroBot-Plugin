package aireply

import (
	"errors"
	"net/url"

	"github.com/RomiChan/syncx"
	"github.com/sirupsen/logrus"
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

// extrattsname is the tts other than genshin vits
var extrattsname = []string{"百度", "拟声鸟"}

const (
	defaultttsindexkey    = -2905
	gsapikeyextragrp      = -1
	chatgptapikeyextragrp = -2
)

type replymode struct {
	APIKey     string   // APIKey is for chatgpt
	replyModes []string `json:"-"`
}

func (r *replymode) setReplyMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range r.replyModes {
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
	return m.SetData(gid, (m.GetData(index)&^0xff)|(index&0xff))
}

func (r *replymode) getReplyMode(ctx *zero.Ctx) aireply.AIReply {
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
			k := r.getAPIKey(ctx)
			if k != "" {
				return aireply.NewChatGPT(aireply.ChatGPTURL, k)
			}
			return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
		}
	}
	return aireply.NewQYK(aireply.QYKURL, aireply.QYKBotName)
}

func (r *replymode) getAPIKey(ctx *zero.Ctx) string {
	if r.APIKey == "" {
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		_ = m.Manager.GetExtra(chatgptapikeyextragrp, &r)
		logrus.Debugln("[tts] get api key:", r.APIKey)
	}
	return r.APIKey
}

func (r *replymode) setAPIKey(m *ctrl.Control[*zero.Ctx], key string) error {
	r.APIKey = key
	_ = m.Manager.Response(chatgptapikeyextragrp)
	return m.Manager.SetExtra(chatgptapikeyextragrp, r)
}

var ttsins = func() map[string]tts.TTS {
	m := make(map[string]tts.TTS, 128)
	for _, mode := range append(genshin.SoundList[:], extrattsname...) {
		m[mode] = nil
	}
	return m
}()

var ttsModes = func() []string {
	s := append(genshin.SoundList[:], make([]string, 64-len(genshin.SoundList))...) // 0-63
	s = append(s, extrattsname...)                                                  // 64 65 ...
	return s
}()

type ttsmode struct {
	APIKey string                  // APIKey is for genshin vits
	mode   syncx.Map[int64, int64] `json:"-"` // mode grp index
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
	m, ok := control.Lookup("tts")
	t.mode = syncx.Map[int64, int64]{}
	t.mode.Store(defaultttsindexkey, 0)
	if ok {
		index := m.GetData(defaultttsindexkey)
		msk := index & 0xff
		if msk >= 0 && (msk < int64(len(genshin.SoundList)) || msk == baiduttsindex || msk == mockingbirdttsindex) {
			t.mode.Store(defaultttsindexkey, index)
		}
	}
	return t
}

func (t *ttsmode) getAPIKey(ctx *zero.Ctx) string {
	if t.APIKey == "" {
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		_ = m.Manager.GetExtra(gsapikeyextragrp, &t)
		logrus.Debugln("[tts] get api key:", t.APIKey)
	}
	return url.QueryEscape(t.APIKey)
}

func (t *ttsmode) setAPIKey(m *ctrl.Control[*zero.Ctx], key string) error {
	t.APIKey = key
	_ = m.Manager.Response(gsapikeyextragrp)
	return m.Manager.SetExtra(gsapikeyextragrp, t)
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
		case extrattsname[0]:
			index = baiduttsindex
		case extrattsname[1]:
			index = mockingbirdttsindex
		default:
			return errors.New("语音人物" + name + "未注册index")
		}
	}
	m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	t.mode.Store(gid, index)
	return m.SetData(gid, (m.GetData(gid)&^0xffff00)|((index<<8)&0xff00)|((int64(baiduper)<<16)&0x0f0000)|((int64(mockingsynt)<<20)&0xf00000))
}

func (t *ttsmode) getSoundMode(ctx *zero.Ctx) (tts.TTS, error) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	i, ok := t.mode.Load(gid)
	if !ok {
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		i = m.GetData(gid) >> 8
	}
	m := i & 0xff
	if m < 0 || (m >= int64(len(genshin.SoundList)) && m != baiduttsindex && m != mockingbirdttsindex) {
		i, _ = t.mode.Load(defaultttsindexkey)
		m = i & 0xff
	}
	mode := ttsModes[m]
	ins, ok := ttsins[mode]
	if !ok || ins == nil {
		switch mode {
		case extrattsname[0]:
			ins = baidutts.NewBaiduTTS(int(i&0x0f00) >> 8)
		case extrattsname[1]:
			var err error
			ins, err = mockingbird.NewMockingBirdTTS(int(i&0xf000) >> 12)
			if err != nil {
				return nil, err
			}
		default: // 原神
			k := t.getAPIKey(ctx)
			if k != "" {
				ins = genshin.NewGenshin(int(m), t.getAPIKey(ctx))
				ttsins[mode] = ins
			} else {
				return nil, errors.New("no valid speaker")
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
		case extrattsname[0]:
			index = baiduttsindex
		case extrattsname[1]:
			index = mockingbirdttsindex
		default:
			return errors.New("语音人物" + name + "未注册index")
		}
	}
	m, ok := control.Lookup("tts")
	if !ok {
		return errors.New("[tts] service not found")
	}
	t.mode.Store(defaultttsindexkey, index)
	return m.SetData(defaultttsindexkey, (index&0xff)|((int64(baiduper)<<8)&0x0f00)|((int64(mockingsynt)<<12)&0xf000))
}
