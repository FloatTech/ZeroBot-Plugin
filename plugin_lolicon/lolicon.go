// Package lolicon 基于 https://api.lolicon.app 随机图片
package lolicon

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/data"
)

const (
	api      = "https://api.lolicon.app/setu/v2"
	capacity = 10
)

var (
	queue = make(chan string, capacity)
)

func init() {
	control.Register("lolicon", &control.Options{
		DisableOnDefault: false,
		Help: "lolicon\n" +
			"- 来份萝莉",
	}).OnFullMatch("来份萝莉").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			go func() {
				for i := 0; i < data.Min(cap(queue)-len(queue), 2); i++ {
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
					ctx.SendGroupMessage(0, message.Image(url))
					queue <- url
				}
			}()
			select {
			case <-time.After(time.Second * 10):
				ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
			case url := <-queue:
				ctx.SendChain(message.Image(url))
			}
		})
}
