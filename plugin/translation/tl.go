// Package translation 翻译
package translation

import (
	"github.com/FloatTech/zbputils/binary"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strconv"
	"strings"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"
)

func init() {
	control.Register("translation", &control.Options{
		DisableOnDefault: false,
		Help: "翻译\n" +
			">TL 你好",
	}).OnRegex(`^>TL\s(-.{1,10}? )?(.*)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			msg := []string{ctx.State["regex_matched"].([]string)[2]}
			data, err := web.GetData("https://api.cloolc.club/fanyi?data=" + msg[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
			}
			info := gjson.ParseBytes(data)
			repo := info.Get("data.0")
			process.SleepAbout1sTo2s()
			b := make([]byte, 0, 100)
			for _, v := range repo.Get("value").Array() {
				b = strconv.AppendQuote(b, v.String())
			}
			s := strings.ReplaceAll(binary.BytesToString(b), "\"\"", ",")
			s = strings.ReplaceAll(s, "\"", "")
			ctx.SendChain(message.Text(repo.Get("key").String(), ":", s))
		})
}
