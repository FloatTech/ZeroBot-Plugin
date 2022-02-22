package aireply

import (
	"errors"
	"github.com/pkumza/numcn"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"regexp"
	"strconv"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/tts"
	"github.com/FloatTech/AnimeAPI/tts/baidutts"
	"github.com/FloatTech/AnimeAPI/tts/mockingbird"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	ttsServiceName = "tts"
)

var (
	reNumber   = "(\\-|\\+)?\\d+(\\.\\d+)?"
	soundModes = [...]string{"拟声鸟阿梓", "拟声鸟药水哥", "百度女声", "百度男声", "百度度逍遥", "百度度丫丫"}
)

func init() {
	engine := control.Register(ttsServiceName, order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "语音回复(包括拟声鸟和百度)\n- @Bot 任意文本(任意一句话回复)\n- 设置语音模式拟声鸟阿梓 | 设置语音模式拟声鸟药水哥 | 设置语音模式百度女声 | 设置语音模式百度男声| 设置语音模式百度度逍遥 | 设置语音模式百度度丫丫",
	})
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			tts := NewTTS(getSoundMode(ctx))
			ctx.SendChain(message.Record(tts.Speak(ctx.Event.UserID, func() string {
				reply := r.TalkPlain(msg, zero.BotConfig.NickName[0])
				re := regexp.MustCompile(reNumber)
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
	engine.OnPrefix(`设置语音模式`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			err := setSoundMode(ctx, param)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
		})
}

func NewTTS(name string) tts.TTS {
	switch name {
	case "百度女声":
		return baidutts.NewBaiduTTS(0)
	case "百度男声":
		return baidutts.NewBaiduTTS(1)
	case "百度度逍遥":
		return baidutts.NewBaiduTTS(3)
	case "百度度丫丫":
		return baidutts.NewBaiduTTS(4)
	case "拟声鸟阿梓":
		return mockingbird.NewMockingBirdTTS(0)
	case "拟声鸟药水哥":
		return mockingbird.NewMockingBirdTTS(1)
	default:
		return mockingbird.NewMockingBirdTTS(0)
	}
}

func setSoundMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range soundModes {
		if s == name {
			ok = true
			index = int64(i)
			break
		}
	}
	if !ok {
		return errors.New("no such mode")
	}
	m, ok := control.Lookup(ttsServiceName)
	if !ok {
		return errors.New("no such plugin")
	}
	return m.SetData(gid, index)
}

func getSoundMode(ctx *zero.Ctx) (name string) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := control.Lookup(ttsServiceName)
	if ok {
		index := m.GetData(gid)
		if int(index) < len(soundModes) {
			return soundModes[index]
		}
	}
	return "拟声鸟阿梓"
}
