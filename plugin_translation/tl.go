// Package translation 翻译
package translation

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

var (
	prio   = 100
	bucket = rate.NewManager(time.Minute, 20) // 接口回复
)

func tl(d string) ([]byte, error) {
	url := "https://api.cloolc.club/fanyi?data=" + d
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if code := resp.StatusCode; code != 200 {
		// 如果返回不是200则立刻抛出错误
		errmsg := fmt.Sprintf("code %d", code)
		return nil, errors.New(errmsg)
	}
	return data, err
}

func init() {
	control.Register("translation", &control.Options{
		DisableOnDefault: false,
		Help: "翻译\n" +
			">TL 你好",
	}).OnRegex(`^>TL\s(-.{1,10}? )?(.*)$`).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			if !bucket.Load(ctx.Event.UserID).Acquire() {
				// 频繁触发，不回复
				return
			}
			msg := []string{ctx.State["regex_matched"].([]string)[2]}
			rely, err := tl(msg[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			info := gjson.ParseBytes(rely)
			repo := info.Get("data.0")
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text(repo.Get("value.0")))
		})
}
