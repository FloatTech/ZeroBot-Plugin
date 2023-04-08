// Package moegoe 日韩中 VITS 模型拟声
package moegoe

import (
	"fmt"
	"net/url"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/tts/genshin"
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
	"Sua": 0, "Mimiru": 1, "Arin": 2, "Yeonhwa": 3, "Yuhwa": 4, "Seonbae": 5,
	"派蒙": 0, "凯亚": 1, "安柏": 2, "丽莎": 3, "琴": 4, "香菱": 5, "枫原万叶": 6, "迪卢克": 7, "温迪": 8, "可莉": 9, "早柚": 10, "托马": 11, "芭芭拉": 12, "优菈": 13, "云堇": 14, "钟离": 15, "魈": 16, "凝光": 17, "雷电将军": 18, "北斗": 19, "甘雨": 20, "七七": 21, "刻晴": 22, "神里绫华": 23, "戴因斯雷布": 24, "雷泽": 25, "神里绫人": 26, "罗莎莉亚": 27, "阿贝多": 28, "八重神子": 29, "宵宫": 30, "荒泷一斗": 31, "九条裟罗": 32, "夜兰": 33, "珊瑚宫心海": 34, "五郎": 35, "散兵": 36, "女士": 37, "达达利亚": 38, "莫娜": 39, "班尼特": 40, "申鹤": 41, "行秋": 42, "烟绯": 43, "久岐忍": 44, "辛焱": 45, "砂糖": 46, "胡桃": 47, "重云": 48, "菲谢尔": 49, "诺艾尔": 50, "迪奥娜": 51, "鹿野院平藏": 52,
}

var 原 = newapikeystore("./data/tts/o.txt")

func init() {
	en := control.Register("moegoe", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "日韩中 VITS 模型拟声",
		Help: "- 让[宁宁|爱瑠|芳乃|茉子|丛雨|小春|七海]说(日语)\n" +
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
			if 原.k == "" {
				return
			}
			text := ctx.State["regex_matched"].([]string)[2]
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			ctx.SendChain(message.Record(fmt.Sprintf(genshin.CNAPI, id, url.QueryEscape(text), url.QueryEscape(原.k))))
		})
}
