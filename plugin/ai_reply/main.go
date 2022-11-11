// Package aireply AI 回复
package aireply

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/aireply"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkumza/numcn"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var replyModes = [...]string{"青云客", "小爱"}

func init() { // 插件主体
	enOftts := control.Register("tts", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "人工智能语音回复",
		Help: "- @Bot 任意文本(任意一句话回复)\n" +
			"- 设置语音模式[原神人物]\n" +
			"- 设置默认语音模式[原神人物]\n" +
			"- 恢复成默认语音模式\n" +
			"当前适用的原神人物含有以下：\n" + list(soundList[:], 5),
	})
	tts := newttsmode()
	enOfreply := control.Register("aireply", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "人工智能回复",
		Help:             "- @Bot 任意文本(任意一句话回复)\n- 设置回复模式[青云客|小爱]",
	})
	/*************************************************************
	*******************************AIreply************************
	*************************************************************/
	enOfreply.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			aireply := aireply.NewAIReply(getReplyMode(ctx))
			reply := message.ParseMessageFromString(aireply.Talk(ctx.ExtractPlainText(), zero.BotConfig.NickName[0]))
			// 回复
			time.Sleep(time.Second * 1)
			if zero.OnlyPublic(ctx) {
				reply = append(reply, message.Reply(ctx.Event.MessageID))
				ctx.Send(reply)
				return
			}
			ctx.Send(reply)
		})
	enOfreply.OnPrefix("设置回复模式", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["args"].(string)
		err := setReplyMode(ctx, param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("成功"))
	})
	/*************************************************************
	***********************tts************************************
	*************************************************************/
	enOftts.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			// 获取回复模式
			r := aireply.NewAIReply(getReplyMode(ctx))
			// 获取回复的文本
			reply := r.TalkPlain(msg, zero.BotConfig.NickName[0])
			// 获取语音
			index := tts.getSoundMode(ctx)
			record := message.Record(fmt.Sprintf(cnapi, index, url.QueryEscape(
				// 将数字转文字
				re.ReplaceAllStringFunc(reply, func(s string) string {
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						log.Errorln("[tts]:", err)
						return s
					}
					return numcn.EncodeFromFloat64(f)
				}),
			))).Add("cache", 0)
			// 发送语音
			if ID := ctx.SendChain(record); ID.ID() == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			}
		})
	enOftts.OnRegex(`^设置语音模式(.*)$`, zero.AdminPermission, func(ctx *zero.Ctx) bool {
		param := ctx.State["regex_matched"].([]string)[1]
		if _, ok := testRecord[param]; !ok {
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["regex_matched"].([]string)[1]
		// 保存设置
		err := tts.setSoundMode(ctx, param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		i := tts.getSoundMode(ctx)
		if _, ok := testRecord[soundList[i]]; !ok {
			ctx.SendChain(message.Text("配置的语音人物数据丢失！请重新设置语音人物。"))
			return
		}
		record := message.Record(fmt.Sprintf(cnapi, i, url.QueryEscape(testRecord[soundList[i]]))).Add("cache", 0)
		if ID := ctx.SendChain(record); ID.ID() == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置失败！无法发送测试语音，请重试。"))
			return
		}
		time.Sleep(time.Second * 2)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功"))
	})
	enOftts.OnRegex(`^设置默认语音模式(.*)$`, zero.SuperUserPermission, func(ctx *zero.Ctx) bool {
		param := ctx.State["regex_matched"].([]string)[1]
		if _, ok := testRecord[param]; !ok {
			return false
		}
		return true
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		param := ctx.State["regex_matched"].([]string)[1]
		// 保存设置
		err := tts.setDefaultSoundMode(param)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功"))
	})
	enOftts.OnFullMatch("恢复成默认语音模式", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := tts.resetSoundMode(ctx)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		index := tts.getSoundMode(ctx)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，当前为", soundList[index]))
	})
}
