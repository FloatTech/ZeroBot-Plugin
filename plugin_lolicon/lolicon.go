/*
基于 https://api.lolicon.app 随机图片
*/
package plugin_lolicon

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	API = "https://api.lolicon.app/setu/v2"
	CAP = 10
)

var (
	QUEUE = make(chan string, CAP)
)

func init() {
	zero.OnFullMatch("来份萝莉").
		Handle(func(ctx *zero.Ctx) {
			go func() {
				min := func(a, b int) int {
					if a < b {
						return a
					} else {
						return b
					}
				}
				for i := 0; i < min(cap(QUEUE)-len(QUEUE), 2); i++ {
					resp, err := http.Get(API)
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
					QUEUE <- url
				}
			}()
			select {
			case <-time.After(time.Second * 10):
				ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
			case url := <-QUEUE:
				ctx.SendChain(message.Image(url))
			}
		})
}
