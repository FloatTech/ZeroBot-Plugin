// Package aireply AI 回复
package aireply

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/aireply"
	"github.com/FloatTech/AnimeAPI/chatgpt"
	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/pkumza/numcn"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var replyModes = [...]string{"青云客", "小爱", "ChatGPT"}

func init() { // 插件主体
	ent := control.Register("tts", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "人工智能语音回复",
		Help: "- @Bot 任意文本(任意一句话回复)\n" +
			"- 设置语音模式[原神人物]\n" +
			"- 设置默认语音模式[原神人物]\n" +
			"- 恢复成默认语音模式\n" +
			"- 为群 xxx 设置原神语音 api key xxxxxx (key请加开发群获得)\n" +
			"当前适用的原神人物含有以下：\n" + list(soundList[:], 5),
	})
	tts := newttsmode()
	enr := control.Register("aireply", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "人工智能回复",
		Help:              "- @Bot 任意文本(任意一句话回复)\n- 设置回复模式[青云客|小爱|ChatGPT]\n- 设置 ChatGPT SessionToken xxx",
		PrivateDataFolder: "aireply",
	})
	/*************************************************************
	*******************************AIreply************************
	*************************************************************/
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
	/*************************************************************
	***********************tts************************************
	*************************************************************/
	ent.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.ExtractPlainText()
			// 获取回复模式
			r := getReplyMode(ctx)
			// 获取回复的文本
			reply := r.TalkPlain(ctx.Event.UserID, msg, zero.BotConfig.NickName[0])
			// 获取语音
			index := tts.getSoundMode(ctx)
			record := message.Record(fmt.Sprintf(cnapi, index, url.QueryEscape(
				// 将数字转文字
				re.ReplaceAllStringFunc(reply, func(s string) string {
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						logrus.Errorln("[tts]", err)
						return s
					}
					return numcn.EncodeFromFloat64(f)
				}),
			), tts.getAPIKey(ctx)))
			// 发送语音
			if ID := ctx.SendChain(record); ID.ID() == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			}
		})
	ent.OnRegex(`^设置语音模式(.*)$`, zero.AdminPermission, func(ctx *zero.Ctx) bool {
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
		record := message.Record(fmt.Sprintf(cnapi, i, url.QueryEscape(testRecord[soundList[i]]), tts.getAPIKey(ctx))).Add("cache", 0)
		if ID := ctx.SendChain(record); ID.ID() == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置失败！无法发送测试语音，请重试。"))
			return
		}
		time.Sleep(time.Second * 2)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功"))
	})
	ent.OnRegex(`^设置默认语音模式(.*)$`, zero.SuperUserPermission, func(ctx *zero.Ctx) bool {
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
	ent.OnFullMatch("恢复成默认语音模式", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := tts.resetSoundMode(ctx)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err))
			return
		}
		// 设置验证
		index := tts.getSoundMode(ctx)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("设置成功，当前为", soundList[index]))
	})
	ent.OnRegex(`^为群\s*(-?\d+)\s*设置原神语音\s*api\s*key\s*([0-9a-zA-Z-_]{54}==)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		grp, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		err := tts.setAPIKey(ctx.State["manager"].(*ctrl.Control[*zero.Ctx]), grp, ctx.State["regex_matched"].([]string)[2])
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("设置成功"))
	})
	chatgptfile := ent.DataFolder() + "chatgpt.txt"
	cfg := &chatgpt.Config{
		UA:              chatgpt.UA,
		RefreshInterval: time.Hour,
		Timeout:         time.Minute,
	}
	data, err := os.ReadFile(chatgptfile)
	if err == nil {
		cfg.SessionToken = binary.BytesToString(data)
		chats = aireply.NewChatGPT(cfg)
	}
	go func() {
		for range time.NewTicker(time.Hour).C {
			if chats == nil {
				continue
			}
			err := os.WriteFile(chatgptfile, binary.StringToBytes(cfg.SessionToken), 0644)
			if err != nil {
				logrus.Warnln("[aireply] 保存 chatgpt session token 到", chatgptfile, "失败:", err)
			}
		}
	}()
	ent.OnRegex(`^设置\s*ChatGPT\s*SessionToken\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		token := ctx.State["regex_matched"].([]string)[1]
		f, err := os.Create(chatgptfile)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		defer f.Close()
		_, err = f.WriteString(token)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		chats = aireply.NewChatGPT(&chatgpt.Config{
			UA:              chatgpt.UA,
			SessionToken:    token,
			RefreshInterval: time.Hour,
			Timeout:         time.Minute,
		})
		ctx.SendChain(message.Text("设置成功"))
	})
	ent.OnFullMatch("重置ChatGPT连接").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		chats.Reset(ctx.Event.UserID)
	})
}
