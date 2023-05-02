// Package kokomi 原神面板查询
package kokomi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	api = "http://8.134.179.136/genshin/"
)

func init() {
	en := control.Register("kokomi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "原神面板查询",
		Help: "- 绑定xxx\n" +
			"- xx面板",
	})
	en.OnRegex(`^(?:#|＃)?\s*绑定+?\s*(?:uid|UID|Uid)?\s*(\d+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		suid := ctx.State["regex_matched"].([]string)[1] // 获取uid
		body, err := web.GetData(api + "bound?qq=" + strconv.Itoa(int(ctx.Event.UserID)) + "&uid=" + suid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		msg, _, err := fixmessage(body)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(msg))
	})
	en.OnRegex(`^(?:#|＃)?(.*)面板\s*(?:(?:\[CQ:at,qq=)(\d+))?(\d+)?(.*)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fmt.Println(ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2], ctx.State["regex_matched"].([]string)[3], ctx.State["regex_matched"].([]string)[4])
		var i string
		str := ctx.State["regex_matched"].([]string)[1] // 获取key
		if str == "" {
			str = ctx.State["regex_matched"].([]string)[4]
		}
		if ctx.State["regex_matched"].([]string)[3] == "" {
			if i = ctx.State["regex_matched"].([]string)[2]; i == "" {
				i = strconv.FormatInt(ctx.Event.UserID, 10)
			}
			if str == "更新" {
				body, err := web.GetData(api + "find?qq=" + i)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				msg, _, err := fixmessage(body)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Text(msg))
			} else {
				body, err := web.GetData(api + "qtop?qq=" + i + "&role=" + str)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				_, url, err := fixmessage(body)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				ctx.SendChain(message.Image(url))
			}
			return
		}
		i = ctx.State["regex_matched"].([]string)[3]
		if str == "更新" {
			body, err := web.GetData(api + "find?uid=" + i)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			msg, _, err := fixmessage(body)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text(msg))
			return
		}
		body, err := web.GetData(api + "utop?uid=" + i + "&role=" + str)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		_, url, err := fixmessage(body)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image(url))
	})
	en.OnRegex(`^(?:#|＃)?\s*更新+?\s*(?:uid|UID|Uid)?\s*(\d+)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		suid := ctx.State["regex_matched"].([]string)[1] // 获取uid
		body, err := web.GetData(api + "find?uid=" + suid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		msg, _, err := fixmessage(body)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(msg))
	})

	en.OnRegex(`^(?:\[CQ:at,qq=)?(\d+)?\]?\s*(?:#|＃)?队伍伤害\s*((\D+)\s(\D+)\s(\D+)\s(\D+))?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		names := []string{ctx.State["regex_matched"].([]string)[3], ctx.State["regex_matched"].([]string)[4], ctx.State["regex_matched"].([]string)[5], ctx.State["regex_matched"].([]string)[6]} // 获取key
		sqquid := ctx.State["regex_matched"].([]string)[1]                                                                                                                                        // 获取第三者qquid
		if sqquid == "" {
			sqquid = strconv.FormatInt(ctx.Event.UserID, 10)
		}
		body, err := web.GetData(fmt.Sprintf("%sgroup?qq=%s&1=%s&2=%s&3=%s&4=%s", api, sqquid, names[0], names[1], names[2], names[3]))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		_, url, err := fixmessage(body)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image(url))
	})
}

type result struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	URL  string `json:"url"`
	UID  string `json:"uid"`
	Msg  string `json:"msg"`
}

func fixmessage(data []byte) (msg, url string, err error) {
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return "", "", errors.New(r.Msg)
	}
	if r.Code != 200 {
		return "", "", errors.New(r.Msg)
	}
	return r.Msg, r.URL, nil
}
