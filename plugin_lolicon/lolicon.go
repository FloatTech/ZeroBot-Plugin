// Package lolicon 基于 https://api.lolicon.app 随机图片
package lolicon

import (
	"io/ioutil"
	"net/http"
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

	"github.com/FloatTech/zbputils/control/order"
)

const (
	api      = "https://api.lolicon.app/setu/v2"
	capacity = 10
)

var (
	queue = make(chan [2]string, capacity)
)

func init() {
	control.Register("lolicon", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "lolicon\n" +
			"- 来份萝莉",
	}).ApplySingle(ctxext.DefaultSingle).OnFullMatch("来份萝莉").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			go func() {
				for i := 0; i < math.Min(cap(queue)-len(queue), 2); i++ {
					resp, err := http.Get(api)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						continue
					}
					if resp.StatusCode != http.StatusOK {
						ctx.SendChain(message.Text("ERROR: code ", resp.StatusCode))
						continue
					}
					data, _ := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					json := gjson.ParseBytes(data)
					if e := json.Get("error").Str; e != "" {
						ctx.SendChain(message.Text("ERROR: ", e))
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
						queue <- [2]string{name, m.String()}
					} else {
						queue <- [2]string{name, url}
					}
				}
			}()
			select {
			case <-time.After(time.Minute):
				ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
			case o := <-queue:
				name := o[0]
				url := o[1]
				err := pool.SendRemoteImageFromPool(name, url, ctxext.Send(ctx), ctxext.GetMessage(ctx))
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
			}
		})
}
