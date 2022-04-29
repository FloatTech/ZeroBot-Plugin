// Package lolicon 基于 https://api.lolicon.app 随机图片
package lolicon

import (
	"encoding/base64"
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
	api      = "https://api.lolicon.app/setu/v2"
	capacity = 10
)

var (
	queue   = make(chan string, capacity)
	custapi = ""
)

func init() {
	en := control.Register("lolicon", &control.Options{
		DisableOnDefault: false,
		Help: "lolicon\n" +
			"- 来份萝莉\n" +
			"- 设置随机图片地址[http...]",
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnFullMatch("来份萝莉").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			go func() {
				for i := 0; i < math.Min(cap(queue)-len(queue), 2); i++ {
					if custapi != "" {
						data, err := web.GetData(custapi)
						if err != nil {
							ctx.SendChain(message.Text("ERROR:", err))
							continue
						}
						queue <- "base64://" + base64.StdEncoding.EncodeToString(data)
						continue
					}
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
				id := ctx.SendChain(message.Image(img))
				if id.ID() == 0 {
					process.SleepAbout1sTo2s()
					id = ctx.SendChain(message.Image(img).Add("cache", "0"))
					if id.ID() == 0 {
						ctx.SendChain(message.Text("ERROR:图片发送失败，可能被风控了~"))
					}
				}
			}
		})
	en.OnPrefix("设置随机图片地址", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			u := strings.TrimSpace(ctx.State["args"].(string))
			if !strings.HasPrefix(u, "http") {
				ctx.SendChain(message.Text("ERROR:url非法!"))
				return
			}
			custapi = u
		})
}
