// Package aireply AI 回复
package aireply

import (
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/tts/genshin"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var t = newttsmode()

func init() { // 插件主体
	ent := control.Register("tts", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "人工智能语音回复",
		Help: "- @Bot 任意文本(任意一句话回复)\n" +
			"- 设置语音模式[原神人物/百度/拟声鸟] 数字(百度/拟声鸟模式)\n" +
			"- 设置默认语音模式[原神人物/百度/拟声鸟] 数字(百度/拟声鸟模式)\n" +
			"- 恢复成默认语音模式\n" +
			"- 设置原神语音 api key xxxxxx (key请加开发群获得)\n" +
			"当前适用的原神人物含有以下：\n" + list(genshin.SoundList[:], 5),
	})

	enr := control.Register("aireply", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "人工智能回复",
		Help:              "- @Bot 任意文本(任意一句话回复)\n- 设置回复模式[青云客|小爱]",
		PrivateDataFolder: "aireply",
	})

	enr.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			aireply := getReplyMode(ctx)
			reply := message.ParseMessageFromString(aireply.Talk(ctx.Event.UserID, ctx.ExtractPlainText(), zero.BotConfig.NickName[0]))
			// 回复
			time.Sleep(time.Second * 1)
			if zero.OnlyPublic(ctx) {
				reply = append(reply, message.Reply(ctx.Event.MessageID))
				ctx.Send(reply)
				return
			}
			ctx.Send(reply)
		})

	enr.OnPrefix("设置回复模式", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["args"].(string)
		err := setReplyMode(ctx, param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})

	ent.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			// 获取回复模式
			r := getReplyMode(ctx)
			// 获取回复的文本
			reply := r.TalkPlain(ctx.Event.UserID, msg, zero.BotConfig.NickName[0])
			// 获取语音
			speaker, err := t.getSoundMode(ctx)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			rec, err := speaker.Speak(ctx.Event.UserID, func() string { return reply })
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
				return
			}
			// 发送语音
			if id := ctx.SendChain(message.Record(rec)); id.ID() == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			}
		})

	ent.OnRegex(`^设置语音模式\s*([\S\D]*)\s*(\d*)$`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["regex_matched"].([]string)[1]
		num := ctx.State["regex_matched"].([]string)[2]
		n := 0
		var err error
		if num != "" {
			n, err = strconv.Atoi(num)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}
		// 保存设置
		logrus.Debugln("[tts] t.setSoundMode( ctx", param, n, n, ")")
		err = t.setSoundMode(ctx, param, n, n)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		if banner, ok := genshin.TestRecord[param]; ok {
			logrus.Debugln("[tts] banner:", banner, "get sound mode...")
			// 设置验证
			speaker, err := t.getSoundMode(ctx)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			logrus.Debugln("[tts] got sound mode, speaking...")
			rec, err := speaker.Speak(ctx.Event.UserID, func() string { return banner })
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("无法发送测试语音，请重试。"))
				return
			}
			logrus.Debugln("[tts] sending...")
			if id := ctx.SendChain(message.Record(rec).Add("cache", 0)); id.ID() == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("无法发送测试语音，请重试。"))
				return
			}
			time.Sleep(time.Second * 2)
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功"))
	})

	ent.OnRegex(`^设置默认语音模式\s*([\S\D]*)\s*(\d*)$`, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["regex_matched"].([]string)[1]
		num := ctx.State["regex_matched"].([]string)[2]
		n := 0
		var err error
		if num != "" {
			n, err = strconv.Atoi(num)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}
		// 保存设置
		err = t.setDefaultSoundMode(param, n, n)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功"))
	})

	ent.OnFullMatch("恢复成默认语音模式", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := t.resetSoundMode(ctx)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		speaker, err := t.getSoundMode(ctx)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，当前为", speaker))
	})

	ent.OnRegex(`^设置原神语音\s*api\s*key\s*([0-9a-zA-Z-_]{54}==)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := t.setAPIKey(ctx.State["manager"].(*ctrl.Control[*zero.Ctx]), ctx.State["regex_matched"].([]string)[1])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("设置成功"))
	})
}
