// Package translation 翻译
package translation

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/zbputils/control/order"
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
	control.Register("translation", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "翻译\n" +
			">TL 你好",
	}).OnRegex(`^>TL\s(-.{1,10}? )?(.*)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
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
