// Package lolicon 基于 https://api.lolicon.app 随机图片
package lolicon

import (
	"strings"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/pool"
	"github.com/FloatTech/zbputils/math"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"
)

const (
	api      = "https://api.lolicon.app/setu/v2?tag=%E8%90%9D%E8%8E%89"
	capacity = 10
)

var (
	queue = make(chan string, capacity)
)

func init() {
	control.Register("lolicon", &control.Options{
		DisableOnDefault: true,
		Help: "lolicon\n" +
			"- 来份萝莉",
	}).ApplySingle(ctxext.DefaultSingle).OnFullMatch("来份萝莉").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			go func() {
				for i := 0; i < math.Min(cap(queue)-len(queue), 2); i++ {
					data, err := web.GetData(api)
					if err != nil {
						ctx.SendChain(message.Text("ERROR:", err))
						continue
					}
					json := gjson.ParseBytes(data)
					if e := json.Get("error").Str; e != "" {
						ctx.SendChain(message.Text("ERROR:", e))
						continue
					}
					url := json.Get("data.0.urls.original").Str
					url = strings.ReplaceAll(url, "i.pixiv.cat", "i.pixiv.re")
					name := url[strings.LastIndex(url, "/")+1 : len(url)-4]
					m, err := pool.GetImage(name)
					if err != nil {
						m.SetFile(url)
						_, err = m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
						process.SleepAbout1sTo2s()
					}
					if err == nil {
						queue <- m.String()
					} else {
						queue <- url
					}
				}
			}()
			select {
			case <-time.After(time.Minute):
				ctx.SendChain(message.Text("ERROR:等待填充，请稍后再试......"))
			case img := <-queue:
				msg := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image(img))}
				if id := ctx.SendGroupForwardMessage(
					ctx.Event.GroupID,
					msg,
				).Get("message_id").Int(); id == 0 {
					ctx.SendChain(message.Text("ERROR:可能被风控了"))
				}
			}
		})
}
