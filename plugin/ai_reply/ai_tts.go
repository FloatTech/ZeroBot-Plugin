package aireply

import (
	"errors"
	"regexp"
	"strconv"
	"sync"

	"github.com/pkumza/numcn"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/tts"
	"github.com/FloatTech/AnimeAPI/tts/baidutts"
	"github.com/FloatTech/AnimeAPI/tts/mockingbird"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"

	"github.com/FloatTech/zbputils/control/order"
)

const ttsServiceName = "tts"

var (
	t = &ttsInstances{
		m: map[string]tts.TTS{
			"百度女声":   baidutts.NewBaiduTTS(0),
			"百度男声":   baidutts.NewBaiduTTS(1),
			"百度度逍遥":  baidutts.NewBaiduTTS(3),
			"百度度丫丫":  baidutts.NewBaiduTTS(4),
			"拟声鸟阿梓":  mockingbird.NewMockingBirdTTS(0),
			"拟声鸟药水哥": mockingbird.NewMockingBirdTTS(1),
		},
		l: []string{"拟声鸟阿梓", "拟声鸟药水哥", "百度女声", "百度男声", "百度度逍遥", "百度度丫丫"},
	}
	re = regexp.MustCompile(`(\-|\+)?\d+(\.\d+)?`)
)

type ttsInstances struct {
	sync.RWMutex
	m map[string]tts.TTS
	l []string
}

func (t *ttsInstances) List() []string {
	t.RLock()
	cl := make([]string, len(t.l))
	_ = copy(cl, t.l)
	t.RUnlock()
	return cl
}

func init() {
	engine := control.Register(ttsServiceName, order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "语音回复(包括拟声鸟和百度)\n" +
			"- @Bot 任意文本(任意一句话回复)\n" +
			"- 设置语音模式[拟声鸟阿梓 | 拟声鸟药水哥 | 百度女声 | 百度男声| 百度度逍遥 | 百度度丫丫]\n" +
			"- 设置默认语音模式[拟声鸟阿梓 | 拟声鸟药水哥 | 百度女声 | 百度男声| 百度度逍遥 | 百度度丫丫]\n",
	})
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			tts := t.new(t.getSoundMode(ctx))
			ctx.SendChain(message.Record(tts.Speak(ctx.Event.UserID, func() string {
				reply := r.TalkPlain(msg, zero.BotConfig.NickName[0])
				reply = re.ReplaceAllStringFunc(reply, func(s string) string {
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						log.Errorln("[tts]:", err)
						return s
					}
					return numcn.EncodeFromFloat64(f)
				})
				log.Println("[tts]:", reply)
				return reply
			})))
		})
	engine.OnRegex(`^设置语音模式(.*)$`, ctxext.FirstValueInList(t)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["regex_matched"].([]string)[1]
			err := t.setSoundMode(ctx, param)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，当前模式为", param))
		})
	engine.OnRegex(`^设置默认语音模式(.*)$`, ctxext.FirstValueInList(t)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["regex_matched"].([]string)[1]
			t.setDefaultSoundMode(param)
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，默认模式为", param))
		})
}

// new 语音简单工厂
func (t *ttsInstances) new(name string) tts.TTS {
	return t.m[name]
}

func (t *ttsInstances) setSoundMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var index int64
	t.RLock()
	for i, s := range t.l {
		if s == name {
			index = int64(i)
			break
		}
	}
	t.RUnlock()
	m, ok := control.Lookup(ttsServiceName)
	if !ok {
		return errors.New("no such plugin")
	}
	return m.SetData(gid, index)
}

func (t *ttsInstances) getSoundMode(ctx *zero.Ctx) (name string) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := control.Lookup(ttsServiceName)
	if ok {
		t.RLock()
		defer t.RUnlock()
		index := m.GetData(gid)
		if int(index) < len(t.l) {
			return t.l[index]
		}
	}
	return "拟声鸟阿梓"
}

func (t *ttsInstances) setDefaultSoundMode(name string) {
	var index int
	t.RLock()
	for _, s := range t.l {
		if s == name {
			break
		}
		index++
	}
	t.RUnlock()
	t.Lock()
	t.l[0], t.l[index] = t.l[index], t.l[0]
	t.Unlock()
}
