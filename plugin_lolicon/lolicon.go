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

	"github.com/FloatTech/AnimeAPI/imgpool"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const (
	api      = "https://api.lolicon.app/setu/v2"
	capacity = 10
)

var (
	queue = make(chan string, capacity)
)

func init() {
	control.Register("lolicon", order.PrioLolicon, &control.Options{
		DisableOnDefault: false,
		Help: "lolicon\n" +
			"- 来份萝莉",
	}).OnFullMatch("来份萝莉").SetBlock(true).
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
					m, err := imgpool.GetImage(name)
					if err != nil {
						m.SetFile(url)
						err = m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
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
				ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
			case url := <-queue:
				ctx.SendChain(message.Image(url))
			}
		})
}
