// Package moegoe 日韩 VITS 模型拟声
package moegoe

import (
	"fmt"
	"net/url"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	jpapi = "https://moegoe.azurewebsites.net/api/speak?text=%s&id=%d"
	krapi = "https://moegoe.azurewebsites.net/api/speakkr?text=%s&id=%d"
)

var speakers = map[string]uint{
	"宁宁": 0, "爱瑠": 1, "芳乃": 2, "茉子": 3, "丛雨": 4, "小春": 5, "七海": 6,
	"수아": 0, "미미르": 1, "아린": 2, "연화": 3, "유화": 4, "선배": 5,
}

func init() {
	en := control.Register("moegoe", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "moegoe\n" +
			"- 让[宁宁|爱瑠|芳乃|茉子|丛雨|小春|七海]说(日语)\n" +
			"- 让[수아|미미르|아린|연화|유화|선배]说(韩语)",
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnRegex("^让(宁宁|爱瑠|芳乃|茉子|丛雨|小春|七海)说([A-Za-z\\s\\d\u3005\u3040-\u30ff\u4e00-\u9fff\uff11-\uff19\uff21-\uff3a\uff41-\uff5a\uff66-\uff9d]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ctx.State["regex_matched"].([]string)[2]
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			ctx.SendChain(message.Record(fmt.Sprintf(jpapi, url.QueryEscape(text), id)))
		})
	en.OnRegex("^让(수아|미미르|아린|연화|유화|선배)说([A-Za-z\\s\\d\u3131-\u3163\uac00-\ud7ff]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ctx.State["regex_matched"].([]string)[2]
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			ctx.SendChain(message.Record(fmt.Sprintf(krapi, url.QueryEscape(text), id)))
		})
}
