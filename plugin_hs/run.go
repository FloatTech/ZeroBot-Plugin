// Package hs 炉石
package hs

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

var botpath, _ = os.Getwd()
var cachedir = botpath + "/data/hs/"

var header = req.Header{
	"user-agent": `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36`,
	"referer":    `https://hs.fbigame.com`,
}

func init() {
	os.RemoveAll(cachedir)
	err := os.MkdirAll(cachedir, 0755)
	if err != nil {
		panic(err)
	}

	engine := control.Register("hs", &control.Options{
		DisableOnDefault: false,
		Help: "炉石\n" +
			"- 搜卡[xxxx]\n" +
			"- [卡组代码xxx]\n" +
			"- 更多搜卡指令参数：https://hs.fbigame.com/misc/searchhelp",
	})
	engine.OnRegex(`^搜卡(.+)$`).
		SetBlock(true).SetPriority(20).Handle(func(ctx *zero.Ctx) {
		List := ctx.State["regex_matched"].([]string)[1]
		g := sh(List)
		t := int(gjson.Get(g, `list.#`).Int())
		if t == 0 {
			ctx.SendChain(message.Text("查询为空！"))
			return
		}
		var sk message.Message
		var imgcq string
		var data []byte
		for i := 0; i < t && i < 5; i++ {
			cid := gjson.Get(g, `list.`+strconv.Itoa(i)+`.CardID`).String()
			cachefile := cachedir + cid
			if _, err := os.Stat(cachefile); err != nil {
				im, err := req.Get(`https://res.fbigame.com/hs/v13/`+cid+
					`.png?auth_key=`+gjson.Get(g, `list.`+strconv.Itoa(i)+`.auth_key`).String(),
					header,
				)
				if err == nil {
					data, err = io.ReadAll(im.Response().Body)
					if err == nil {
						err = im.Response().Body.Close()
						if err == nil {
							err = os.WriteFile(cachefile, data, 0644)
						}
					}
				}
				if err != nil {
					imgcq = err.Error()
				} else {
					imgcq = `[CQ:image,file=` + "file:///" + cachefile + `]`
				}
			}
			sk = append(
				sk,
				message.CustomNode(
					ctx.Event.Sender.NickName,
					ctx.Event.UserID,
					imgcq, // 图片
				),
			)
		}
		ctx.SendGroupForwardMessage(
			ctx.Event.GroupID,
			sk,
		)
	})
	// 卡组
	engine.OnRegex(`^[\s\S]*?(AAE[a-zA-Z0-9/\+=]{70,})[\s\S]*$`).
		SetBlock(true).SetPriority(20).Handle(func(ctx *zero.Ctx) {
		fmt.Print("成功")
		List := ctx.State["regex_matched"].([]string)[1]
		ctx.SendChain(
			message.Image(kz(List)),
		)
	})
}

func sh(s string) string {
	var hs = `https://hs.fbigame.com/ajax.php`
	h, _ := req.Get("https://hs.fbigame.com", header)
	var param = req.Param{
		"mod":          `get_cards_list`,
		"mode":         `-1`,
		"extend":       `-1`,
		"mutil_extend": ``,
		"hero":         `-1`,
		"rarity":       `-1`,
		"cost":         `-1`,
		"mutil_cost":   ``,
		"techlevel":    `-1`,
		"type":         `-1`,
		"collectible":  `-1`,
		"isbacon":      `-1`,
		"page":         `1`,
		"search_type":  `1`,
		"deckmode":     "normal",
		"hash":         strings.SplitN(strings.SplitN(h.String(), `var hash = "`, 2)[1], `"`, 2)[0],
	}
	r, _ := req.Get(hs, header, param, req.Param{"search": s})
	return r.String()
}

func kz(s string) string {
	h, _ := req.Get("https://hs.fbigame.com")
	param := req.Param{
		"mod":       `general_deck_image`,
		"deck_code": s,
		"deck_text": ``,
		"hash":      strings.SplitN(strings.SplitN(h.String(), `var hash = "`, 2)[1], `"`, 2)[0],
	}
	r, _ := req.Get(`https://hs.fbigame.com/ajax.php`, param, h.Request().Header)
	im := gjson.Get(r.String(), "img").String()
	return `base64://` + im
}
