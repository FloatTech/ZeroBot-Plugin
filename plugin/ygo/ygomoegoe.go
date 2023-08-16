// Package ygo 一些关于ygo的插件
package ygo

import (
	"fmt"
	"net/url"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

var (
	speakers = map[string]uint{
		"游城十代": 0, "十代": 0,
		"丸藤亮": 1, "亮": 1, "凯撒": 1,
		"海马濑人": 2, "海马": 2, "社长": 2,
		"爱德菲尼克斯": 3, "爱德": 3,
		"不动游星": 4, "游星": 4,
		"鬼柳京介": 5, "鬼柳": 5,
		"榊遊矢": 6, "榊游矢": 6, "游矢": 6,
	}
)

func init() {
	speakerList := make([]string, 0, len(speakers))
	for speaker := range speakers {
		speakerList = append(speakerList, speaker)
	}
	en := control.Register("ygomoegoe", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏王 moegoe 模型拟声",
		Help: "- 让[xxxx]说(日语)\n" +
			"当前角色:\n" + strings.Join(speakerList, "\n"),
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnRegex("^让(" + strings.Join(speakerList, "|") + ")说([A-Za-z\\s\\d\u3005\u3040-\u30ff\u4e00-\u9fff\uff11-\uff19\uff21-\uff3a\uff41-\uff5a\uff66-\uff9d\\pP]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("正在尝试"))
			text := ctx.State["regex_matched"].([]string)[2]
			if len([]rune(text)) > 30 {
				ctx.SendChain(message.Text("仅支持30符号以内的文字"))
				return
			}
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			url := fmt.Sprintf("http://127.0.0.1:8000/ygo/%v?text=%v&type=wav&output=%v", id, url.QueryEscape(text), ctx.Event.UserID)
			ctx.SendChain(message.Record(url))
		})
	en.OnRegex("^让洛天依说(.+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("正在尝试"))
			text := ctx.State["regex_matched"].([]string)[1]
			url := fmt.Sprintf("http://127.0.0.1:8000/LuoTianyi/0?text=[ZH]%v[ZH]&type=wav&output=%v", url.QueryEscape(text), ctx.Event.UserID)
			ctx.SendChain(message.Record(url))
		})
}
