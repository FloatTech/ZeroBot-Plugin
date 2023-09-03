// Package hs 炉石
package hs

import (
	"os"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

var reqconf = [...]string{"GET", "https://hs.fbigame.com",
	"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"}

const (
	hs   = `https://hs.fbigame.com/ajax.php?`
	para = "mod=get_cards_list&" +
		"mode=-1&" +
		"extend=-1&" +
		"mutil_extend=&" +
		"hero=-1&" +
		"rarity=-1&" +
		"cost=-1&" +
		"mutil_cost=&" +
		"techlevel=-1&" +
		"type=-1&" +
		"collectible=-1&" +
		"isbacon=-1&" +
		"page=1&" +
		"search_type=1&" +
		"deckmode=normal"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "炉石搜卡",
		Help: "- 搜卡[xxxx]\n" +
			"- [卡组代码xxx]\n" +
			"- 更多搜卡指令参数：https://hs.fbigame.com/misc/searchhelp",
		PrivateDataFolder: "hs",
	}).ApplySingle(ctxext.DefaultSingle)
	cachedir := file.BOTPATH + "/" + engine.DataFolder()
	engine.OnRegex(`^搜卡(.+)$`).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		List := ctx.State["regex_matched"].([]string)[1]
		g := sh(List)
		t := int(gjson.Get(g, `list.#`).Int())
		if t == 0 {
			ctx.SendChain(message.Text("查询为空！"))
			return
		}
		var sk message.Message
		for i := 0; i < t && i < 5; i++ {
			cid := gjson.Get(g, `list.`+strconv.Itoa(i)+`.CardID`).String()
			cachefile := cachedir + cid
			if file.IsNotExist(cachefile) {
				data, err := web.RequestDataWith(web.NewDefaultClient(),
					`https://res.fbigame.com/hs/v13/`+cid+`.png?auth_key=`+
						gjson.Get(g, `list.`+strconv.Itoa(i)+`.auth_key`).String(),
					reqconf[0], reqconf[1], reqconf[2], nil)
				if err == nil {
					err = os.WriteFile(cachefile, data, 0644)
				}
				if err != nil {
					continue
				}
			}
			sk = append(sk, ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+cachefile)))
		}
		if id := ctx.SendGroupForwardMessage(
			ctx.Event.GroupID,
			sk,
		).Get("message_id").Int(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})
	// 卡组
	engine.OnRegex(`^[\s\S]*?(AAE[a-zA-Z0-9/\+=]{70,})[\s\S]*$`).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		List := ctx.State["regex_matched"].([]string)[1]
		ctx.SendChain(
			message.Image(kz(List)),
		)
	})
}

func sh(s string) string {
	data, err := web.RequestDataWith(web.NewDefaultClient(), "https://hs.fbigame.com", reqconf[0], reqconf[1], reqconf[2], nil)
	if err == nil {
		url := hs + para + "&hash=" + strings.SplitN(strings.SplitN(helper.BytesToString(data), `var hash = "`, 2)[1], `"`, 2)[0] + "&search=" + s
		r, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
		if err == nil {
			return helper.BytesToString(r)
		}
	}
	return ""
}

func kz(s string) string {
	data, err := web.RequestDataWith(web.NewDefaultClient(), "https://hs.fbigame.com", reqconf[0], reqconf[1], reqconf[2], nil)
	if err == nil {
		url := hs + para + "mod=general_deck_image&deck_code=" + s + "&deck_text=&hash=" + strings.SplitN(strings.SplitN(helper.BytesToString(data), `var hash = "`, 2)[1], `"`, 2)[0] + "&search=" + s
		r, err := web.RequestDataWith(web.NewDefaultClient(), url, reqconf[0], reqconf[1], reqconf[2], nil)
		if err == nil {
			return "base64://" + gjson.Get(helper.BytesToString(r), "img").String()
		}
	}
	return ""
}
