package aireply

import (
	"errors"
	"github.com/FloatTech/AnimeAPI/tts/baidutts"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/tts/mockingbird"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	mockingbirdServiceName = "mockingbird"
	baiduttsServiceName    = "baidutts"
)

var (
	baiduttsModes = map[int]string{1: "女声", 2: "男声", 3: "度逍遥", 4: "度丫丫"}
	perMap        = [...]int{1: 0, 2: 1, 3: 3, 4: 4}
)

func init() {
	baiduttsEngine := control.Register(baiduttsServiceName, order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "百度文字转语音\n- @Bot 任意文本(任意一句话回复)\n- 设置百度语音 度丫丫(4种模式,女声,男声,度逍遥,度丫丫)",
	})
	baiduttsEngine.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			per := perMap[getBaiduttsMode(ctx)]
			bt := baidutts.NewBaiduTTS(per)
			ctx.SendChain(message.Record(bt.Speak(ctx.Event.UserID, func() string {
				return r.TalkPlain(msg, zero.BotConfig.NickName[0])
			})))
		})
	baiduttsEngine.OnPrefix(`设置百度语音`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			err := setBaiduttsMode(ctx, param)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
		})
	control.Register(mockingbirdServiceName, order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "拟声鸟\n- @Bot 任意文本(任意一句话回复)",
	}).OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			r := aireply.NewAIReply(getReplyMode(ctx))
			ctx.SendChain(message.Record(mockingbird.NewMockingBirdTTS(1).Speak(ctx.Event.UserID, func() string {
				return r.TalkPlain(msg, zero.BotConfig.NickName[0])
			})))
		})
}

func setBaiduttsMode(ctx *zero.Ctx, name string) error {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	var ok bool
	var index int64
	for i, s := range baiduttsModes {
		if s == name {
			ok = true
			index = int64(i)
			break
		}
	}
	if !ok {
		return errors.New("no such mode")
	}
	m, ok := control.Lookup(baiduttsServiceName)
	if !ok {
		return errors.New("no such plugin")
	}
	return m.SetData(gid, index)
}

func getBaiduttsMode(ctx *zero.Ctx) (index int) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	m, ok := control.Lookup(baiduttsServiceName)
	if ok {
		index = int(m.GetData(gid))
		if _, ok := baiduttsModes[index]; ok {
			return index
		}
	}
	return 4
}
