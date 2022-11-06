// Package cloudmusic 网易云音乐热评
package cloudmusic

import (
	"encoding/json"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

type result struct {
	ID       string `json:"id"`
	Songid   string `json:"songid"`
	Name     string `json:"name"`
	Songname string `json:"songname"`
	Userid   string `json:"userid"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

const (
	api = "https://api.4gml.com/NeteaseMusic?type=json"
)

func init() { // 插件主体
	control.Register("cloudmusic", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "网易云热评",
		Help:             "- 来句网易云热评",
	}).OnFullMatch("来句网易云热评").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData(api)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			var r result
			err = json.Unmarshal(data, &r)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			ctx.SendChain(message.Text("歌曲名:", r.Name,
				"\n评论内容:", r.Content,
				"\n评论者:", r.Username))
		})
}
