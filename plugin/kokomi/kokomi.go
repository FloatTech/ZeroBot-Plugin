// Package kokomi 原神面板查询
package kokomi

import (
	"strconv"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	api     = "http://8.134.179.136/"
	genshin = "genshin/"
)

func init() {
	en := control.Register("kokomi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "原神面板查询",
		Help:             "- 绑定xxx\n",
	})
	en.OnRegex(`^(?:#|＃)?\s*绑定+?\s*(?:uid|UID|Uid)?\s*(\d+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		suid := ctx.State["regex_matched"].([]string)[1] // 获取uid
		body, err := web.GetData(api + genshin + "bound?qq=" + strconv.Itoa(int(ctx.Event.UserID)) + "&uid=" + suid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(body))
	})
	en.OnRegex(`^(?:#|＃)?(.*)面板\s*(?:(?:\[CQ:at,qq=)(\d+))?(\d+)?(.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var i string
		str := ctx.State["regex_matched"].([]string)[1] // 获取key
		if str == "" {
			str = ctx.State["regex_matched"].([]string)[4]
		}
		if ctx.State["regex_matched"].([]string)[3] == "" {
			if i = ctx.State["regex_matched"].([]string)[2]; i == "" {
				i = strconv.FormatInt(ctx.Event.UserID, 10)
			}
			body, err := web.GetData(api + genshin + "qtop?qq=" + i + "&role=" + str)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(body))
			return
		}
		i = ctx.State["regex_matched"].([]string)[3]
		body, err := web.GetData(api + genshin + "utop?uid=" + i + "&role=" + str)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.ImageBytes(body))
	})
	en.OnRegex(`^(?:#|＃)?\s*更新+?\s*(?:uid|UID|Uid)?\s*(\d+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		suid := ctx.State["regex_matched"].([]string)[1] // 获取uid
		body, err := web.GetData(api + genshin + "find?uid=" + suid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(body))
	})
}
