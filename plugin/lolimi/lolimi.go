// Package lolimi 来源于 https://api.lolimi.cn/ 的接口
package lolimi

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	lolimiURL = "https://api.lolimi.cn"
	jiaranURL = lolimiURL + "/API/yyhc/jr.php?msg=%v&type=2"
	tafeiURL  = lolimiURL + "/API/yyhc/taf.php?msg=%v&type=2"
	dxlURL    = lolimiURL + "/API/yyhc/dxl.php?msg=%v&type=2"
	raoURL    = lolimiURL + "/API/rao/api.php"
	yanURL    = lolimiURL + "/API/yan/?url=%v"
	xjjURL    = lolimiURL + "/API/tup/xjj.php"
	qingURL   = lolimiURL + "/API/qing/api.php"
	fabingURL = lolimiURL + "/API/fabing/fb.php?name=%v"
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "桑帛云 API",
		Help: "- 让[塔菲|嘉然|东雪莲]说我测尼玛\n- 随机绕口令\n- 颜值鉴定[图片]\n" +
			"- 随机妹子\n- 随机情话\n- 发病 嘉然\n\n说明: 颜值鉴定只能鉴定三次元图片",
	})
)

func init() {
	engine.OnFullMatch("随机妹子").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Image(xjjURL))
	})
	engine.OnFullMatch("随机绕口令").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(raoURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(gjson.Get(binary.BytesToString(data), "data.Msg").String()))
	})
	engine.OnFullMatch("随机情话").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(qingURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(binary.BytesToString(data)))
	})
	engine.OnPrefix(`发病`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		data, err := web.GetData(fmt.Sprintf(fabingURL, url.QueryEscape(name)))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(gjson.Get(binary.BytesToString(data), "data").String()))
	})
	engine.OnKeywordGroup([]string{"颜值鉴定"}, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			list := ctx.State["image_url"].([]string)
			if len(list) > 0 {
				ctx.SendChain(message.Text("少女祈祷中..."))
				data, err := web.GetData(fmt.Sprintf(yanURL, url.QueryEscape(list[0])))
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				assessment := gjson.Get(binary.BytesToString(data), "data.text").String()
				if assessment == "" {
					ctx.SendChain(message.Text("ERROR: 请输入正确的图片"))
					return
				}
				var text strings.Builder // 创建一个strings.Builder实例
				text.WriteString("评价: ")
				text.WriteString(assessment) // 添加评估信息

				for i := 0; i <= 2; i++ {
					key := gjson.Get(binary.BytesToString(data), "data.grade.key"+strconv.Itoa(i)).String()
					score := gjson.Get(binary.BytesToString(data), "data.grade.score"+strconv.Itoa(i)).String()
					if key != "" {
						text.WriteString("\n")
						text.WriteString(key)
						text.WriteString(": ")
						text.WriteString(score)
					}
				}

				ctx.SendChain(message.Text(text.String())) // 发送构建好的字符串
			}
		})
	engine.OnRegex("^让(塔菲|嘉然|东雪莲)说([\\s\u4e00-\u9fa5\u3040-\u309F\u30A0-\u30FF\\w\\p{P}\u3000-\u303F\uFF00-\uFFEF]+)$").Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		name := ctx.State["regex_matched"].([]string)[1]
		msg := ctx.State["regex_matched"].([]string)[2]
		recordURL := ""
		switch name {
		case "塔菲", "东雪莲":
			ttsURL := ""
			if name == "塔菲" {
				ttsURL = tafeiURL
			} else {
				ttsURL = dxlURL
			}
			data, err := web.GetData(fmt.Sprintf(ttsURL, url.QueryEscape(msg)))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			recordURL = gjson.Get(binary.BytesToString(data), "music").String()
		case "嘉然":
			recordURL = fmt.Sprintf(jiaranURL, url.QueryEscape(msg))
		default:
			recordURL = fmt.Sprintf(jiaranURL, url.QueryEscape(msg))
		}
		if recordURL == "" {
			ctx.SendChain(message.Text("ERROR: 语音生成失败"))
			return
		}
		ctx.SendChain(message.Record(recordURL))
	})
}
