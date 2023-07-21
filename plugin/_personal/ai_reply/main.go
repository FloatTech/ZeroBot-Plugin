// Package aireply AI 回复
package aireply

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const xiaoAIurl = "http://81.70.100.130/api/xiaoai.php?n=%v&msg=%v"

var (
	setmap = make(map[int64]int64, 255)
)

type respData struct {
	Code    int    `json:"code"`
	ID      uint   `json:"id"`
	Speaker string `json:"speaker"`
	URL     string `json:"url"`
	Message string `json:"message"`
}

func init() { // 插件主体
	en := control.Register("aireply", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "小爱智能回复",
		Help: "- @Bot 任意文本(任意一句话回复)\n" +
			"- [开启|关闭]语音模式",
	}).ApplySingle(ctxext.DefaultSingle)

	en.OnRegex(`^(开启|关闭)语音模式$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: 找不到 manager"))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		var setting int64 = 1
		if ctx.State["regex_matched"].([]string)[1] == "关闭" {
			setting = 0
		}
		if err := c.SetData(gid, int64(setting)); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		setmap[gid] = setting
		ctx.SendChain(message.Text("设置成功"))
	})

	en.OnMessage(zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		model, ok := setmap[gid]
		if !ok {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				ctx.SendChain(message.Text("ERROR: 找不到 manager"))
				return
			}
			model = c.GetData(gid)
			setmap[gid] = model
		}
		msg := ctx.ExtractPlainText()
		reply := getTalkString(msg, zero.BotConfig.NickName[1])
		if model == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			return
		}
		replyMsg, err := web.GetData("http://127.0.0.1:25525/tts?model=cantonese&id=21&text=" + url.QueryEscape("[ZH]"+reply+"[ZH]"+"&outputName="+strconv.FormatInt(ctx.Event.UserID, 10)))
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			return
		}
		var reslut respData
		err = json.Unmarshal(replyMsg, &reslut)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
			return
		}
		if reslut.Code == -1 {
			ctx.SendChain(message.Text("ERROR: ", reslut.Message))
			return
		}
		if id := ctx.SendChain(message.Record("file:///Users/liuyu.fang/Documents/Vits/MoeGoe/"+reslut.URL).Add("cache", 0)); id.ID() == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(reply))
		}
	})
}

func getTalkString(msg, nickname string) string {
	msg = strings.ReplaceAll(msg, nickname, "小爱")
	u := fmt.Sprintf(xiaoAIurl, "text", url.QueryEscape(msg))
	replyMsg, err := web.GetData(u)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	textReply := strings.ReplaceAll(binary.BytesToString(replyMsg), "小爱", nickname)
	if textReply == "" {
		textReply = nickname + "听不懂你的话了, 能再说一遍吗"
	}
	textReply = strings.ReplaceAll(textReply, "小米智能助理", "聊天伙伴")
	textReply = strings.ReplaceAll(textReply, " ", "")
	textReply = strings.ReplaceAll(textReply, "\n", "")
	return textReply
}
