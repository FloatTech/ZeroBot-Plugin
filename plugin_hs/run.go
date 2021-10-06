package hs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var header = req.Header{
	"user-agent": `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36`,
	"referer":    `https://hs.fbigame.com`,
}

func init() {
	zero.OnRegex(`^搜卡(.+)$`).
		SetBlock(true).SetPriority(20).Handle(func(ctx *zero.Ctx) {
		List := ctx.State["regex_matched"].([]string)[1]
		g := sh(List)
		t := int(gjson.Get(g, `list.#`).Int())
		if t == 0 {
			ctx.SendChain(message.Text("查询为空！"))
			return
		}
		var sk message.Message
		for i := 0; i < t && i < 5; i++ {
			im, _ := req.Get(`https://res.fbigame.com/hs/v13/`+gjson.Get(g, `list.`+strconv.Itoa(i)+`.CardID`).String()+
				`.png?auth_key=`+gjson.Get(g, `list.`+strconv.Itoa(i)+`.auth_key`).String(),
				header,
			)
			sg, _ := req.Post("https://pic.sogou.com/pic/upload_pic.jsp", req.FileUpload{
				File:      im.Response().Body,
				FieldName: "image",      // FieldName 是表单字段名
				FileName:  "avatar.png", // Filename 是要上传的文件的名称
			})
			sk = append(
				sk,
				message.CustomNode(
					ctx.Event.Sender.NickName,
					ctx.Event.UserID,
					`[CQ:image,file=`+sg.String()+`]`, //图片
				),
			)
		}
		ctx.SendGroupForwardMessage(
			ctx.Event.GroupID,
			sk,
		)
	})
	//卡组
	zero.OnRegex(`^[\s\S]*?(AAE[a-zA-Z0-9/\+=]{70,})[\s\S]*$`).
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
