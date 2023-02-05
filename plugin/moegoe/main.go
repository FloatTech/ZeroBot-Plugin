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
	cnapi = "http://267978.proxy.nscc-gz.cn:8888?text=%s&speaker=%s"
)

var speakers = map[string]uint{
	"宁宁": 0, "爱瑠": 1, "芳乃": 2, "茉子": 3, "丛雨": 4, "小春": 5, "七海": 6,
	"Sua": 0, "Mimiru": 1, "Arin": 2, "Yeonhwa": 3, "Yuhwa": 4, "Seonbae": 5,
}

func init() {
	en := control.Register("moegoe", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "日韩中 VITS 模型拟声",
		Help: "- 让[宁宁|爱瑠|芳乃|茉子|丛雨|小春|七海]说(日语)\n" +
			"- 让[Sua|Mimiru|Arin|Yeonhwa|Yuhwa|Seonbae]说(韩语)\n" +
			"- 让[派蒙|空|荧|阿贝多|枫原万叶|温迪|八重神子|纳西妲|钟离|诺艾尔|凝光|托马|北斗|莫娜|荒泷一斗|提纳里|芭芭拉|艾尔海森|雷电将军|赛诺|琴|班尼特|五郎|神里绫华|迪希雅|夜兰|辛焱|安柏|宵宫|云堇|妮露|烟绯|鹿野院平藏|凯亚|达达利亚|迪卢克|可莉|早柚|香菱|重云|刻晴|久岐忍|珊瑚宫心海|迪奥娜|戴因斯雷布|魈|神里绫人|丽莎|优菈|凯瑟琳|雷泽|菲谢尔|九条裟罗|甘雨|行秋|胡桃|迪娜泽黛|柯莱|申鹤|砂糖|萍姥姥|奥兹|罗莎莉亚|式大将|哲平|坎蒂丝|托克|留云借风真君|昆钧|塞琉斯|多莉|大肉丸|莱依拉|散兵|拉赫曼|杜拉夫|阿守|玛乔丽|纳比尔|海芭夏|九条镰治|阿娜耶|阿晃|阿扎尔|七七|博士|白术|埃洛伊|大慈树王|女士|丽塔|失落迷迭|缭乱星棘|伊甸|伏特加女孩|狂热蓝调|莉莉娅|萝莎莉娅|八重樱|八重霞|卡莲|第六夜想曲|卡萝尔|姬子|极地战刃|布洛妮娅|次生银翼|理之律者|迷城骇兔|希儿|魇夜星渊|黑希儿|帕朵菲莉丝|天元骑英|幽兰黛尔|德丽莎|月下初拥|朔夜观星|暮光骑士|明日香|李素裳|格蕾修|梅比乌斯|渡鸦|人之律者|爱莉希雅|爱衣|天穹游侠|琪亚娜|空之律者|薪炎之律者|云墨丹心|符华|识之律者|维尔薇|芽衣|雷之律者|阿波尼亚]说(中文)",
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
	en.OnRegex("^让(派蒙|空|荧|阿贝多|枫原万叶|温迪|八重神子|纳西妲|钟离|诺艾尔|凝光|托马|北斗|莫娜|荒泷一斗|提纳里|芭芭拉|艾尔海森|雷电将军|赛诺|琴|班尼特|五郎|神里绫华|迪希雅|夜兰|辛焱|安柏|宵宫|云堇|妮露|烟绯|鹿野院平藏|凯亚|达达利亚|迪卢克|可莉|早柚|香菱|重云|刻晴|久岐忍|珊瑚宫心海|迪奥娜|戴因斯雷布|魈|神里绫人|丽莎|优菈|凯瑟琳|雷泽|菲谢尔|九条裟罗|甘雨|行秋|胡桃|迪娜泽黛|柯莱|申鹤|砂糖|萍姥姥|奥兹|罗莎莉亚|式大将|哲平|坎蒂丝|托克|留云借风真君|昆钧|塞琉斯|多莉|大肉丸|莱依拉|散兵|拉赫曼|杜拉夫|阿守|玛乔丽|纳比尔|海芭夏|九条镰治|阿娜耶|阿晃|阿扎尔|七七|博士|白术|埃洛伊|大慈树王|女士|丽塔|失落迷迭|缭乱星棘|伊甸|伏特加女孩|狂热蓝调|莉莉娅|萝莎莉娅|八重樱|八重霞|卡莲|第六夜想曲|卡萝尔|姬子|极地战刃|布洛妮娅|次生银翼|理之律者|迷城骇兔|希儿|魇夜星渊|黑希儿|帕朵菲莉丝|天元骑英|幽兰黛尔|德丽莎|月下初拥|朔夜观星|暮光骑士|明日香|李素裳|格蕾修|梅比乌斯|渡鸦|人之律者|爱莉希雅|爱衣|天穹游侠|琪亚娜|空之律者|薪炎之律者|云墨丹心|符华|识之律者|维尔薇|芽衣|雷之律者|阿波尼亚)说([\\s\u4e00-\u9fa5\\pP]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ctx.State["regex_matched"].([]string)[2]
			speaker := ctx.State["regex_matched"].([]string)[1]
			ctx.SendChain(message.Record(fmt.Sprintf(cnapi, url.QueryEscape(text), speaker)))
		})
}
