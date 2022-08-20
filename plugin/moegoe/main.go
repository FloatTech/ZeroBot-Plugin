// Package moegoe 日韩中 VITS 模型拟声
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
	cnapi = "http://233366.proxy.nscc-gz.cn:8888?speaker=%s&text=%s"
)

var speakers = map[string]uint{
	"宁宁": 0, "爱瑠": 1, "芳乃": 2, "茉子": 3, "丛雨": 4, "小春": 5, "七海": 6,
	"Sua": 0, "Mimiru": 1, "Arin": 2, "Yeonhwa": 3, "Yuhwa": 4, "Seonbae": 5,
}

func init() {
	en := control.Register("moegoe", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "moegoe\n" +
			"- 让[宁宁|爱瑠|芳乃|茉子|丛雨|小春|七海]说(日语)\n" +
			"- 让[Sua|Mimiru|Arin|Yeonhwa|Yuhwa|Seonbae]说(韩语)\n" +
			"- 让[派蒙|凯亚|安柏|丽莎|琴|香菱|枫原万叶|迪卢克|温迪|可莉|早柚|托马|芭芭拉|优菈|云堇|钟离|魈|凝光|雷电将军|北斗|甘雨|七七|刻晴|神里绫华|雷泽|神里绫人|罗莎莉亚|阿贝多|八重神子|宵宫|荒泷一斗|九条裟罗|夜兰|珊瑚宫心海|五郎|达达利亚|莫娜|班尼特|申鹤|行秋|烟绯|久岐忍|辛焱|砂糖|胡桃|重云|菲谢尔|诺艾尔|迪奥娜|鹿野院平藏]说(中文)",
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnRegex("^让(宁宁|爱瑠|芳乃|茉子|丛雨|小春|七海)说([A-Za-z\\s\\d\u3005\u3040-\u30ff\u4e00-\u9fff\uff11-\uff19\uff21-\uff3a\uff41-\uff5a\uff66-\uff9d\\pP]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ctx.State["regex_matched"].([]string)[2]
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			ctx.SendChain(message.Record(fmt.Sprintf(jpapi, url.QueryEscape(text), id)))
		})
	en.OnRegex("^让(Sua|Mimiru|Arin|Yeonhwa|Yuhwa|Seonbae)说([A-Za-z\\s\\d\u3131-\u3163\uac00-\ud7ff\\pP]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ctx.State["regex_matched"].([]string)[2]
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			ctx.SendChain(message.Record(fmt.Sprintf(krapi, url.QueryEscape(text), id)))
		})
	en.OnRegex("^让(派蒙|凯亚|安柏|丽莎|琴|香菱|枫原万叶|迪卢克|温迪|可莉|早柚|托马|芭芭拉|优菈|云堇|钟离|魈|凝光|雷电将军|北斗|甘雨|七七|刻晴|神里绫华|雷泽|神里绫人|罗莎莉亚|阿贝多|八重神子|宵宫|荒泷一斗|九条裟罗|夜兰|珊瑚宫心海|五郎|达达利亚|莫娜|班尼特|申鹤|行秋|烟绯|久岐忍|辛焱|砂糖|胡桃|重云|菲谢尔|诺艾尔|迪奥娜|鹿野院平藏)说([\\s\u4e00-\u9fa5\\pP]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			speaker := ctx.State["regex_matched"].([]string)[1]
			text := ctx.State["regex_matched"].([]string)[2]
			ctx.SendChain(message.Record(fmt.Sprintf(cnapi, url.QueryEscape(speaker), url.QueryEscape(text))))
		})
}
