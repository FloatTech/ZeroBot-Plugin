// Package cloudmusic 网易云音乐热评
package cloudmusic

import (
	"io/ioutil"
	"net/http"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	api = "https://api.4gml.com/NeteaseMusic?type=bq"
)

func init() { // 插件主体
	control.Register("cloudmusic", &control.Options{
		DisableOnDefault: false,
		Help: "cloudmusic\n" +
			"- 来句网易云热评",
	}).OnFullMatch("来句网易云热评").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			res, err := http.Get(api)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if res.StatusCode != http.StatusOK {
				ctx.SendChain(message.Text("ERROR: code ", res.StatusCode))
				return
			}
			body, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			original := strings.ReplaceAll(string(body), "&nbsp;", "")
			ctx.SendChain(message.Text("歌曲名:", original[strings.Index(original, "「"):strings.Index(original, "」")+len("」")], "\n评论内容:", original[strings.Index(original, "『"):strings.Index(original, "』")+len("』")], "\n评论者:", original[strings.LastIndex(original, "「"):strings.LastIndex(original, "」")+len("」")]))
		})
}
