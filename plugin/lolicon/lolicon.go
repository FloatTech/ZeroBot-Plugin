// Package lolicon 基于 https://api.lolicon.app 随机图片
package lolicon

import (
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	api = "https://api.lolicon.app/setu/v2"
)

func init() {
	en := control.Register("lolicon", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "lolicon\n" +
			"- 随机图片\n" +
			"- 随机图片 萝莉|少女\n",
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnPrefix("随机图片").Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			rapi := api
			args := strings.TrimSpace(ctx.State["args"].(string))
			if args != "" {
				rapi += "?tag=" + url.QueryEscape(args)
			}
			data, err := web.GetData(rapi)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			json := gjson.ParseBytes(data)
			if e := json.Get("error").Str; e != "" {
				ctx.SendChain(message.Text("ERROR: ", e))
				return
			}
			if url := json.Get("data.0"); url.Str == "" {
				ctx.SendChain(message.Text("未找到相关内容, 换个tag试试吧"))
				return
			}
			url := json.Get("data.0.urls.original").Str
			url = strings.ReplaceAll(url, "i.pixiv.cat", "i.pixiv.re")
			if id := ctx.Send(message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image(url))}).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
		})
}
