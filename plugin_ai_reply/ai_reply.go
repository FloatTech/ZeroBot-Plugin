// Package aireply 人工智能回复
package aireply

import (
	"github.com/FloatTech/ZeroBot-Plugin/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"time"
)

var (
	prio   = 256
	bucket = rate.NewManager(time.Minute, 20) // 青云客接口回复
	engine = control.Register("aireply", &control.Options{
		DisableOnDefault: false,
		Help: "人工智能回复\n" +
			"- @Bot 任意文本(任意一句话回复)\n- 设置回复模式[青云客|小爱]\n- ",
	})
	// Mode 智能回复模式(1. 青云客，2. 小爱)
	Mode = 1
)

const (
	qykURL        = "http://api.qingyunke.com/api.php?key=free&appid=0&msg=%s"
	qykBotName    = "菲菲"
	xiaoaiURL     = "http://81.70.100.130/api/xiaoai.php?msg=%s&n=text"
	xiaoaiBotName = "小爱"
)

// AIReply 公用智能回复类
type AIReply interface {
	// DealQuestion 把椛椛替换为各api接口的bot名字
	DealQuestion(preMsg string) (msg string)
	// GetReply 取得回复消息
	GetReply(msg string) (reply string)
	// DealReply 处理回复消息
	DealReply(reply string) (textReply string, faceReply int)
}

// NewAIReply 智能回复简单工厂
func NewAIReply(mode int) AIReply {
	if mode == 1 {
		return &QYKReply{}
	} else if mode == 2 {
		return &XiaoAiReply{}
	}
	return nil
}

func init() { // 插件主体
	// 回复 @和包括名字
	engine.OnMessage(zero.OnlyToMe).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			AIReply := NewAIReply(Mode)
			if !bucket.Load(ctx.Event.UserID).Acquire() {
				// 频繁触发，不回复
				return
			}
			msg := ctx.ExtractPlainText()
			// 把消息里的椛椛替换成对应接口机器人的名字
			msg = AIReply.DealQuestion(msg)
			reply := AIReply.GetReply(msg)
			// 挑出 face 表情
			textReply, faceReply := AIReply.DealReply(reply)
			// 回复
			time.Sleep(time.Second * 1)
			if ctx.Event.MessageType == "group" {
				if faceReply != -1 {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(textReply), message.Face(faceReply))
				} else {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(textReply))
				}
			}
			if ctx.Event.MessageType == "private" {
				if faceReply != -1 {
					ctx.SendChain(message.Text(textReply), message.Face(faceReply))
				} else {
					ctx.SendChain(message.Text(textReply))
				}
			}
		})
	engine.OnPrefix(`设置回复模式`).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			switch param {
			case "青云客":
				Mode = 1
				ctx.SendChain(message.Text("设置为青云客回复"))
			case "小爱":
				Mode = 2
				ctx.SendChain(message.Text("设置为小爱回复"))
			default:
				ctx.SendChain(message.Text("设置失败"))
			}
		})
}

func getAgent() string {
	agent := [...]string{
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
		"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
		"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
		"User-Agent,Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"User-Agent, Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
		"User-Agent,Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	len1 := len(agent)
	return agent[r.Intn(len1)]
}
